#!/usr/bin/env bash
# Orchestrated proxy conformance matrix. Everything runs in docker compose.
#
#   ./run.sh              # test all five proxies + the cross-node checks
#   ./run.sh nginx caddy  # test a subset
#
# Exit code is non-zero if any proxy/transport combination fails.
set -uo pipefail
cd "$(dirname "$0")"

ALL_PROXIES=("nginx" "haproxy" "envoy" "traefik" "caddy")
PROXIES=("${@:-}")
if [ -z "${PROXIES[*]}" ]; then PROXIES=("${ALL_PROXIES[@]}"); fi

dc() { docker compose "$@"; }

FAILS=()

cleanup() {
  echo "### tearing down"
  dc --profile nginx --profile haproxy --profile envoy --profile traefik --profile caddy --profile tester \
     down -v --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "### building tester image"
dc build tester

echo "### generating certs"
dc run --rm certgen

echo "### starting Redis + 2 Centrifugo nodes"
dc up -d --wait redis centrifugo-1 centrifugo-2

echo ""
echo "########## cross-node emulation (stream=node1, /emulation POST=node2) ##########"
if ! dc run --rm -e MODE=crossnode tester; then FAILS+=("crossnode"); fi

for p in "${PROXIES[@]}"; do
  echo ""
  echo "########## proxy: $p ##########"
  # --no-deps: certs are already generated and Centrifugo is already up, so start
  # only the proxy. --force-recreate so an edited proxy config is always picked up.
  if ! dc --profile "$p" up -d --wait --no-deps --force-recreate "$p"; then
    echo "  proxy $p failed to start"
    FAILS+=("$p:startup")
    dc --profile "$p" logs "$p" | tail -20 || true
    continue
  fi
  if ! dc run --rm -e PROXY="$p" tester; then FAILS+=("$p:matrix"); fi
  # Cross-node PUB/SUB: many connections spread across both nodes by the
  # balancer, every one must receive every published message.
  if ! dc run --rm -e MODE=fanout -e PROXY="$p" tester; then FAILS+=("$p:fanout"); fi
  # Idle-timeout probe: hold connections open with only pings for ~65s and
  # assert none drop (catches an untuned proxy read/idle timeout).
  if ! dc run --rm -e MODE=idle -e PROXY="$p" tester; then FAILS+=("$p:idle"); fi
  dc --profile "$p" stop "$p" >/dev/null 2>&1 || true
done

echo ""
echo "############################################################"
if [ ${#FAILS[@]} -eq 0 ]; then
  echo "ALL GREEN: every proxy passed every transport (http + https)."
  exit 0
else
  echo "FAILURES: ${FAILS[*]}"
  exit 1
fi
