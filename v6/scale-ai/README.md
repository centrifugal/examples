# AI Token Streaming Playground

Interactive demo showing why Centrifugo is the right infrastructure for AI token streaming.

See blog post [Scaling AI token streams with Centrifugo](https://centrifugal.dev/blog/2026/03/01/scaling-ai-token-streams-with-centrifugo)

## Quick start

```bash
docker compose up --build
```

Open [http://localhost:9000](http://localhost:9000).

## What it demonstrates

- **Token streaming** — real-time delivery of AI-generated tokens via Centrifugo
- **Publisher-side aggregation** — toggle aggregation to batch multiple tokens per message, reducing message count while keeping token throughput
- **Recovery** — simulate a disconnect mid-stream and watch Centrifugo recover missed messages from history
- **Redis engine** — history and recovery backed by Redis, enabling horizontal scaling
- **Multi-tab sync** — open the same stream in a new tab, both tabs receive tokens in real-time
- **Transport fallbacks** — switch between WebSocket, SSE, and HTTP-streaming transports

## Architecture

```
                 ┌─────────────────────┐
                 │    nginx  :9000     │
                 └──┬───────────────┬──┘
                    │               │
                    ▼               ▼
┌──────────┐   ┌─────────┐   ┌────────────┐   ┌───────┐
│ postgres │◀──│ backend │──▶│ centrifugo │──▶│ redis │
└──────────┘   │  :5000  │   │   :8000    │   └───────┘
               └─────────┘   └────────────┘
```

Nginx serves the frontend and proxies `/api` to the backend, `/connection` and `/emulation` to Centrifugo.
