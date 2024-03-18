Showcase how to use Cenrifugo subscription streams to consume logs from Loki. See [blog post](https://centrifugal.dev/blog/2024/03/18/stream-loki-logs-to-browser-with-websocket-to-grpc-subscriptions) in Centrifugal blog.

To run:

```
docker compose up --build
```

Then visit `http://localhost:9000/` and put `{source="backend1"}` into input.
