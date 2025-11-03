"""
Database models and connection for htmx-centrifugo backend

Uses SQLAlchemy with async support for SQLite (dev) or PostgreSQL (prod)
"""

import os
from datetime import datetime
from typing import AsyncGenerator

from sqlalchemy import String, Integer, DateTime, Text
from sqlalchemy.ext.asyncio import AsyncSession, create_async_engine, async_sessionmaker
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column

# Database URL - supports SQLite and PostgreSQL
DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "sqlite+aiosqlite:///./chat.db"  # Default to SQLite for easy setup
)

# Create async engine
engine = create_async_engine(
    DATABASE_URL,
    echo=False,  # Set to True for SQL logging
    future=True
)

# Create session factory
AsyncSessionLocal = async_sessionmaker(
    engine,
    class_=AsyncSession,
    expire_on_commit=False
)


class Base(DeclarativeBase):
    """Base class for all models"""
    pass


class User(Base):
    """User model"""
    __tablename__ = "users"

    id: Mapped[str] = mapped_column(String(100), primary_key=True)
    username: Mapped[str] = mapped_column(String(50), nullable=False)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)
    updated_at: Mapped[datetime] = mapped_column(
        DateTime,
        default=datetime.utcnow,
        onupdate=datetime.utcnow
    )

    def __repr__(self):
        return f"<User(id={self.id}, username={self.username})>"


class Message(Base):
    """Chat message model"""
    __tablename__ = "messages"

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    user_id: Mapped[str] = mapped_column(String(100), nullable=False)
    username: Mapped[str] = mapped_column(String(50), nullable=False)
    channel: Mapped[str] = mapped_column(String(100), nullable=False, index=True)
    text: Mapped[str] = mapped_column(Text, nullable=False)
    message_type: Mapped[str] = mapped_column(
        String(20),
        default="chat",
        nullable=False
    )  # 'chat' or 'system'
    created_at: Mapped[datetime] = mapped_column(
        DateTime,
        default=datetime.utcnow,
        index=True
    )

    def __repr__(self):
        return f"<Message(id={self.id}, user={self.username}, channel={self.channel})>"


async def get_session() -> AsyncGenerator[AsyncSession, None]:
    """
    Dependency for getting database session

    Usage in FastAPI:
        @app.get("/messages")
        async def get_messages(session: AsyncSession = Depends(get_session)):
            ...
    """
    async with AsyncSessionLocal() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()


async def init_db():
    """
    Initialize database - create all tables

    Call this on startup
    """
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)


async def get_or_create_user(session: AsyncSession, user_id: str, username: str) -> User:
    """
    Get existing user or create new one

    Args:
        session: Database session
        user_id: User identifier
        username: Username

    Returns:
        User object
    """
    from sqlalchemy import select

    # Try to get existing user
    result = await session.execute(
        select(User).where(User.id == user_id)
    )
    user = result.scalar_one_or_none()

    if user:
        # Update username if changed
        if user.username != username:
            user.username = username
            user.updated_at = datetime.utcnow()
            await session.commit()
        return user

    # Create new user
    user = User(id=user_id, username=username)
    session.add(user)
    await session.commit()
    return user


async def save_message(
    session: AsyncSession,
    user_id: str,
    username: str,
    channel: str,
    text: str,
    message_type: str = "chat"
) -> Message:
    """
    Save a message to the database

    Args:
        session: Database session
        user_id: User identifier
        username: Username
        channel: Channel name
        text: Message text
        message_type: Type of message ('chat' or 'system')

    Returns:
        Saved message object
    """
    # Ensure user exists
    await get_or_create_user(session, user_id, username)

    # Create message
    message = Message(
        user_id=user_id,
        username=username,
        channel=channel,
        text=text,
        message_type=message_type
    )

    session.add(message)
    await session.commit()
    await session.refresh(message)

    return message


async def get_recent_messages(
    session: AsyncSession,
    channel: str,
    limit: int = 50
) -> list[Message]:
    """
    Get recent messages from a channel

    Args:
        session: Database session
        channel: Channel name
        limit: Maximum number of messages to return

    Returns:
        List of messages (ordered by creation time, oldest first)
    """
    from sqlalchemy import select

    result = await session.execute(
        select(Message)
        .where(Message.channel == channel)
        .order_by(Message.created_at.desc())
        .limit(limit)
    )

    messages = list(result.scalars().all())
    messages.reverse()  # Return oldest first
    return messages
