# PG Cluster Demo

Three-node Centrifugo cluster running on **PostgreSQL alone** — no Redis, no NATS.

All three pieces of the messaging plane share the same database:

- **Stream broker** (`broker.type = postgres`) — chat publications fan out across nodes.
- **Map broker** (`map_broker.type = postgres`) — cluster-wide presence rows.
- **Controller** (`controller.type = postgres`) — heartbeats, surveys, subscribe/disconnect propagation.

Companion to the blog post
[*Multi-node Centrifugo on PostgreSQL alone*](../../../centrifugal.dev/blog/2026-05-02-pg-controller-multi-node.md).

## Layout

- `centrifugo.json` — single shared config used by all 3 nodes.
- `start-nodes.sh` — convenience launcher that starts 3 local nodes on ports 8000/8001/8002,
  each with its own `CENTRIFUGO_NODE_NAME`. Logs are tee'd to `node-N.log`.
- `docker-compose.yml` — PostgreSQL + nginx (load balancer + static server).
- `nginx/default.conf` — exposes one `/connection/n{1,2,3}/...` per Centrifugo node so a
  browser tab can pick which node to talk to via `?n=1|2|3`. Also proxies `/api/info`
  with the API key injected.
- `static/index.html` — the demo UI: cluster topology, online tabs, live chat.

## Run

```bash
# 1. Bring up PostgreSQL and the static server.
docker compose up -d

# 2. Start 3 local Centrifugo nodes (in another terminal).
bash start-nodes.sh
# (or: chmod +x start-nodes.sh && ./start-nodes.sh)

# 3. Open the demo.
open http://localhost:9000/?n=1
open http://localhost:9000/?n=2
open http://localhost:9000/?n=3
```

> The future plan is to bake the 3 Centrifugo containers into `docker-compose.yml`
> directly once the release tag exists. For now `start-nodes.sh` runs them locally
> against the host PostgreSQL exposed on `localhost:5432`.

## Prerequisites

- Centrifugo built from a branch that includes the PG controller (`controller_postgres.go`).
- Docker Compose.
- `centrifuge-js` dev build serving on `http://localhost:2000/centrifuge.js`
  (matches the convention used by the other v6 demos).

## What each panel proves

| Panel | Component exercised | What you should see |
|---|---|---|
| **Cluster topology** | PG controller heartbeats | All 3 node names appear, each with its `num_clients` updating in real time. |
| **Online here** | PG map broker | One row per browser tab regardless of which node served it; rows disappear when a tab closes. |
| **Live chat** | PG stream broker | A message typed in one tab arrives in the other two within milliseconds via `LISTEN/NOTIFY`. |

## Wiping state between runs

Drop the demo database to reset all PG-managed Centrifugo schema:

```bash
docker compose down -v
```
