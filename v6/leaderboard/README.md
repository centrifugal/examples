Source code for [Building a real-time WebSocket leaderboard with Centrifugo and Redis](https://centrifugal.dev/blog/2025/04/28/websocket-real-time-leaderboard)

To run:

```
docker compose up
```

Then go to http://localhost:8080

## Flush data in Redis

For example, if you want to start scores collection from scratch: 

```
docker compose exec redis redis-cli
```

And then:

```
flushdb
```