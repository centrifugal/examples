# FastAPI Backend for htmx-centrifugo

A **Python/FastAPI** backend that handles RPC calls from Centrifugo for the chat example.

Perfect for htmx users! Since htmx is most popular in the Python ecosystem, this backend uses familiar Python tools.

## Why Python + FastAPI?

- 🐍 **Python** - The language htmx users know and love
- ⚡ **FastAPI** - Modern, fast (high-performance) async framework
- 📝 **Pydantic** - Automatic input validation
- 🔒 **Type hints** - Better IDE support and fewer bugs
- 📚 **Auto docs** - Built-in OpenAPI/Swagger UI

## How It Works

### RPC Proxy Flow

```
┌─────────┐         ┌────────────┐         ┌─────────┐         ┌─────────┐
│ Browser │ ──RPC──>│ Centrifugo │ ──HTTP─>│ FastAPI │ ──API──>│ Centrifugo │
│ (htmx)  │         │   (Proxy)  │         │ Backend │         │ (Publish)  │
└─────────┘         └────────────┘         └─────────┘         └─────────┘
                           │                                          │
                           └──────────── WebSocket ──────────────────┘
                                          (broadcast)
```

1. **Browser** sends RPC via htmx extension
2. **Centrifugo** proxies RPC to FastAPI
3. **FastAPI** validates, formats, and processes
4. **FastAPI** publishes to Centrifugo via HTTP API
5. **Centrifugo** broadcasts to all connected clients

## Features

### JWT Authentication

The backend implements JWT (JSON Web Token) authentication for Centrifugo connections:

- 🔐 **Token Generation** - `/auth/login` endpoint creates JWT tokens
- 👤 **User Identification** - Each user gets a unique ID based on username
- ⏰ **Token Expiration** - Tokens expire after 1 hour (configurable)
- 🔑 **HMAC SHA256** - Tokens signed with shared secret
- ✅ **Centrifugo Compatible** - Uses standard JWT format

**How it works:**
1. User submits username via login form
2. Backend generates unique user ID from username
3. Backend creates/updates user in database
4. Backend generates JWT token with user ID
5. Frontend uses token to connect to Centrifugo
6. Centrifugo validates token and establishes connection

**Token structure:**
```json
{
  "sub": "user-id-here",    // User identifier
  "exp": 1672656000,        // Expiration timestamp
  "iat": 1672652400         // Issued at timestamp
}
```

### Database Persistence

The backend uses **SQLAlchemy** with async support for database persistence:

- 🗄️ **SQLite** (default) - Easy setup for development and demos
- 🐘 **PostgreSQL** - Production-ready with async support
- 💾 **Message History** - All chat messages are persisted
- 👤 **User Management** - Usernames are stored and tracked
- ⚡ **Async I/O** - Non-blocking database operations

**Models:**
- `User` - User profiles with usernames
- `Message` - Chat messages with type, channel, timestamp

**Database operations:**
```python
# Save message
db_message = await save_message(
    session,
    user_id=user_id,
    username=username,
    channel="chat",
    text=message_text,
    message_type="chat"
)

# Get recent messages
messages = await get_recent_messages(session, channel="chat", limit=50)
```

### Implemented RPC Methods

#### `sendChatMessage`
Handles chat message submission with full validation.

**Python handler:**
```python
@app.post("/centrifugo/rpc")
async def handle_rpc(rpc_request: RPCRequest, request: Request):
    if rpc_request.method == "sendChatMessage":
        return await handle_send_message(rpc_request.data, user_id)
```

**Features:**
- ✅ Pydantic validation (1-1000 characters)
- ✅ HTML escaping (XSS prevention)
- ✅ User ID from Centrifugo headers
- ✅ Message formatting with timestamp
- ✅ Async HTTP client (httpx)
- ✅ Database persistence (messages saved to DB)

#### `setUsername`
Changes user's display name with validation.

**Features:**
- ✅ Length validation (1-50 characters)
- ✅ System message broadcast
- ✅ Username persistence

### Security Features

- **XSS Prevention** - `html.escape()` on all user input
- **Input Validation** - Pydantic models enforce constraints
- **Type Safety** - Type hints throughout
- **Async Safety** - Proper async/await patterns

### Automatic API Documentation

FastAPI generates interactive docs automatically:

- **Swagger UI**: http://localhost:4000/docs
- **ReDoc**: http://localhost:4000/redoc

## API Endpoints

### `POST /auth/login`
Login endpoint - creates or updates user and returns JWT token.

**Request:**
```json
{
  "username": "Alice"
}
```

**Response:**
```json
{
  "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
  "userId": "a1b2c3d4e5f6g7h8",
  "username": "Alice",
  "expiresIn": 3600
}
```

**Frontend usage:**
```javascript
// Login and get token
const response = await fetch('/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ username: 'Alice' })
});
const data = await response.json();

// Use token to connect to Centrifugo
container.setAttribute('centrifugo-token', data.token);
container.setAttribute('centrifugo-init', 'true');
```

### `GET /api/messages`
Get recent messages from a channel (message history).

**Query Parameters:**
- `channel` (optional, default: "chat") - Channel name
- `limit` (optional, default: 50) - Maximum number of messages

**Response:**
```json
{
  "channel": "chat",
  "messages": [
    {
      "id": 1,
      "userId": "user-123",
      "username": "Alice",
      "text": "Hello!",
      "messageType": "chat",
      "timestamp": 1672656000000,
      "html": "<div class=\"message\">...</div>"
    }
  ],
  "count": 1
}
```

### `POST /centrifugo/rpc`
RPC proxy endpoint - receives all RPC calls from Centrifugo.

**Request:**
```json
{
  "method": "sendChatMessage",
  "data": {
    "message": "Hello from Python!"
  }
}
```

**Response (success):**
```json
{
  "result": {
    "success": true,
    "messageId": 123
  }
}
```

**Response (error):**
```json
{
  "error": {
    "code": 400,
    "message": "Message cannot be empty"
  }
}
```

### `GET /health`
Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "service": "htmx-centrifugo-backend",
  "framework": "FastAPI",
  "timestamp": "2025-01-02T12:00:00",
  "stats": {
    "totalMessages": 42,
    "totalUsers": 5
  }
}
```

Note: Stats are now queried from the database.

### `GET /`
API information and available endpoints.

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `CENTRIFUGO_API_URL` | `http://centrifugo:8000/api` | Centrifugo HTTP API endpoint |
| `CENTRIFUGO_API_KEY` | `my-api-key` | API key for publishing |
| `CENTRIFUGO_TOKEN_SECRET` | `my-secret-key` | Secret key for JWT token signing (must match Centrifugo config) |
| `PORT` | `4000` | Server port |
| `DATABASE_URL` | `sqlite+aiosqlite:///./chat.db` | Database connection URL |

**Important:** The `CENTRIFUGO_TOKEN_SECRET` must match the `client.token.hmac_secret_key` in Centrifugo configuration!

## Running

### With Docker (Recommended)

```bash
# From project root
docker-compose up backend
```

### Without Docker

```bash
cd backend
pip install -r requirements.txt
python main.py
```

Or with uvicorn directly:
```bash
uvicorn main:app --host 0.0.0.0 --port 4000 --reload
```

### Development Mode

```bash
cd backend
pip install -r requirements.txt

# Run with auto-reload
uvicorn main:app --reload --port 4000
```

## Testing

### Test RPC directly

```bash
# Simulate Centrifugo RPC call
curl -X POST http://localhost:4000/centrifugo/rpc \
  -H "Content-Type: application/json" \
  -H "x-centrifugo-user-id: test-user" \
  -H "x-centrifugo-client-id: test-client" \
  -d '{
    "method": "sendChatMessage",
    "data": {
      "message": "Test from Python!"
    }
  }'
```

### Check health

```bash
curl http://localhost:4000/health
```

### Interactive API docs

Open browser to:
- http://localhost:4000/docs (Swagger UI)
- http://localhost:4000/redoc (ReDoc)

## Code Structure

### Pydantic Models

```python
class SendMessageData(BaseModel):
    message: str = Field(..., min_length=1, max_length=1000)

class SetUsernameData(BaseModel):
    username: str = Field(..., min_length=1, max_length=50)
```

Benefits:
- Automatic validation
- Clear error messages
- Type safety
- Auto-generated docs

### Async HTTP Client

```python
async def publish_to_channel(channel: str, data: dict) -> dict:
    async with httpx.AsyncClient() as client:
        response = await client.post(CENTRIFUGO_API_URL, ...)
        response.raise_for_status()
        return response.json()
```

Benefits:
- Non-blocking I/O
- Better performance
- Proper resource cleanup

### Security Helpers

```python
def escape_html(text: str) -> str:
    """Escape HTML to prevent XSS attacks"""
    return html.escape(text)
```

## Extending

### Add New RPC Method

1. Create Pydantic model:

```python
class MyMethodData(BaseModel):
    field1: str
    field2: int = Field(..., ge=0, le=100)
```

2. Create handler:

```python
async def handle_my_method(data: dict, user_id: str) -> dict:
    # Validate
    my_data = MyMethodData(**data)

    # Process
    result = await do_something(my_data)

    # Publish
    await publish_to_channel("my-channel", {
        "html": f"<div>{result}</div>"
    })

    # Return
    return {"result": {"success": True}}
```

3. Register in main handler:

```python
@app.post("/centrifugo/rpc")
async def handle_rpc(rpc_request: RPCRequest, request: Request):
    if rpc_request.method == "myMethod":
        return await handle_my_method(rpc_request.data, user_id)
```

### Switch to PostgreSQL

The default is SQLite for easy setup. For production, use PostgreSQL:

```bash
# Install asyncpg
pip install asyncpg

# Set DATABASE_URL
export DATABASE_URL="postgresql+asyncpg://user:password@localhost/dbname"

# Run the app
python main.py
```

**database.py already supports both SQLite and PostgreSQL!** Just change the `DATABASE_URL` environment variable.

Example PostgreSQL URL:
```
postgresql+asyncpg://postgres:password@db:5432/chat
```

### Add Authentication

```python
from fastapi import Depends, HTTPException, Header
from jose import jwt

async def verify_token(authorization: str = Header(...)):
    try:
        token = authorization.replace("Bearer ", "")
        payload = jwt.decode(token, SECRET_KEY, algorithms=["HS256"])
        return payload
    except Exception:
        raise HTTPException(status_code=401, detail="Invalid token")

@app.post("/centrifugo/rpc")
async def handle_rpc(
    rpc_request: RPCRequest,
    request: Request,
    user: dict = Depends(verify_token)
):
    # Handler with authentication
    pass
```

### Add Caching (Redis)

```python
from redis import asyncio as aioredis

redis = await aioredis.from_url("redis://localhost")

async def get_user(user_id: str) -> Optional[str]:
    # Try cache first
    cached = await redis.get(f"user:{user_id}")
    if cached:
        return cached.decode()

    # Fetch from DB and cache
    user = await db.get_user(user_id)
    await redis.setex(f"user:{user_id}", 3600, user)
    return user
```

## Production Checklist

### Security
- [ ] Enable HTTPS
- [ ] Add rate limiting (slowapi)
- [ ] Implement authentication
- [ ] Validate all inputs
- [ ] Use environment variables for secrets
- [ ] Enable CORS properly

### Performance
- [ ] Use database connection pooling
- [ ] Add Redis for caching
- [ ] Enable response compression
- [ ] Use production ASGI server (Gunicorn + Uvicorn)
- [ ] Add CDN for static assets

### Reliability
- [ ] Add error tracking (Sentry)
- [ ] Implement retries for Centrifugo API
- [ ] Add health checks
- [ ] Use process manager (systemd, supervisor)
- [ ] Add graceful shutdown

### Monitoring
- [ ] Add logging (structlog)
- [ ] Add metrics (Prometheus)
- [ ] Add tracing (OpenTelemetry)
- [ ] Set up alerts

## Dependencies

```
fastapi           - Web framework
uvicorn           - ASGI server
httpx             - Async HTTP client
pydantic          - Data validation
sqlalchemy        - Async ORM
aiosqlite         - Async SQLite driver
```

All pinned in `requirements.txt`.

For PostgreSQL, also install: `pip install asyncpg`

## Why FastAPI?

Perfect for htmx users because:

1. **Pythonic** - Natural for Django/Flask users
2. **Fast** - One of the fastest Python frameworks
3. **Type hints** - Modern Python with IDE support
4. **Auto validation** - Pydantic handles input validation
5. **Auto docs** - OpenAPI/Swagger built-in
6. **Async support** - Non-blocking I/O for real-time
7. **Easy testing** - Built-in test client

## Learn More

- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Pydantic Documentation](https://docs.pydantic.dev/)
- [httpx Documentation](https://www.python-httpx.org/)
- [Centrifugo RPC Proxy](https://centrifugal.dev/docs/server/proxy#rpc-proxy)

## Files

- `main.py` - FastAPI application with RPC handlers
- `database.py` - SQLAlchemy models and database functions
- `requirements.txt` - Python dependencies
- `Dockerfile` - Container image
- `README.md` - This file

## Troubleshooting

### Import errors

```bash
pip install -r requirements.txt
```

### Port already in use

```bash
# Change port
PORT=4001 python main.py
```

### Can't reach Centrifugo

Check network:
```bash
curl http://centrifugo:8000/health
# or
curl http://localhost:8000/health
```

### Validation errors

Check Pydantic models - they enforce constraints automatically.
Error messages are clear and include which field failed.
