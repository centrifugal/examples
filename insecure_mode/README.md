Demo of using insecure mode of Centrifugo.

Centrifugo must be started in insecure mode:

```
centrifugo --config=config.json --insecure
```

Make sure you have `publish: true` in Centrifugo configuration file to enable publishing messages to channel from client side.
