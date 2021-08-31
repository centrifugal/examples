Centrifugo v3 grpc-proxy example
=======================================

Follow next steps:
1) Generate grpc stubs
```bash
bash generate.sh
```
2) Start grpc-service
```
go run main.go
```
3) Run Centrifugo v3 (https://github.com/centrifugal/centrifugo)
with config like this:  
For more information see - https://centrifugal.github.io/centrifugo/server/proxy/

```json
{
  "connect_endpoint": "grpc://localhost:10001",
  "proxy_connect_endpoint": "grpc://localhost:10001",
  "proxy_refresh_endpoint": "grpc://localhost:10001",
  "proxy_rpc_endpoint": "grpc://localhost:10001",
  "proxy_publish_endpoint": "grpc://localhost:10001",
  "proxy_subscribe_endpoint": "grpc://localhost:10001",
  "proxy_connect_timeout":  "1s",
  "proxy_publish_timeout":  "1s",
  "proxy_subscribe_timeout":  "1s",
  "namespaces": [
    {
      "name": "chat",
      "publish": true,
      "proxy_publish": true,
      "proxy_subscribe": true,
      "anonymous": true,
      "history_size": 1000,
      "history_ttl": "1000s",
      "recover": true
    }
  ]
}
```

4) Run some centrifuge client  
For example, https://github.com/centrifugal/centrifuge-go/tree/master/examples/chat  
```
go run main.go
```
Also you may try other examples - https://github.com/centrifugal/centrifuge-go/tree/master/examples