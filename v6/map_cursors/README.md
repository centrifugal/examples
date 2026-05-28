# Map Cursors

Real-time multi-user cursors built on Centrifugo map subscriptions. Each client publishes its own cursor position into an `ephemeral` map channel keyed by client ID; every subscriber sees a live snapshot plus per-key updates.

This demo uses the **Redis map broker** — a better fit than PostgreSQL for high-frequency ephemeral updates like cursor movement.

## Run

```bash
docker compose up
```

Then open <http://localhost:9000> in two or more browser windows and move your mouse.

## Prerequisites

- Docker Compose
