In this example we show how to use both Centrifugo GRPC unistream and Centrifugo GRPC API from a single Go program.

First, run Centrifugo with the config like:

```json
{
  "token_hmac_secret_key": "keep-it-secret",
  "api_key": "keep-it-secret",
  "uni_grpc": true,
  "grpc_api": true
}
```

Then run this example:

```
go run main.go
```

You should see sth like this:

```
❯ go run main.go
2022/10/13 18:04:54 establishing a unidirectional stream
2022/10/13 18:04:54 stream established
Publish OK
2022/10/13 18:04:54 connected to a server with ID: b3eb6778-4d18-4453-a425-b40ec301ac7a
Publish OK
2022/10/13 18:04:55 new publication from channel test_channel: "{\"input\":\"test_1665669895\"}"
Publish OK
2022/10/13 18:04:56 new publication from channel test_channel: "{\"input\":\"test_1665669896\"}"
```
