Centrifuge client's application example
=======================================

First, run Centrifuge with config:

```json
{
  "secret": "secret",
  "namespaces": [
    {
      "name": "chat",
      "anonymous": true,
      "publish": true,
      "watch": true,
      "presence": true,
      "join_leave": true,
      "history_size": 10,
      "history_lifetime": 30
    }
  ]
}
```

and then run this app with correct Centrifuge address and secret key:

```bash
python main.py --port=3000 --centrifuge=localhost:8000 --secret=secret
```

Then visit `http://localhost:3000` and select SockJS or pure websocket example.
