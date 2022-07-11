This example demonstrates integration of Centrifugo with Tornado server (Python).

First, run Centrifugo with config:

```json
{
  "token_hmac_secret_key": "secret",
  "allowed_origins": ["http://localhost:3000"],
  "namespaces": [
    {
      "name": "chat",
      "presence": true,
      "join_leave": true,
      "history_size": 10,
      "history_ttl": "30s",
      "allow_publish_for_subscriber": true,
    }
  ]
}
```

and then run this app with correct Centrifugo address and secret key:

```bash
python main.py --port=3000 --centrifuge=localhost:8000 --secret=secret
```

Then visit `http://localhost:3000`.
