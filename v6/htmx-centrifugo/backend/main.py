"""
FastAPI backend for htmx-centrifugo chat example

This backend handles RPC calls from Centrifugo and publishes messages back.
Perfect for htmx users who are familiar with Python!
"""

import os
import logging
import hmac
import hashlib
import time
from datetime import datetime
from typing import Dict, Optional
import html

import httpx
from fastapi import FastAPI, Request, HTTPException, Depends, Form, Response, Cookie
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
from pydantic import BaseModel, Field
from sqlalchemy.ext.asyncio import AsyncSession

# Import database functions
from database import (
    init_db,
    get_session,
    save_message,
    get_recent_messages,
    get_or_create_user,
    User,
    Message
)

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

# Configuration
CENTRIFUGO_API_URL = os.getenv("CENTRIFUGO_API_URL", "http://centrifugo:8000/api")
CENTRIFUGO_API_KEY = os.getenv("CENTRIFUGO_API_KEY", "my-api-key")
CENTRIFUGO_TOKEN_SECRET = os.getenv("CENTRIFUGO_TOKEN_SECRET", "my-secret-key")
SESSION_SECRET = os.getenv("SESSION_SECRET", "session-secret-key-change-in-production")

# Simple in-memory session store (use Redis in production)
sessions: Dict[str, Dict] = {}

# Create FastAPI app
app = FastAPI(
    title="htmx-centrifugo Backend",
    description="FastAPI backend for htmx-centrifugo chat example with database persistence",
    version="1.0.0"
)

# Initialize Jinja2 templates
templates = Jinja2Templates(directory="templates")


@app.on_event("startup")
async def startup_event():
    """Initialize database on startup"""
    logger.info("Initializing database...")
    await init_db()
    logger.info("Database initialized")


# Session management helpers
def create_session_id(user_id: str) -> str:
    """Create a secure session ID"""
    timestamp = str(time.time())
    data = f"{user_id}{timestamp}{SESSION_SECRET}"
    return hashlib.sha256(data.encode()).hexdigest()


def get_session_data(session_id: Optional[str]) -> Optional[Dict]:
    """Get session data from session ID"""
    if not session_id or session_id not in sessions:
        return None

    session_data = sessions[session_id]

    # Check if session is expired (1 hour)
    if time.time() - session_data.get('created_at', 0) > 3600:
        del sessions[session_id]
        return None

    return session_data


# Pydantic models for data validation
class SendMessageData(BaseModel):
    message: str = Field(..., min_length=1, max_length=1000)


class SetUsernameData(BaseModel):
    username: str = Field(..., min_length=1, max_length=50)


# Centrifugo API client
async def publish_to_channel(channel: str, data: dict) -> dict:
    """
    Publish a message to a Centrifugo channel via HTTP API
    """
    async with httpx.AsyncClient() as client:
        response = await client.post(
            CENTRIFUGO_API_URL,
            headers={
                "Authorization": f"apikey {CENTRIFUGO_API_KEY}",
                "Content-Type": "application/json"
            },
            json={
                "method": "publish",
                "params": {
                    "channel": channel,
                    "data": data
                }
            },
            timeout=5.0
        )
        response.raise_for_status()
        return response.json()


def escape_html(text: str) -> str:
    """Escape HTML to prevent XSS attacks"""
    return html.escape(text)


def format_time() -> str:
    """Format current time for display"""
    return datetime.now().strftime("%H:%M")


def generate_centrifugo_token(user_id: str, exp_seconds: int = 3600) -> str:
    """
    Generate JWT token for Centrifugo connection

    Uses HMAC SHA256 to sign the token with the secret key.
    This is compatible with Centrifugo's token verification.

    Args:
        user_id: User identifier
        exp_seconds: Token expiration time in seconds (default: 1 hour)

    Returns:
        JWT token string
    """
    import json
    import base64

    # Header
    header = {
        "typ": "JWT",
        "alg": "HS256"
    }

    # Payload
    now = int(time.time())
    payload = {
        "sub": user_id,  # User ID
        "exp": now + exp_seconds,  # Expiration time
        "iat": now  # Issued at
    }

    # Encode header and payload
    def base64url_encode(data):
        json_bytes = json.dumps(data, separators=(',', ':')).encode('utf-8')
        return base64.urlsafe_b64encode(json_bytes).rstrip(b'=').decode('utf-8')

    header_encoded = base64url_encode(header)
    payload_encoded = base64url_encode(payload)

    # Create signature
    message = f"{header_encoded}.{payload_encoded}".encode('utf-8')
    signature = hmac.new(
        CENTRIFUGO_TOKEN_SECRET.encode('utf-8'),
        message,
        hashlib.sha256
    ).digest()
    signature_encoded = base64.urlsafe_b64encode(signature).rstrip(b'=').decode('utf-8')

    # Combine all parts
    token = f"{header_encoded}.{payload_encoded}.{signature_encoded}"

    return token


@app.post("/centrifugo/rpc")
async def handle_rpc(
    request: Request,
    session: AsyncSession = Depends(get_session)
):
    """
    RPC endpoint - handles all RPC calls proxied from Centrifugo

    Centrifugo sends RPC proxy requests with this structure:
    {
      "method": "sendChatMessage",
      "data": {...},
      "client": "client-id",
      "transport": "websocket",
      "user": "user-id"  # from JWT token "sub" claim
    }
    """
    # Parse the request body
    body = await request.json()

    logger.info(f"RPC proxy request: {body}")

    method = body.get("method")
    data = body.get("data", {})
    user_id = body.get("user")
    client_id = body.get("client", "unknown")

    if not user_id:
        raise HTTPException(status_code=401, detail="User not authenticated")

    logger.info(f"RPC call: {method} from user={user_id}, client={client_id}")

    try:
        if method == "sendChatMessage":
            return await handle_send_message(data, user_id, session)
        elif method == "setUsername":
            return await handle_set_username(data, user_id, session)
        else:
            return {
                "error": {
                    "code": 404,
                    "message": f"Unknown RPC method: {method}"
                }
            }
    except HTTPException as e:
        return {
            "error": {
                "code": e.status_code,
                "message": e.detail
            }
        }
    except Exception as e:
        logger.error(f"Error handling RPC: {e}", exc_info=True)
        return {
            "error": {
                "code": 500,
                "message": str(e)
            }
        }


async def handle_send_message(data: dict, user_id: str, session: AsyncSession) -> dict:
    """
    Handle sendChatMessage RPC call

    1. Validate message
    2. Get username from database
    3. Format message with HTML escaping
    4. Publish to chat channel
    5. Return success response
    """
    # Validate input
    try:
        message_data = SendMessageData(**data)
    except Exception as e:
        raise HTTPException(status_code=400, detail=f"Invalid message data: {e}")

    message_text = message_data.message.strip()

    if not message_text:
        raise HTTPException(status_code=400, detail="Message cannot be empty")

    # Get user from database (should exist from login)
    from sqlalchemy import select
    result = await session.execute(
        select(User).where(User.id == user_id)
    )
    user = result.scalar_one_or_none()

    if not user:
        raise HTTPException(status_code=404, detail="User not found. Please log in again.")

    username = user.username

    # Save message to database
    db_message = await save_message(
        session,
        user_id=user_id,
        username=username,
        channel="chat",
        text=message_text,
        message_type="chat"
    )

    # Create HTML for the message (with XSS protection)
    message_html = f"""
    <div class="message" data-message-id="{db_message.id}">
      <div class="message-author">{escape_html(username)}</div>
      <div class="message-text">{escape_html(message_text)}</div>
      <div class="message-time">{format_time()}</div>
    </div>
    """

    # Publish to chat channel
    await publish_to_channel("chat", {
        "html": message_html.strip(),
        "messageId": db_message.id,
        "userId": user_id,
        "username": username,
        "text": message_text,
        "timestamp": int(db_message.created_at.timestamp() * 1000)
    })

    logger.info(f"[Chat] {username}: {message_text} (saved as ID {db_message.id})")

    # Return success response
    return {
        "result": {
            "success": True,
            "messageId": db_message.id
        }
    }


async def handle_set_username(data: dict, user_id: str, session: AsyncSession) -> dict:
    """
    Handle setUsername RPC call

    1. Validate username
    2. Update user mapping
    3. Publish system message
    4. Return success response
    """
    # Validate input
    try:
        username_data = SetUsernameData(**data)
    except Exception as e:
        raise HTTPException(status_code=400, detail=f"Invalid username data: {e}")

    new_username = username_data.username.strip()

    # Get current user from database
    user = await get_or_create_user(session, user_id, new_username)
    old_username = user.username if user.username != new_username else "Anonymous"

    # Update username in database
    user.username = new_username
    await session.commit()

    # Create system message
    system_message_html = f"""
    <div class="message system-message">
      <div class="message-text"><em>{escape_html(old_username)} is now known as {escape_html(new_username)}</em></div>
      <div class="message-time">{format_time()}</div>
    </div>
    """

    # Save system message to database
    system_msg = await save_message(
        session,
        user_id="system",
        username="System",
        channel="chat",
        text=f"{old_username} is now known as {new_username}",
        message_type="system"
    )

    # Publish system message
    await publish_to_channel("chat", {
        "html": system_message_html.strip(),
        "type": "system",
        "text": f"{old_username} is now known as {new_username}",
        "timestamp": int(system_msg.created_at.timestamp() * 1000)
    })

    logger.info(f"[System] {old_username} → {new_username}")

    return {
        "result": {
            "success": True,
            "username": new_username
        }
    }


@app.get("/health")
async def health_check(session: AsyncSession = Depends(get_session)):
    """
    Health check endpoint

    Returns server status and statistics
    """
    from sqlalchemy import select, func

    # Get message count
    result = await session.execute(
        select(func.count(Message.id))
    )
    total_messages = result.scalar() or 0

    # Get user count
    result = await session.execute(
        select(func.count(User.id))
    )
    total_users = result.scalar() or 0

    return {
        "status": "ok",
        "service": "htmx-centrifugo-backend",
        "framework": "FastAPI",
        "timestamp": datetime.now().isoformat(),
        "stats": {
            "totalMessages": total_messages,
            "totalUsers": total_users
        }
    }


@app.get("/api/messages", response_class=HTMLResponse)
async def get_messages(
    channel: str = "chat",
    limit: int = 50,
    session: AsyncSession = Depends(get_session)
):
    """
    Get recent messages as HTML (for htmx)

    Args:
        channel: Channel name (default: "chat")
        limit: Maximum number of messages to return (default: 50)

    Returns:
        HTML fragment with messages
    """
    messages = await get_recent_messages(session, channel, limit)

    if not messages:
        return """
        <div class="message">
          <div class="message-author">System</div>
          <div class="message-text">Welcome to the chat! Send a message below.</div>
          <div class="message-time">Just now</div>
        </div>
        """

    # Build HTML for all messages
    html_parts = []
    for msg in messages:
        if msg.message_type == "chat":
            html_parts.append(f"""
        <div class="message" data-message-id="{msg.id}">
          <div class="message-author">{escape_html(msg.username)}</div>
          <div class="message-text">{escape_html(msg.text)}</div>
          <div class="message-time">{msg.created_at.strftime("%H:%M")}</div>
        </div>
            """)
        else:
            html_parts.append(f"""
        <div class="message system-message">
          <div class="message-text"><em>{escape_html(msg.text)}</em></div>
          <div class="message-time">{msg.created_at.strftime("%H:%M")}</div>
        </div>
            """)

    return "".join(html_parts)


@app.post("/auth/login")
async def login(
    request: Request,
    response: Response,
    username: str = Form(...),
    session: AsyncSession = Depends(get_session)
):
    """
    Login endpoint for htmx - returns HTML chat interface and sets session cookie

    Args:
        request: FastAPI request object (needed for templates)
        response: FastAPI response object for setting cookies
        username: Username from form

    Returns:
        HTML fragment with chat interface and auth token
    """
    username = username.strip()

    if not username:
        return templates.TemplateResponse(
            "login.html",
            {"request": request, "error": "Username cannot be empty"}
        )

    # Generate unique user ID from username
    user_id = hashlib.sha256(username.encode()).hexdigest()[:16]

    # Create or update user in database
    user = await get_or_create_user(session, user_id, username)

    # Generate Centrifugo token
    token = generate_centrifugo_token(user_id, exp_seconds=3600)

    # Create session
    session_id = create_session_id(user_id)
    sessions[session_id] = {
        'user_id': user_id,
        'username': username,
        'token': token,
        'created_at': time.time()
    }

    logger.info(f"User logged in: {username} (ID: {user_id}, Session: {session_id[:8]}...)")

    # Render chat template
    html_response = templates.TemplateResponse(
        "chat.html",
        {"request": request, "username": username, "token": token}
    )

    # Set session cookie
    html_response.set_cookie(
        key="session_id",
        value=session_id,
        max_age=3600,  # 1 hour
        httponly=True,
        samesite="lax"
    )
    return html_response


@app.post("/auth/logout")
async def logout_html(request: Request, session_id: Optional[str] = Cookie(None)):
    """
    Logout endpoint for htmx - returns login screen and clears session

    Args:
        request: FastAPI request object (needed for templates)
        session_id: Session cookie

    Returns:
        HTML fragment with login form
    """
    # Clear session from store
    if session_id and session_id in sessions:
        del sessions[session_id]
        logger.info(f"Session logged out: {session_id[:8]}...")

    # Render login template
    html_response = templates.TemplateResponse(
        "login.html",
        {"request": request, "error": None}
    )

    # Clear session cookie
    html_response.delete_cookie(key="session_id")
    return html_response




@app.get("/chat")
async def chat_page(
    request: Request,
    session_id: Optional[str] = Cookie(None),
    session: AsyncSession = Depends(get_session)
):
    """
    Main chat page - checks session and returns appropriate view

    Returns full HTML page (not fragment) with session already validated.
    No extra round-trip needed.
    """
    # Check if session exists and is valid
    session_data = get_session_data(session_id)

    if not session_data:
        # No valid session - return login page
        return templates.TemplateResponse(
            "login.html",
            {"request": request, "error": None}
        )

    # Valid session exists - return chat page
    username = session_data['username']
    token = session_data['token']

    logger.info(f"Chat page loaded for: {username}")

    return templates.TemplateResponse(
        "chat.html",
        {"request": request, "username": username, "token": token}
    )


@app.get("/")
async def root():
    """
    Root endpoint - API information
    """
    return {
        "name": "htmx-centrifugo Backend",
        "framework": "FastAPI",
        "description": "Backend server for htmx-centrifugo chat example with database persistence",
        "endpoints": {
            "chat": "GET /chat",
            "rpc": "POST /centrifugo/rpc",
            "login": "POST /auth/login",
            "logout": "POST /auth/logout",
            "messages": "GET /api/messages",
            "health": "GET /health"
        },
        "docs": "/docs",
        "redoc": "/redoc"
    }


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "4000"))

    logger.info(f"Starting FastAPI backend on port {port}")
    logger.info(f"Centrifugo API: {CENTRIFUGO_API_URL}")

    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        log_level="info"
    )
