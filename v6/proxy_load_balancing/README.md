# Centrifugo behind a load balancer / reverse proxy — a tested reference

This example is an **executable conformance harness**: it stands up Centrifugo
behind five popular L7 proxies and proves — in Docker, end to end — that every
bidirectional transport works correctly through each of them.

It exists because "does Centrifugo work behind my proxy?" is the most common
operational question, and the answer depends entirely on the proxy being
configured to (a) upgrade WebSocket, (b) **not buffer** the SSE / HTTP-streaming
response, (c) pass the `/emulation` POST to any node, and (d) keep idle
connections alive. Each proxy config here is minimal and correct, and the test
runner fails loudly if any of those four things is wrong.

## What it covers

| | |
|---|---|
| **Proxies** | nginx, HAProxy, Envoy, Traefik, Caddy |
| **Transports** | WebSocket, HTTP-streaming, SSE (all bidirectional, via [emulation](https://centrifugal.dev/docs/transports/overview)) |
| **Schemes** | plaintext (`ws`/`http`) **and** TLS (`wss`/`https`, self-signed) |
| **Topology** | 2 Centrifugo nodes + Redis, load-balanced with **no sticky sessions** |

## Why two nodes (this is the important part)

SSE and HTTP-streaming are one-way (server→client) streams. The client sends its
commands — including `subscribe` — as `POST /emulation` requests, which the proxy
load-balances **independently** of the long-lived stream. So the stream can live
on node A while a command lands on node B; Centrifugo routes the command to the
owning node internally (via a broker survey). A single-node test would pass even
through a broken proxy and give false confidence — so this harness always runs
two nodes, and includes a **deterministic cross-node test** that pins the stream
to node-1 and the `/emulation` POST to node-2.

## What each test actually asserts

For every `(proxy, transport, scheme)` combination:

1. **Connect** through the proxy.
2. **Subscribe** to a fresh channel.
3. **Publish** via the server API and require receipt within a bounded latency —
   this is what catches a **buffering** proxy. ("Connected OK" proves nothing: a
   buffering proxy still connects and only flushes at close.)
4. **Idle hold**, then publish again and require receipt — and fail on *any*
   disconnect/reconnect during the hold (even one the SDK silently recovers
   from). This catches a proxy **read/idle timeout** set below the ping interval.

Plus two dedicated modes:

- **`fanout`** — opens 9 connections (mixed across all three transports) through
  the balancer, confirms via `/api/info` that they were **spread across both
  nodes**, then publishes several messages and requires **every** connection to
  receive **every** message. This proves the broker fans a publication out across
  nodes and the balancer really is distributing connections (not pinning them all
  to one node).
- **`idle`** — holds all three transports open for 65s with only Centrifugo pings
  flowing (65s crosses the common 60s default idle timeout), asserting none drop.

## Run it

```bash
./run.sh                 # all five proxies + cross-node + idle probes
./run.sh nginx caddy     # a subset
```

Everything runs in `docker compose`; nothing is installed on the host. The script
exits non-zero if any combination fails.

Individual pieces:

```bash
docker compose run --rm certgen                       # generate CA + server cert
docker compose up -d --wait redis centrifugo-1 centrifugo-2
docker compose run --rm -e MODE=crossnode tester      # cross-node emulation
docker compose --profile nginx up -d --wait nginx
docker compose run --rm -e PROXY=nginx tester         # matrix through nginx
docker compose run --rm -e MODE=fanout -e PROXY=nginx tester # cross-node PUB/SUB
docker compose run --rm -e MODE=idle -e PROXY=nginx tester   # 65s idle probe
```

Tune with `-e IDLE_HOLD_MS=120000`, `-e FANOUT_CONNS=20`, `-e FANOUT_MSGS=10`.

### Load / connection-scaling test (10k+)

Opt-in (not part of `run.sh` — it's heavy). Ramps up many connections through a
proxy, holds them all, confirms they spread across both nodes, then broadcasts
and requires every connection to receive every message. Every proxy here is
verified holding **10,000** connections:

```bash
docker compose --profile nginx up -d --wait --no-deps --force-recreate nginx
docker compose run --rm -e MODE=load -e LOAD_CONNS=10000 -e PROXY=nginx tester
```

Tune with `-e LOAD_CONNS=20000`, `-e LOAD_MSGS=5`, `-e LOAD_TRANSPORT=http_stream`.
The connection-scaling knobs live in `docker-compose.yml` (`ulimits`) and the
proxy configs — see the "Scaling to many connections" section of the
[load balancing docs](https://centrifugal.dev/docs/server/load_balancing).

## The proxy configs

Each is deliberately small — the four things that matter are commented inline:

- [`proxies/nginx/nginx.conf`](proxies/nginx/nginx.conf)
- [`proxies/haproxy/haproxy.cfg`](proxies/haproxy/haproxy.cfg)
- [`proxies/envoy/envoy.yaml`](proxies/envoy/envoy.yaml)
- [`proxies/traefik/`](proxies/traefik) (`traefik.yml` + `dynamic.yml`)
- [`proxies/caddy/Caddyfile`](proxies/caddy/Caddyfile)

They map straight onto the guidance in the
[Load balancing docs](https://centrifugal.dev/docs/server/load_balancing).

> These configs are tuned for a demo (short ping interval, self-signed certs,
> `client.insecure`). For production, use real certificates and authentication —
> the proxy-facing settings (upgrade, buffering, timeouts, `/emulation`) are what
> carries over.
