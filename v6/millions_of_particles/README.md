# 2 million particles — multiplayer simulation over Centrifugo

A Go backend simulates 2,000,000 particles in a 2200 × 2200 world at 60 Hz.
Two transport modes selectable via the `MODE` env var:

- **`fanout`** (default) — one shared bitmap of the world is published to
  a single channel; every viewer receives the same bytes. Browser window
  clips the centered canvas. No panning. Default `K=3` downsample puts
  the per-frame payload at ~67 KB.
- **`shared_poll`** — the world is split into a 16 × 16 tile grid; each
  tile is one key on a shared-poll subscription. Each viewer tracks
  only the tiles its viewport intersects (with a 1-tile prefetch
  margin). Backend publishes all tiles each tick via Centrifugo's
  `/api/batch` API as a single `shared_poll_publish`-per-tile burst.
  Centrifugo fans out only the tiles each viewer tracks. Viewer pans
  the world with right-mouse drag, arrow keys, or WASD. Always
  full-resolution.

Each viewer can also influence the simulation regardless of mode: left
clicking and dragging sends an attractor position over Centrifugo RPC,
proxied to the backend.

The simulation logic is lifted with minor adaptations from
[dgerrells/how-fast-is-it](https://github.com/dgerrells/how-fast-is-it)
(see the [blog post](https://dgerrells.com/blog/how-fast-is-go-simulating-millions-of-particles-on-a-smart-tv)).

## Run

1. Start a local Centrifugo (v6.8+, needed for `shared_poll`) on port 8000:
   ```sh
   centrifugo -c centrifugo.json
   ```
2. Start the demo (default `fanout` mode):
   ```sh
   docker compose up --build
   ```
   Or `shared_poll` mode (full-resolution tiles, pan):
   ```sh
   MODE=shared_poll docker compose up --build
   ```
3. Open <http://localhost:9000>. Open more tabs to see the multiplayer effect.

`--build` is required after backend code changes.

## Prerequisites

- Centrifugo v6.8+ on port 8000 (binary, not the docker image — needs
  `shared_poll` support).
- Docker Compose for the backend + nginx.
- centrifuge-js dev build on port 2000 — the viewer loads
  `http://localhost:2000/centrifuge.protobuf.js` (the unpkg release
  doesn't yet include `newSharedPollSubscription`).

## Tuning

Backend env vars:

```
MODE=fanout|shared_poll       # default fanout
PARTICLE_COUNT=2000000
WORLD_WIDTH=2200
WORLD_HEIGHT=2200
DOWNSAMPLE=3                  # fanout-only; 1 = full resolution
HMAC_SECRET=...               # shared_poll signing key
```

## Bandwidth comparison (1410 × 730 viewport on a MacBook)

| Mode | Per frame per viewer | Per-viewer rate at 60 Hz | Notes |
|---|---|---|---|
| Original (per-client crop) | ~129 KB | ~7.7 MB/s | exactly the viewport, panable |
| `fanout`, K=1 (full-res) | 605 KB | 36 MB/s | shared bitmap, whole world |
| `fanout`, K=3 (default) | 67 KB | 4 MB/s | shared bitmap, downsampled |
| `shared_poll`, K=1 + 1-tile prefetch | ~260 KB | ~16 MB/s | full-res, only viewport tiles |

Backend → Centrifugo bandwidth in `shared_poll` mode is constant
~38 MB/s (256 tiles × ~2.5 KB × 60 Hz, regardless of viewer count) —
that's what enables fan-out at scale on the viewer side.

## How the bytes flow

### Fanout mode

```
            ┌──────────────────┐
            │  Go backend      │
            │  2M sim @ 60 Hz  │
            │  pack viewport   │
            └────────┬─────────┘
                     │  POST /api/publish (one bitmap)
                     ▼
                Centrifugo
                     │  fan-out (same bytes to all)
              ┌──────┼──────┐
              ▼      ▼      ▼
            viewer  viewer  viewer
```

### Shared-poll mode

```
            ┌──────────────────┐
            │  Go backend      │
            │  2M sim @ 60 Hz  │
            │  pack 256 tiles  │
            └────────┬─────────┘
                     │  POST /api/batch (256 shared_poll_publish)
                     ▼
                Centrifugo
                     │  per-key fan-out
                     │  (each viewer gets only their tracked tiles)
              ┌──────┼──────┐
              ▼      ▼      ▼
            viewer  viewer  viewer
              │      │      │
              │ track(visible tile keys) on pan
              │ getSignature → /api/track_refresh (HMAC)
              ▼      ▼      ▼
                 (RPC `input` for attractor — same in both modes)
```
