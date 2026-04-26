# 2 million particles — multiplayer simulation over Centrifugo

A Go backend simulates 2,000,000 particles in a 2200 × 2200 world at 60 Hz.
Every other tick it packs the entire world into a 1-bpp bitmap
(~605 KB) and publishes the bytes to the `particles:frame` channel via
Centrifugo's HTTP API. All connected browsers subscribe to the same
channel using the **Protobuf transport** so the bitmap arrives as raw
`Uint8Array`, then unpack and render it to a 2200 × 2200 canvas via
`ImageData`.

The canvas is shown at native resolution and centered inside an
`overflow: hidden` container — so smaller windows just show a smaller
*area* of the same world, not a downscaled copy. Two tabs side-by-side
at the same window size show the exact same crop, particles in
lockstep.

Each viewer can also influence the simulation: clicking and dragging
sends an attractor position over Centrifugo RPC, which is proxied to
the backend and applied to the next tick. All viewers see the same
world and influence the same particles.

The simulation logic is lifted with minor adaptations from
[dgerrells/how-fast-is-it](https://github.com/dgerrells/how-fast-is-it)
(see the [blog post](https://dgerrells.com/blog/how-fast-is-go-simulating-millions-of-particles-on-a-smart-tv)).
The differences relative to the original:

- One **shared** viewport, not per-client cameras — a single published
  frame fans out to every subscriber via Centrifugo's pub/sub.
- Inputs are keyed by Centrifugo client id and pruned on TTL.
- The frame bytes are published via Centrifugo's HTTP API instead of a
  raw WebSocket per client.

## Run

```sh
docker compose up --build
```

Then open <http://localhost:9000>. Open the same URL in two or more
tabs to see the multiplayer effect — each tab attracts particles
independently and all of them see the same shared simulation.

`--build` is required after backend code changes.

## Prerequisites

- Docker Compose
- The viewer loads `centrifuge.protobuf.js` from unpkg (`centrifuge@^5`)
  so binary publications arrive as `Uint8Array`. Centrifugo itself runs
  in the compose stack as `centrifugo/centrifugo:v6.7.1`.

## Tuning

The backend reads the following env vars (defaults shown match the
original demo):

```
PARTICLE_COUNT=2000000
WORLD_WIDTH=2200
WORLD_HEIGHT=2200
VIEWPORT_X=500
VIEWPORT_Y=500
VIEWPORT_W=1200
VIEWPORT_H=1200
```

Bandwidth: `W * H / 8` bytes per frame × 30 fps. The default
2200 × 2200 publishes ~605 KB/frame, ~18 MB/s per viewer. Smaller
viewports cut this proportionally — e.g. 1600 × 1600 ≈ 320 KB/frame,
~9.6 MB/s. The published frame is the same for everyone, so
publish-side bandwidth (backend → Centrifugo) stays constant
regardless of viewer count.

## How the bytes flow

```
            ┌──────────────────┐
            │  Go backend      │
            │  ┌────────────┐  │
            │  │ 2M particle │  │
            │  │ sim @ 60Hz │  │
            │  └────┬───────┘  │
            │       │ pack 1bpp
            │       ▼
            │  POST /api/publish
            └───────┬──────────┘
                    │   ~30 Hz, ~180 KB
                    ▼
              Centrifugo
                    │   protobuf-WS fan-out
        ┌───────────┼───────────┐
        ▼           ▼           ▼
     viewer 1   viewer 2   viewer N
        │           │           │
        └───────────┼───────────┘
                    │ RPC `input`  (centrifuge.rpc)
                    ▼
              Centrifugo  →  proxies to backend  →  sim.SetInput
```
