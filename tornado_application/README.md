Centrifuge client's application example
=======================================

First, run Centrifuge with config:

```json
{
  "token_hmac_secret_key": "secret",
  "namespaces": [
    {
      "name": "chat",
      "publish": true,
      "presence": true,
      "join_leave": true,
      "history_size": 10,
      "history_ttl": "30s"
    }
  ]
}
```

and then run this app with correct Centrifuge address and secret key:

```bash
python main.py --port=3000 --centrifuge=localhost:8000 --secret=secret
```

Then visit `http://localhost:3000` and select SockJS or pure websocket example.
