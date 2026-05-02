#!/usr/bin/env bash
# Start three local Centrifugo nodes that share one PostgreSQL.
# Each node uses the same centrifugo.json — port and node name come from env.
#
# Logs are tee'd to node-N.log. Ctrl-C stops all three.

set -euo pipefail
cd "$(dirname "$0")"

CENTRIFUGO_BIN="${CENTRIFUGO_BIN:-centrifugo}"

cleanup() {
    echo
    echo "Stopping nodes..."
    kill 0
}
trap cleanup INT TERM

start_node() {
    local idx="$1" port="$2"
    CENTRIFUGO_NODE_NAME="node-${idx}" \
    CENTRIFUGO_HTTP_SERVER_PORT="${port}" \
        "${CENTRIFUGO_BIN}" -c centrifugo.json 2>&1 \
        | sed -u "s/^/[node-${idx}] /" \
        | tee "node-${idx}.log" &
}

start_node 1 8000
start_node 2 8001
start_node 3 8002

wait
