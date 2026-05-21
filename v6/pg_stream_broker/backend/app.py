"""PG Stream Broker + getState demo (Kitchen Orders).

The app's `orders` table is the source of truth for state. Publications
are written to Centrifugo's PG stream broker inside the same transaction
as the app state change — so state + notification commit atomically.

Clients use a regular stream Subscription with a position-only `getState`
callback:
  1. SDK calls getState() on subscribe.
  2. App reads stream top_position FIRST, then its own DB rows.
  3. App renders state, returns {offset, epoch}.
  4. SDK subscribes from that position — any publications committed
     between the read and the subscribe arrive as 'publication' events.

There is no map broker and no broker-side state table in this pattern.
"""

from __future__ import annotations

import asyncio
import json
import logging
import os
import random
import time
import uuid
from datetime import timedelta

import asyncpg
import httpx
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("pg_stream_broker_demo")

DATABASE_URL = os.environ.get("DATABASE_URL", "postgres://test:test@postgres:5432/test")
CENTRIFUGO_API_URL = os.environ.get("CENTRIFUGO_API_URL", "http://centrifugo:8000/api")

ORDERS_CHANNEL = "orders:kitchen"

app = FastAPI()
pool: asyncpg.Pool | None = None
http_client: httpx.AsyncClient | None = None


# ---------------------------------------------------------------------------
# PG stream broker helpers
# ---------------------------------------------------------------------------
# cf_stream_publish writes to the stream + meta tables and fires a pg_notify
# that Centrifugo's outbox worker picks up. Running this inside an app tx
# makes state mutation + notification atomic.
async def pg_stream_publish(conn, channel: str, payload: dict,
                            meta_ttl: timedelta | None = None) -> dict:
    row = await conn.fetchrow(
        """
        SELECT * FROM cf_stream_publish(
            p_channel := $1,
            p_data    := $2::jsonb,
            p_meta_ttl := $3
        )
        """,
        channel, json.dumps(payload), meta_ttl,
    )
    return {"offset": row["out_channel_offset"], "epoch": row["out_epoch"]}


# cf_stream_top_position returns the current stream top for a channel.
# Called FIRST inside getState so the position is a lower bound.
async def pg_stream_top_position(conn, channel: str) -> dict:
    row = await conn.fetchrow(
        "SELECT * FROM cf_stream_top_position($1)", channel,
    )
    return {"offset": int(row["out_top_offset"]), "epoch": row["out_epoch"]}


# ---------------------------------------------------------------------------
# Orders REST API
# ---------------------------------------------------------------------------
@app.get("/api/orders/state")
async def orders_get_state():
    """Return all active orders + current stream position.

    Position is read FIRST within a REPEATABLE READ transaction, then the
    rows — this guarantees the returned position is a lower bound on any
    data included in `entries`. Any publications committed after this
    read will be delivered via stream catch-up to the SDK.
    """
    async with pool.acquire() as conn:
        async with conn.transaction(isolation="repeatable_read"):
            pos = await pg_stream_top_position(conn, ORDERS_CHANNEL)
            rows = await conn.fetch(
                "SELECT id, table_number, items, status, notes, customer_name, "
                "       color, created_at, updated_at "
                "FROM orders WHERE status != 'cancelled' ORDER BY created_at ASC"
            )

    entries = []
    for row in rows:
        items = row["items"] if isinstance(row["items"], list) else json.loads(row["items"])
        entries.append({
            "key": row["id"],
            "data": {
                "tableNumber": row["table_number"],
                "items": items,
                "status": row["status"],
                "notes": row["notes"],
                "customerName": row["customer_name"],
                "color": row["color"],
                "createdAt": row["created_at"],
                "updatedAt": row["updated_at"],
            }
        })

    return JSONResponse({"entries": entries, "offset": pos["offset"], "epoch": pos["epoch"]})


@app.post("/api/orders/create")
async def orders_create(request: Request):
    data = await request.json()
    order_id = "order_" + uuid.uuid4().hex[:8]
    now = int(time.time() * 1000)
    table_number = data.get("tableNumber", 1)
    items = data.get("items", [])
    notes = data.get("notes", "")
    customer_name = data.get("customerName", "Guest")
    color = data.get("color", "#888")

    order_data = {
        "tableNumber": table_number,
        "items": items,
        "status": "pending",
        "notes": notes,
        "customerName": customer_name,
        "color": color,
        "createdAt": now,
        "updatedAt": now,
    }

    async with pool.acquire() as conn:
        async with conn.transaction():
            await conn.execute(
                """INSERT INTO orders (id, table_number, items, status, notes,
                                       customer_name, color, created_at, updated_at)
                   VALUES ($1, $2, $3::jsonb, $4, $5, $6, $7, $8, $8)""",
                order_id, table_number, json.dumps(items), "pending",
                notes, customer_name, color, now,
            )
            # Publication carries the key inside the payload. Clients apply
            # it to their own in-memory map keyed by order_id.
            await pg_stream_publish(conn, ORDERS_CHANNEL, {
                "key": order_id,
                "data": order_data,
            })

    return JSONResponse({"orderId": order_id})


@app.post("/api/orders/update-status")
async def orders_update_status(request: Request):
    data = await request.json()
    order_id = data.get("orderId", "")
    new_status = data.get("status", "")

    valid = ["pending", "preparing", "ready", "served", "cancelled"]
    if new_status not in valid:
        return JSONResponse({"error": "invalid status"}, status_code=400)

    async with pool.acquire() as conn:
        async with conn.transaction():
            row = await conn.fetchrow(
                "SELECT id, table_number, items, status, notes, customer_name, "
                "       color, created_at, updated_at "
                "FROM orders WHERE id = $1 FOR UPDATE",
                order_id,
            )
            if not row:
                return JSONResponse({"error": "order not found"}, status_code=404)

            now = int(time.time() * 1000)
            await conn.execute(
                "UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3",
                new_status, now, order_id,
            )

            items = row["items"] if isinstance(row["items"], list) else json.loads(row["items"])
            order_data = {
                "tableNumber": row["table_number"],
                "items": items,
                "status": new_status,
                "notes": row["notes"],
                "customerName": row["customer_name"],
                "color": row["color"],
                "createdAt": row["created_at"],
                "updatedAt": now,
            }

            if new_status == "cancelled":
                # Cancellation -> publish a tombstone. Clients drop the entry.
                await pg_stream_publish(conn, ORDERS_CHANNEL, {
                    "key": order_id,
                    "removed": True,
                })
            else:
                await pg_stream_publish(conn, ORDERS_CHANNEL, {
                    "key": order_id,
                    "data": order_data,
                })

    return JSONResponse({"success": True})


# ---------------------------------------------------------------------------
# Background task: demo order generator
# ---------------------------------------------------------------------------
SAMPLE_MENU = [
    {"name": "Margherita Pizza", "emoji": "\U0001f355"},
    {"name": "Caesar Salad", "emoji": "\U0001f957"},
    {"name": "Grilled Salmon", "emoji": "\U0001f41f"},
    {"name": "Pasta Carbonara", "emoji": "\U0001f35d"},
    {"name": "Beef Burger", "emoji": "\U0001f354"},
    {"name": "Chicken Wings", "emoji": "\U0001f357"},
    {"name": "Mushroom Risotto", "emoji": "\U0001f344"},
    {"name": "Fish & Chips", "emoji": "\U0001f420"},
    {"name": "Tom Yum Soup", "emoji": "\U0001f35c"},
    {"name": "Tiramisu", "emoji": "\U0001f370"},
]

CUSTOMERS = ["Alice", "Bob", "Carlos", "Diana", "Emma", "Felix",
             "Grace", "Hugo", "Iris", "Jack"]


async def orders_demo_task():
    await pool.execute("DELETE FROM orders")
    await asyncio.sleep(5)

    colors = ['#e91e63', '#9c27b0', '#3f51b5', '#00bcd4',
              '#4caf50', '#ff9800', '#f44336', '#009688']

    while True:
        try:
            await asyncio.sleep(random.uniform(8, 15))

            items = [
                {"name": m["name"], "emoji": m["emoji"], "qty": random.randint(1, 2)}
                for m in random.sample(SAMPLE_MENU, random.randint(1, 3))
            ]
            resp = await http_client.post(
                "http://localhost:5000/api/orders/create",
                json={
                    "tableNumber": random.randint(1, 12),
                    "items": items,
                    "customerName": random.choice(CUSTOMERS),
                    "color": random.choice(colors),
                },
                timeout=5.0,
            )
            if resp.status_code != 200:
                continue
            order_id = resp.json().get("orderId")
            if not order_id:
                continue

            for status, wait in [("preparing", (5, 10)),
                                 ("ready",     (8, 15)),
                                 ("served",    (5, 10))]:
                await asyncio.sleep(random.uniform(*wait))
                await http_client.post(
                    "http://localhost:5000/api/orders/update-status",
                    json={"orderId": order_id, "status": status},
                    timeout=5.0,
                )
        except Exception:
            logger.exception("orders_demo_task error")
            await asyncio.sleep(5)


# ---------------------------------------------------------------------------
# Lifecycle
# ---------------------------------------------------------------------------
@app.on_event("startup")
async def startup():
    global pool, http_client

    for attempt in range(30):
        try:
            pool = await asyncpg.create_pool(DATABASE_URL, min_size=2, max_size=10)
            break
        except Exception:
            logger.info("waiting for PostgreSQL... (%d)", attempt + 1)
            await asyncio.sleep(2)
    else:
        raise RuntimeError("PostgreSQL unavailable")

    http_client = httpx.AsyncClient(timeout=10.0)

    await pool.execute("""
        CREATE TABLE IF NOT EXISTS orders (
            id TEXT PRIMARY KEY,
            table_number INT NOT NULL,
            items JSONB NOT NULL DEFAULT '[]',
            status TEXT NOT NULL DEFAULT 'pending',
            notes TEXT NOT NULL DEFAULT '',
            customer_name TEXT NOT NULL DEFAULT '',
            color TEXT NOT NULL DEFAULT '#888',
            created_at BIGINT NOT NULL DEFAULT 0,
            updated_at BIGINT NOT NULL DEFAULT 0
        )
    """)

    asyncio.create_task(orders_demo_task())
    logger.info("pg_stream_broker demo started")


@app.on_event("shutdown")
async def shutdown():
    if http_client:
        await http_client.aclose()
    if pool:
        await pool.close()
