# htmx-centrifugo

An [htmx](https://htmx.org) extension for [Centrifugo](https://centrifugal.dev) real-time messaging server. This extension enables htmx applications to receive real-time updates through Centrifugo's WebSocket, SSE, or HTTP-streaming transports.

## Features

- 🔌 **Multiple transports**: WebSocket, Server-Sent Events (SSE), HTTP-streaming
- 🔄 **Automatic reconnection**: Built-in reconnection logic with exponential backoff
- 📡 **Declarative subscriptions**: Use HTML attributes to subscribe to channels
- 🔐 **JWT authentication**: Support for connection and subscription tokens
- 🎯 **Flexible swapping**: Multiple swap strategies like htmx (innerHTML, beforeend, etc.)
- 🚀 **Production-ready**: Built on Centrifuge JS SDK with battle-tested reliability
- 📦 **Lightweight**: Small footprint with no extra dependencies beyond centrifuge
- 🎨 **htmx-idiomatic**: Follows htmx patterns and conventions

## Installation

### Via npm

```bash
npm install htmx-centrifugo
```

### Via CDN

```html
<script src="https://unpkg.com/centrifuge@5/dist/centrifuge.js"></script>
<script src="https://unpkg.com/htmx.org@2"></script>
<script src="https://unpkg.com/htmx-centrifugo@0.1.0/dist/htmx-centrifugo.js"></script>
```

## Quick Start

### Try the Live Example

```bash
docker-compose up
```

Then open http://localhost:3000/examples/ for examples list, or http://localhost:3000/chat for the full-featured chat application.

### Basic Code Example

```html
<!DOCTYPE html>
<html>
<head>
  <script src="https://unpkg.com/centrifuge@5/dist/centrifuge.js"></script>
  <script src="https://unpkg.com/htmx.org@2"></script>
  <script src="https://unpkg.com/htmx-centrifugo/dist/htmx-centrifugo.js"></script>
</head>
<body>
  <!-- Connect to Centrifugo and subscribe to a channel -->
  <div hx-ext="centrifugo"
       centrifugo-connect
       centrifugo-ws-endpoint="/connection/websocket"
       centrifugo-token="your-jwt-token">

    <!-- This div will be updated with messages from 'news' channel -->
    <div id="news-feed"
         centrifugo-subscribe="news"
         centrifugo-swap="beforeend">
      <!-- Real-time updates appear here -->
    </div>
  </div>
</body>
</html>
```

## Usage

### Basic Connection

Enable the extension with `hx-ext="centrifugo"` and use `centrifugo-connect` to establish connection. Specify at least one transport endpoint:

```html
<div hx-ext="centrifugo"
     centrifugo-connect
     centrifugo-ws-endpoint="/connection/websocket"
     centrifugo-token="your-jwt-token">
  <!-- Your real-time content here -->
</div>
```

### Custom Transport Configuration

You can specify which transports to use:

```html
<!-- Only use WebSocket -->
<div hx-ext="centrifugo"
     centrifugo-ws-endpoint="/connection/websocket">
  <!-- Your real-time content here -->
</div>

<!-- Use all three transports with fallback -->
<div hx-ext="centrifugo"
     centrifugo-ws-endpoint="/connection/websocket"
     centrifugo-http-stream-endpoint="/connection/http_stream"
     centrifugo-sse-endpoint="/connection/sse">
  <!-- Your real-time content here -->
</div>
```

**Available transport attributes:**
- `centrifugo-ws-endpoint` - WebSocket endpoint path
- `centrifugo-http-stream-endpoint` - HTTP-streaming endpoint path
- `centrifugo-sse-endpoint` - Server-Sent Events endpoint path

Endpoints can be relative paths (like `/connection/websocket`) or full URLs (like `ws://localhost:8000/connection/websocket`).

### Authentication

#### Static Token

```html
<div hx-ext="centrifugo"
     centrifugo-token="your-jwt-token-here">
  <!-- Your content -->
</div>
```

#### Dynamic Token (Recommended)

Fetch token from your backend:

```html
<div hx-ext="centrifugo"
     centrifugo-token-url="/api/centrifugo/token">
  <!-- Your content -->
</div>
```

Your backend should return JSON:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Subscribing to Channels

Use `centrifugo-subscribe` to subscribe to a channel:

```html
<div centrifugo-subscribe="news">
  <!-- Updates from 'news' channel will replace this div's content -->
</div>
```

#### Swap Strategies

Control how updates are inserted using `centrifugo-swap`:

```html
<!-- Replace entire content (default) -->
<div centrifugo-subscribe="messages"
     centrifugo-swap="innerHTML">
</div>

<!-- Append to end -->
<div centrifugo-subscribe="chat"
     centrifugo-swap="beforeend">
</div>

<!-- Prepend to beginning -->
<div centrifugo-subscribe="notifications"
     centrifugo-swap="afterbegin">
</div>
```

Available swap strategies:
- `innerHTML` - Replace entire content (default)
- `outerHTML` - Replace the element itself
- `beforebegin` - Insert before the element
- `afterbegin` - Insert as first child
- `beforeend` - Insert as last child
- `afterend` - Insert after the element
- `delete` - Remove the element
- `none` - Don't modify DOM (just fire events)

#### Target Element

Update a different element using `centrifugo-target`:

```html
<div centrifugo-subscribe="news"
     centrifugo-target="#news-container"
     centrifugo-swap="beforeend">
</div>

<div id="news-container">
  <!-- Updates appear here -->
</div>
```

#### Subscription Tokens

For private channels requiring subscription tokens:

```html
<div centrifugo-subscribe="private-chat"
     centrifugo-sub-token-url="/api/centrifugo/subscription-token">
</div>
```

Your backend should return:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Sending Messages

#### Publishing to Channels

Use `centrifugo-send` with `centrifugo-channel`:

```html
<form centrifugo-send
      centrifugo-channel="chat"
      centrifugo-method="publish">
  <input name="message" placeholder="Type a message...">
  <button type="submit">Send</button>
</form>
```

**Note**: Publishing from client requires proper server-side permissions to be configured.

#### RPC Calls

Make RPC calls to your backend through Centrifugo:

```html
<form centrifugo-send
      centrifugo-method="rpc"
      centrifugo-rpc-method="getUserInfo">
  <input name="userId" placeholder="User ID">
  <button type="submit">Get Info</button>
</form>
```

### Message Format

The extension expects messages with an `html` field:

```javascript
// Server-side (Go example)
centrifuge.Publish("news", map[string]interface{}{
    "html": "<div class='news-item'>Breaking news!</div>",
})
```

```javascript
// Server-side (Node.js example)
await centrifugo.publish('news', {
    html: '<div class="news-item">Breaking news!</div>'
});
```

If the message data is a string, it will be used directly as HTML.

## Events

The extension fires custom events following htmx naming conventions (`htmx:extensionName-eventName`):

```javascript
document.addEventListener('htmx:centrifugo-connected', (evt) => {
  console.log('Connected:', evt.detail);
});

document.addEventListener('htmx:centrifugo-after-message', (evt) => {
  console.log('Received message:', evt.detail);
});

document.addEventListener('htmx:centrifugo-subscribed', (evt) => {
  console.log('Subscribed to:', evt.detail.channel);
});
```

### Available Events

**Connection Events:**
- `htmx:centrifugo-connecting` - Connection is being established
- `htmx:centrifugo-connected` - Successfully connected
- `htmx:centrifugo-disconnected` - Connection lost
- `htmx:centrifugo-error` - Connection error

**Subscription Events:**
- `htmx:centrifugo-subscribing` - Subscribing to a channel
- `htmx:centrifugo-subscribed` - Successfully subscribed
- `htmx:centrifugo-unsubscribed` - Unsubscribed from channel
- `htmx:centrifugo-subscription-error` - Subscription error

**Message Events:**
- `htmx:centrifugo-before-message` - Before message is processed (cancelable)
- `htmx:centrifugo-after-message` - After message is displayed

**Send Events:**
- `htmx:centrifugo-before-send` - Before sending message (cancelable)
- `htmx:centrifugo-after-send` - Message sent successfully
- `htmx:centrifugo-send-error` - Error sending message
- `htmx:centrifugo-rpc-result` - RPC call result received

See [EVENTS.md](EVENTS.md) for complete documentation with examples.

## Examples

### Live Chat Application

```html
<div hx-ext="centrifugo"
     centrifugo-connect
     centrifugo-ws-endpoint="/connection/websocket"
     centrifugo-token="{{ token }}">

  <!-- Chat messages -->
  <div id="messages"
       centrifugo-subscribe="chat"
       centrifugo-swap="beforeend"
       hx-on:htmx:after-settle="this.scrollTo(0, this.scrollHeight);"
       style="height: 400px; overflow-y: auto;">
  </div>

  <!-- Send form -->
  <form centrifugo-send
        centrifugo-method="rpc"
        centrifugo-rpc-method="sendChatMessage">
    <input name="message" placeholder="Type a message..." required>
    <button type="submit">Send</button>
  </form>
</div>
```

Check the `/examples` directory for a complete working chat application with authentication.

### Live Notifications

```html
<div hx-ext="centrifugo"
     centrifugo-connect
     centrifugo-ws-endpoint="/connection/websocket"
     centrifugo-token="{{ token }}">

  <div id="notifications"
       centrifugo-subscribe="notifications:user123"
       centrifugo-swap="afterbegin">
    <!-- New notifications appear at top -->
  </div>
</div>
```

### Live Dashboard with Multiple Feeds

```html
<div hx-ext="centrifugo"
     centrifugo-connect
     centrifugo-ws-endpoint="/connection/websocket"
     centrifugo-token="{{ token }}">

  <!-- Stats widget -->
  <div centrifugo-subscribe="stats:global"
       centrifugo-swap="innerHTML">
    Loading stats...
  </div>

  <!-- Activity feed -->
  <div centrifugo-subscribe="activity:recent"
       centrifugo-swap="beforeend">
    Loading activity...
  </div>

  <!-- Alerts -->
  <div centrifugo-subscribe="alerts:critical"
       centrifugo-target="#alert-banner"
       centrifugo-swap="innerHTML">
  </div>
</div>

<div id="alert-banner"></div>
```

**Note**: This demonstrates one of the key advantages over htmx-ext-ws - multiple channel subscriptions over a single WebSocket connection!

## Advanced Configuration

### Debug Mode

Enable debug logging:

```html
<div hx-ext="centrifugo"
     centrifugo-debug="true">
</div>
```

### HTTP/2 Extended CONNECT Workaround

Enable the init endpoint workaround for reliable HTTP/2 Extended CONNECT in Chrome:

```html
<div hx-ext="centrifugo"
     centrifugo-init="true">
</div>
```

This calls `/connection/init` before establishing the connection, ensuring Chrome uses HTTP/2 for WebSocket connections.

### Multiple Connections

You can have multiple independent Centrifugo connections on the same page:

```html
<!-- Connection for user notifications -->
<div hx-ext="centrifugo"
     centrifugo-connect
     centrifugo-ws-endpoint="/connection/websocket"
     centrifugo-token="{{ userToken }}">
  <div centrifugo-subscribe="user:notifications"></div>
</div>

<!-- Separate connection for public data (no token required) -->
<div hx-ext="centrifugo"
     centrifugo-connect
     centrifugo-ws-endpoint="/connection/websocket">
  <div centrifugo-subscribe="public:news"></div>
</div>
```

## Server-Side Setup

### Centrifugo Configuration

Basic `config.json`:

```json
{
  "client": {
    "allowed_origins": ["http://localhost:3000"],
    "token": {
      "hmac_secret_key": "your-secret-key"
    }
  },
  "http_api": {
    "key": "your-api-key"
  },
  "admin": {
    "enabled": true,
    "password": "admin",
    "secret": "admin-secret"
  }
}
```

### Generating Connection Tokens (Go)

```go
import (
    "github.com/centrifugal/centrifugo/v6/internal/jwtverify"
    "github.com/golang-jwt/jwt/v5"
)

func generateToken(userID string) (string, error) {
    claims := jwt.MapClaims{
        "sub": userID,
        "exp": time.Now().Add(time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte("your-secret-key"))
}
```

### Generating Connection Tokens (Node.js)

```javascript
const jwt = require('jsonwebtoken');

function generateToken(userId) {
  return jwt.sign(
    { sub: userId },
    'your-secret-key',
    { expiresIn: '1h' }
  );
}
```

### Publishing Messages (Go)

```go
import "github.com/centrifugal/centrifugo/v6/internal/api"

func publishNews(message string) error {
    data := map[string]interface{}{
        "html": fmt.Sprintf("<div class='news'>%s</div>", message),
    }

    return node.Publish("news", data)
}
```

### Publishing Messages (Node.js)

```javascript
const Centrifuge = require('centrifuge');

async function publishNews(message) {
  await centrifugo.publish('news', {
    html: `<div class="news">${message}</div>`
  });
}
```

## Browser Support

- Chrome/Edge: ✅ Full support
- Firefox: ✅ Full support
- Safari: ✅ Full support
- Mobile browsers: ✅ Full support

WebSocket is supported in all modern browsers. SSE fallback available for older browsers.

## Comparison with Other Solutions

### vs. htmx WebSocket Extension

| Feature | htmx-centrifugo | htmx ws extension |
|---------|----------------|-------------------|
| Reconnection | ✅ Automatic | ❌ Manual |
| Transport fallback | ✅ WebSocket/SSE/HTTP | ❌ WebSocket only |
| Scalability | ✅ Built-in (Redis/Nats) | ❌ Single server |
| Message recovery | ✅ Yes | ❌ No |
| History API | ✅ Yes | ❌ No |
| Presence | ✅ Yes | ❌ No |

### vs. htmx SSE Extension

| Feature | htmx-centrifugo | htmx sse extension |
|---------|----------------|-------------------|
| Bidirectional | ✅ Yes (RPC/publish) | ❌ Server→Client only |
| Transport fallback | ✅ Multiple options | ❌ SSE only |
| Scalability | ✅ Built-in | ❌ Single server |
| Reconnection | ✅ Sophisticated | ✅ Basic |
| Message recovery | ✅ Yes | ❌ No |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Links

- [Centrifugo Documentation](https://centrifugal.dev)
- [htmx Documentation](https://htmx.org)
- [Centrifuge JS SDK](https://github.com/centrifugal/centrifuge-js)
- [GitHub Repository](https://github.com/centrifugal/centrifugo)

## Support

- GitHub Issues: https://github.com/centrifugal/centrifugo/issues
- Centrifugo Community: https://t.me/joinchat/ABFVWBE0AhkyyhREoaboXQ
