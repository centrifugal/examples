# PG Cluster Demo

Three-node Centrifugo cluster running on **PostgreSQL alone** — no Redis, no NATS.

All three pieces of the messaging plane share the same database:

- **Stream broker** (`broker.type = postgres`) — chat publications fan out across nodes.
- **Map broker** (`map_broker.type = postgres`) — cluster-wide presence rows.
- **Controller** (`controller.type = postgres`) — heartbeats, surveys, subscribe/disconnect propagation.

Companion to the blog post
[*Multi-node Centrifugo on PostgreSQL alone*](../../../centrifugal.dev/blog/2026-05-16-pg-controller-multi-node.md).

## Layout

- `centrifugo.json` — single shared config used by all 3 nodes.
- `docker-compose.yml` — PostgreSQL + 3 Centrifugo nodes + nginx (load balancer + static server).
  Nodes are exposed on the host at ports 8000/8001/8002.
- `nginx/default.conf` — exposes one `/connection/n{1,2,3}/...` per Centrifugo node so a
  browser tab can pick which node to talk to via `?n=1|2|3`. Also proxies `/api/info`
  with the API key injected.
- `static/index.html` — the demo UI: cluster topology, online tabs, live chat.

## Run

```bash
docker compose up
```

Then open the demo:

```bash
open http://localhost:9000/?n=1
open http://localhost:9000/?n=2
open http://localhost:9000/?n=3
```

## Prerequisites

- Docker Compose.

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
