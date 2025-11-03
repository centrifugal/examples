# htmx-centrifugo Examples

This directory contains working examples demonstrating the htmx-centrifugo extension.

## Quick Start

### Prerequisites

- Docker and Docker Compose installed

### Running the Examples

From the `htmx-centrifugo` directory, run:

```bash
docker-compose up
```

This will start:
- **Backend server** (Python/FastAPI) on `http://localhost:4000` - Handles RPC calls
- **Centrifugo** server on `http://localhost:8000` - Real-time messaging
- **Web server** (nginx) on `http://localhost:3000` - Serves examples

Then open your browser to:

**http://localhost:3000/examples/**

You'll see an index page with links to all examples.

## Examples

### Chat Application
**URL:** http://localhost:3000/examples/chat.html

A full-featured chat application demonstrating:
- Connection with token authentication
- **Python/FastAPI backend** - Messages handled via RPC proxy
- Sending messages via RPC calls
- Real-time message updates across all connected clients
- Connection status indicators
- Event logging
- XSS protection and input validation (Pydantic)

**How it works:**
1. Browser sends RPC call to Centrifugo
2. Centrifugo proxies to FastAPI backend
3. FastAPI validates (Pydantic), formats, and publishes message
4. All connected clients receive the update instantly

Perfect for htmx users - uses familiar Python tools!

### Centrifugo Admin UI
**URL:** http://localhost:8000/dev

The built-in Centrifugo development page for testing connections directly.

## Testing Real-time Messages

### Using curl

Send a test message to the `updates` channel:

```bash
curl -X POST http://localhost:8000/api \
  -H "Authorization: apikey my-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "method": "publish",
    "params": {
      "channel": "updates",
      "data": {
        "html": "<div class=\"update\" style=\"padding: 10px; background: #e3f2fd; border-radius: 4px; margin: 5px 0;\">Hello from Centrifugo! ⚡</div>"
      }
    }
  }'
```

Send to the `news` channel:

```bash
curl -X POST http://localhost:8000/api \
  -H "Authorization: apikey my-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "method": "publish",
    "params": {
      "channel": "news",
      "data": {
        "html": "<p><strong>Breaking News:</strong> htmx-centrifugo is working perfectly!</p>"
      }
    }
  }'
```

Send to the `chat` channel:

```bash
curl -X POST http://localhost:8000/api \
  -H "Authorization: apikey my-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "method": "publish",
    "params": {
      "channel": "chat",
      "data": {
        "html": "<div class=\"message\"><div class=\"message-author\">API User</div><div class=\"message-text\">Hello from the API!</div><div class=\"message-time\">Just now</div></div>"
      }
    }
  }'
```

### Using the Browser Console

You can also publish messages from the browser console (useful for testing):

```javascript
// Simple message
fetch('http://localhost:8000/api', {
  method: 'POST',
  headers: {
    'Authorization': 'apikey my-api-key',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    method: 'publish',
    params: {
      channel: 'updates',
      data: {
        html: '<div class="update">Browser test message!</div>'
      }
    }
  })
});
```

## Configuration

### Centrifugo Configuration

The examples use `centrifugo-config.json`:

```json
{
  "client": {
    "allowed_origins": ["http://localhost:3000"],
    "insecure": true
  },
  "dev": {
    "enabled": true
  },
  "init": {
    "enabled": true
  },
  "channel": {
    "without_namespace": {
      "allow_subscribe_for_client": true,
      "allow_publish_for_client": true,
      "history_size": 100,
      "history_ttl": "300s",
      "force_recovery": true
    }
  }
}
```

**Key settings:**
- `insecure: true` - Allows anonymous connections (dev only!)
- `dev.enabled` - Enables `/dev` endpoints for easy token generation
- `init.enabled` - Enables `/connection/init` endpoint
- `rpc.proxy.endpoint` - RPC calls proxied to backend server
- `allow_subscribe_for_client` - Clients can subscribe to any channel
- `allow_publish_for_client: false` - Direct publishing disabled (use backend instead)

### Nginx Configuration

Nginx proxies requests to Centrifugo:

- `/connection/*` → Centrifugo WebSocket/SSE/HTTP-streaming endpoints
- `/dev/*` → Centrifugo dev endpoints (token generation)
- `/api` → Centrifugo HTTP API

This allows everything to work on the same origin (`localhost:3000`), avoiding CORS issues.

## Opening Multiple Browser Windows

To see real-time synchronization:

1. Open http://localhost:3000/examples/chat.html in multiple browser windows
2. Send a message in one window
3. Watch it appear instantly in all windows!

## Stopping the Examples

```bash
docker-compose down
```

## Troubleshooting

### Can't connect to Centrifugo

Check that services are running:
```bash
docker-compose ps
```

Check Centrifugo logs:
```bash
docker-compose logs centrifugo
```

### Messages not appearing

1. Check browser console for errors
2. Verify channel name matches between publish and subscribe
3. Check that the message has an `html` field
4. Enable debug mode in the examples: `centrifugo-debug="true"`

### CORS errors

The nginx proxy should handle CORS. If you see CORS errors:
- Make sure you're accessing via `http://localhost:3000` not opening files directly
- Check nginx logs: `docker-compose logs web`

## Next Steps

After trying the examples:

1. Read the [main README](../README.md) for full documentation
2. Check out the [QUICKSTART guide](../QUICKSTART.md)
3. Explore the [source code](../src/htmx-centrifugo.js)
4. Build your own real-time application!

## Example Files

- `index.html` - Landing page
- `chat.html` - Full chat application with authentication, RPC, and database persistence
- `README.md` - This file

## Production Notes

⚠️ **Important:** These examples use `insecure: true` for easy testing. In production:

- Remove `insecure: true`
- Implement proper JWT token generation
- Use HTTPS/WSS
- Configure proper `allowed_origins`
- Use authentication for the HTTP API
- Consider using a proper backend instead of direct client publishing

See the [main README](../README.md#authentication) for production setup guidance.
