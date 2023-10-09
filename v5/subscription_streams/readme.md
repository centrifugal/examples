In this example we show how to use both Centrifugo subscription streams.

First, run Centrifugo with the config like:

```json
{
  "token_hmac_secret_key": "keep-it-secret",
  "proxy_subscribe_stream_endpoint": "grpc://localhost:12000"
  "namespaces": [
    {
      "name": "streams",
      "proxy_subscribe_stream": true,
      "proxy_subscribe_stream_bidirectional": false
    }
  ]
}
```

Then run this example:

```
go run main.go
```

Then upon subscriptions to channels in `streams` namespace this server will get requests for establishing unidirectional (or bidirectional) streams.

You should see sth like this in logs:

```
❯ go run main.go
unidirectional subscribe called with request client:"8e38b0ef-484a-4256-b342-b3d34044c30e"  transport:"websocket"  protocol:"json"  encoding:"json"  user:"2694"  channel:"streams:8e38b0ef-484a-4256-b342-b3d34044c30e"
```
