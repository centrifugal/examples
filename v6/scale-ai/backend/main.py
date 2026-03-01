import asyncio
import os
import random
import uuid
from contextlib import asynccontextmanager

import asyncpg
import httpx
from fastapi import FastAPI

DATABASE_URL = os.environ.get(
    "DATABASE_URL", "postgresql://playground:playground@localhost:5432/playground"
)

CENTRIFUGO_API_URL = "http://centrifugo:8000/api/publish"
CENTRIFUGO_API_KEY = "secret"

VOCABULARY = [
    "the", "a", "is", "are", "was", "were", "will", "be", "have", "has",
    "do", "does", "did", "can", "could", "would", "should", "may", "might",
    "shall", "not", "and", "but", "or", "if", "then", "else", "when",
    "while", "for", "to", "from", "in", "on", "at", "by", "with", "about",
    "AI", "model", "token", "stream", "data", "server", "client", "real-time",
    "latency", "throughput", "WebSocket", "connection", "message", "publish",
    "subscribe", "channel", "event", "response", "request", "protocol",
    "neural", "network", "inference", "transformer", "attention", "layer",
    "output", "input", "process", "compute", "generate", "predict",
    "optimize", "scale", "deploy", "monitor", "aggregate", "buffer",
    "Centrifugo", "infrastructure", "scalable", "reliable", "efficient",
    "performance", "recovery", "history", "Redis", "distributed",
    "horizontal", "vertical", "cluster", "node", "replica", "shard",
    "pipeline", "workflow", "architecture", "system", "design", "pattern",
    "integration", "delivery", "streaming", "batch", "processing",
    "language", "understanding", "generation", "completion", "embedding",
]

FAKE_QUESTIONS = [
    "Explain how transformer attention mechanisms work in large language models.",
    "What are the key differences between supervised and unsupervised learning?",
    "How does reinforcement learning from human feedback improve AI assistants?",
    "Describe the architecture of a typical neural machine translation system.",
    "What is the role of tokenization in natural language processing pipelines?",
    "How do diffusion models generate high-quality images from text prompts?",
    "Explain the concept of embeddings and their use in semantic search.",
    "What are the main challenges in deploying LLMs to production at scale?",
    "How does retrieval-augmented generation combine search with text generation?",
    "What techniques help reduce hallucinations in large language models?",
]

SCHEMA_SQL = """
CREATE TABLE IF NOT EXISTS streams (
    id TEXT PRIMARY KEY,
    channel TEXT NOT NULL,
    question TEXT NOT NULL,
    answer TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'streaming',
    created_at TIMESTAMP DEFAULT NOW()
);
"""

db_pool: asyncpg.Pool | None = None
http_client: httpx.AsyncClient | None = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global db_pool, http_client
    db_pool = await asyncpg.create_pool(DATABASE_URL, min_size=2, max_size=10)
    http_client = httpx.AsyncClient()
    async with db_pool.acquire() as conn:
        await conn.execute(SCHEMA_SQL)
    yield
    await http_client.aclose()
    await db_pool.close()


app = FastAPI(lifespan=lifespan)


async def publish_to_centrifugo(channel: str, data: dict):
    await http_client.post(
        CENTRIFUGO_API_URL,
        json={"channel": channel, "data": data},
        headers={"X-API-Key": CENTRIFUGO_API_KEY},
    )


async def run_stream(
    stream_id: str,
    channel: str,
    tokens_per_second: int,
    total_tokens: int,
    aggregate_size: int,
):
    delay = 1.0 / tokens_per_second
    buffer = []
    answer_parts = []

    for i in range(total_tokens):
        word = random.choice(VOCABULARY)
        if i == 0:
            word = word.capitalize()
        buffer.append(word)

        if len(buffer) >= aggregate_size or i == total_tokens - 1:
            text = " ".join(buffer)
            buffer = []
            answer_parts.append(text)
            await publish_to_centrifugo(channel, {"text": text, "done": False})

        await asyncio.sleep(delay)

    await publish_to_centrifugo(channel, {"text": "", "done": True})

    full_answer = " ".join(answer_parts)
    async with db_pool.acquire() as conn:
        await conn.execute(
            "UPDATE streams SET answer = $1, status = 'done' WHERE id = $2",
            full_answer, stream_id,
        )


@app.post("/api/stream")
async def stream(req: dict):
    tokens_per_second = req.get("tokens_per_second", 30)
    total_tokens = req.get("total_tokens", 100)
    aggregate_size = req.get("aggregate_size", 1)

    stream_id = uuid.uuid4().hex[:12]
    channel = "ai:stream_" + stream_id
    question = random.choice(FAKE_QUESTIONS)

    async with db_pool.acquire() as conn:
        await conn.execute(
            "INSERT INTO streams (id, channel, question, status) VALUES ($1, $2, $3, 'streaming')",
            stream_id, channel, question,
        )

    asyncio.create_task(
        run_stream(stream_id, channel, tokens_per_second, total_tokens, aggregate_size)
    )

    await publish_to_centrifugo("ai:notifications", {
        "type": "stream_started",
        "id": stream_id,
        "channel": channel,
        "question": question,
    })

    return {"id": stream_id, "channel": channel, "question": question}


def row_to_dict(row):
    return {
        "id": row["id"],
        "channel": row["channel"],
        "question": row["question"],
        "answer": row["answer"],
        "status": row["status"],
    }


@app.get("/api/stream/active")
async def stream_active():
    async with db_pool.acquire() as conn:
        row = await conn.fetchrow(
            "SELECT id, channel, question, answer, status "
            "FROM streams ORDER BY created_at DESC LIMIT 1"
        )
    if not row:
        return {"stream": None}
    return {"stream": row_to_dict(row)}


@app.get("/api/stream/{stream_id}")
async def stream_by_id(stream_id: str):
    async with db_pool.acquire() as conn:
        row = await conn.fetchrow(
            "SELECT id, channel, question, answer, status FROM streams WHERE id = $1",
            stream_id,
        )
    if not row:
        return {"stream": None}
    return {"stream": row_to_dict(row)}
