# Map Presence Demo — 100k / 1M members in one browser tab

A single browser tab subscribes to a Centrifugo `map_clients` presence channel
and sees the entire population synchronized over the protocol. Two scales:

- **100k** — `clients_100k:massive` (320 × 320 grid, 2 px cell + 1 px gap)
- **1M**   — `clients_1m:massive`   (1024 × 1024 grid, 1 px per member)

The Go backend populates and churns each channel directly via the Centrifugo
HTTP API (`map_publish` / `map_remove`) — no real WebSocket clients are
simulated. Population stays exactly at the configured count: each "churn"
tick swaps one random live entry with one random non-live entry from a
stable id pool.

## Run

1. Start Centrifugo on port 8000: `centrifugo -c centrifugo.json`
2. Start the demo: `docker compose up --build` (rebuild required after backend code changes)
3. Open <http://localhost:9000>

## Prerequisites

- Centrifugo v6.8+ on port 8000
- centrifuge-js dev build on port 2000
- Docker Compose

## Configuration

Channel options live in `centrifugo.json`:

- **`clients_100k`** namespace — `map_clients`, recoverable, stream_size 50k, stream_ttl 5m.
- **`clients_1m`**   namespace — `map_clients`, recoverable, stream_size 200k, stream_ttl 10m.
- **`subscribe_catch_up_timeout: 30m`** — paginating 1M entries through state +
  stream phases easily exceeds the 5s default; this would otherwise trigger
  `DisconnectSlow` mid-load.
- **`max_page_size: 10000`** — viewer can request up to 10k entries per page
  to keep the round-trip count reasonable for 1M.

The backend reads `CENTRIFUGO_API_URL` and `CENTRIFUGO_API_KEY` from the
environment; defaults match the Centrifugo config.
