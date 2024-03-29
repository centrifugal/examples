A concept how to consume async request results from the backend using Centrifugo (using Centrifugo subscriptions and history with recovery). We can initiate long work, return unique channel to consume result from, subscribe on it with recovery to avoid missing the result, then unsubscribe from it upon receiving the result of operation to free resources.

> This may be a bit esoteric approach, you can also solve this by using one unique channel per browser tab (just subscribe on it before issuing any RPCs), and push results to a client over that channel. If you need to notify a user (not a particular client) – then you better publish async results into a user channel – so all browser tabs opened by that user will receive an async result.

I.e. step by step:

1. Client sends RPC to the server (may be a simple AJAX request to the backend)
2. Backend generates unique channel and returns it to a client together with subscription token valid for some time
3. Client subscribes to a channel and waits for incoming message. It does this with recovery - so that message in channel is guaranteed to be received as soon as it was published to Centrifugo channel (thus avoiding races between publish and subscribe). Note, that we are using empty epoch value `""` here – so that Centrifugo does not check stream epochs upon recovery. This is fine since we are using unique channels here which can be only created once.
4. Client unsuscribes from a channel upon getting result to free resources, history entries will soon expire

In Centrifugo v4 `history_meta_ttl` sets a global history meta information expiration time - i.e. for all namespaces. Probably in the future releases we will let configure `history_meta_ttl` per channel namespace.

Run example
===========

Run Centrifugo with `config.json` provided here:

```
./centrifugo -c config.json
```

Run Go example:

```
go run main.go
```

Open http://localhost:3000
