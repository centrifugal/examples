import asyncio
import base64
import json
import logging
import os
import random
import string
import time
import uuid

import asyncpg
import httpx
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("map_demo")

app = FastAPI()

# ---------------------------------------------------------------------------
# Global state
# ---------------------------------------------------------------------------
pool: asyncpg.Pool | None = None
http_client: httpx.AsyncClient | None = None

CENTRIFUGO_API_URL = os.environ.get("CENTRIFUGO_API_URL", "http://localhost:8000/api")
CENTRIFUGO_API_KEY = os.environ.get("CENTRIFUGO_API_KEY", "map-demo-api-key")
DATABASE_URL = os.environ.get("DATABASE_URL", "postgres://map_demo:map_demo@localhost:5432/map_demo")


# ===================================================================
# Centrifugo HTTP API helper
# ===================================================================
class CentrifugoAPI:
    def __init__(self, client: httpx.AsyncClient, base_url: str, api_key: str):
        self.client = client
        self.base_url = base_url
        self.headers = {"X-API-Key": api_key, "Content-Type": "application/json"}

    async def _call(self, method: str, payload: dict) -> dict:
        url = f"{self.base_url}/{method}"
        resp = await self.client.post(url, json=payload, headers=self.headers)
        resp.raise_for_status()
        return resp.json()

    async def map_publish(self, channel: str, key: str, data: dict, **kwargs) -> dict:
        payload: dict = {"channel": channel, "key": key, "data": data}
        if kwargs.get("tags"):
            payload["tags"] = kwargs["tags"]
        if kwargs.get("delta"):
            payload["delta"] = True
        if kwargs.get("key_mode"):
            payload["key_mode"] = kwargs["key_mode"]
        if kwargs.get("idempotency_key"):
            payload["idempotency_key"] = kwargs["idempotency_key"]
        if kwargs.get("version") is not None:
            payload["version"] = kwargs["version"]
        if kwargs.get("version_epoch"):
            payload["version_epoch"] = kwargs["version_epoch"]
        if kwargs.get("score") is not None:
            payload["score"] = kwargs["score"]
        return await self._call("map_publish", payload)

    async def map_remove(self, channel: str, key: str) -> dict:
        return await self._call("map_remove", {"channel": channel, "key": key})

    async def map_read_state(self, channel: str, *, key: str | None = None,
                             limit: int | None = None) -> dict:
        payload: dict = {"channel": channel}
        if key is not None:
            payload["key"] = key
        if limit is not None:
            payload["limit"] = limit
        return await self._call("map_read_state", payload)

    async def map_stats(self, channel: str) -> dict:
        return await self._call("map_stats", {"channel": channel})

    async def map_clear(self, channel: str) -> dict:
        return await self._call("map_clear", {"channel": channel})


centrifugo: CentrifugoAPI | None = None


def decode_entry_data(raw) -> dict:
    """Decode entry data — may be a dict (JSON) or a base64-encoded string."""
    if isinstance(raw, dict):
        return raw
    return json.loads(base64.b64decode(raw))


# ===================================================================
# PostgreSQL helpers (cf_map_publish / cf_map_remove)
# ===================================================================
async def pg_map_publish(
    channel: str, key: str, data: dict, *,
    key_mode: str | None = None,
) -> dict:
    row = await pool.fetchrow(
        """
        SELECT * FROM cf_map_publish(
            p_channel := $1,
            p_key := $2,
            p_data := $3::jsonb,
            p_key_mode := $4
        )
        """,
        channel, key, json.dumps(data), key_mode,
    )
    return {
        "offset": row["channel_offset"],
        "epoch": row["epoch"],
        "suppressed": row["suppressed"],
        "suppress_reason": row["suppress_reason"],
    }


async def pg_map_remove(channel: str, key: str) -> dict:
    row = await pool.fetchrow(
        "SELECT * FROM cf_map_remove(p_channel := $1, p_key := $2)",
        channel, key,
    )
    return {
        "offset": row["channel_offset"],
        "epoch": row["epoch"],
        "suppressed": row["suppressed"],
    }


# ===================================================================
# RPC proxy handler
# ===================================================================
@app.post("/centrifugo/rpc")
async def rpc_handler(request: Request):
    body = await request.json()
    method = body.get("method", "")
    data = body.get("data", {})
    user = body.get("user", "")
    client = body.get("client", "")

    if isinstance(data, str):
        data = json.loads(data)

    try:
        if method == "game:create":
            result = await handle_game_create(data, user, client)
        elif method == "game:join":
            result = await handle_game_join(data, user, client)
        elif method == "game:leave":
            result = await handle_game_leave(data, user, client)
        elif method == "inventory:buy":
            result = await handle_inventory_buy(data, user, client)
        elif method == "inventory:restock":
            result = await handle_inventory_restock(data, user, client)
        else:
            return JSONResponse({"error": {"code": 1, "message": f"unknown method: {method}"}})
        return JSONResponse({"result": {"data": result}})
    except Exception as e:
        logger.exception("RPC error: %s", method)
        return JSONResponse({"error": {"code": 100, "message": str(e)}})


# ===================================================================
# Game lobby (Centrifugo HTTP API via RPC proxy)
# ===================================================================
async def handle_game_create(data: dict, user: str, client: str) -> dict:
    game_id = "game_" + "".join(random.choices(string.ascii_lowercase + string.digits, k=8))
    name = data.get("name", "Untitled")
    max_players = data.get("maxPlayers", 2)
    game_data = {"name": name, "maxPlayers": max_players}
    await centrifugo.map_publish("games:lobby", game_id, game_data)
    return {"gameId": game_id}


async def handle_game_join(data: dict, user: str, client: str) -> dict:
    game_id = data.get("gameId", "")
    slot = data.get("slot", 1)
    name = data.get("name", "Anonymous")
    slot_key = f"slot_{slot}"

    # Verify game exists.
    state = await centrifugo.map_read_state("games:lobby", key=game_id)
    entries = state.get("result", {}).get("entries", [])
    if not entries:
        raise Exception("game not found")

    game_data = decode_entry_data(entries[0]["data"])
    max_players = game_data.get("maxPlayers", 2)

    game_channel = f"game:{game_id}"
    player_data = {"userId": user or client, "name": name, "slot": slot}
    result = await centrifugo.map_publish(game_channel, slot_key, player_data, key_mode="if_new")

    result_data = result.get("result", {})
    if result_data.get("suppressed"):
        raise Exception("slot already taken")

    # Check if game is full — spawn background check.
    asyncio.create_task(_check_game_full(game_id, game_channel, max_players))
    return {"success": True}


async def _check_game_full(game_id: str, game_channel: str, max_players: int):
    await asyncio.sleep(0.3)
    try:
        state = await centrifugo.map_read_state(game_channel, limit=max_players + 5)
        entries = state.get("result", {}).get("entries", [])
        slot_entries = [e for e in entries if e["key"].startswith("slot_")]
        if len(slot_entries) >= max_players:
            players = []
            for e in slot_entries:
                pd = decode_entry_data(e["data"])
                players.append({"userId": pd["userId"], "name": pd["name"]})
            await centrifugo.map_publish(
                game_channel, "game_event",
                {"event": "game_start", "players": players},
            )
            await asyncio.sleep(3)
            # Cleanup: remove all keys from game channel and games list.
            for e in entries:
                try:
                    await centrifugo.map_remove(game_channel, e["key"])
                except Exception:
                    pass
            try:
                await centrifugo.map_remove("games:lobby", game_id)
            except Exception:
                pass
    except Exception:
        logger.exception("check_game_full error")


async def handle_game_leave(data: dict, user: str, client: str) -> dict:
    game_id = data.get("gameId", "")
    slot = data.get("slot", 1)
    game_channel = f"game:{game_id}"
    await centrifugo.map_remove(game_channel, f"slot_{slot}")
    return {"success": True}


# ===================================================================
# Inventory (Centrifugo HTTP API via RPC proxy — CAS)
# ===================================================================
INVENTORY_ITEMS = [
    {"id": "golden_ticket", "emoji": "🎫", "name": "Golden Ticket", "price": 100, "stock": 3},
    {"id": "rare_potion", "emoji": "🧪", "name": "Rare Potion", "price": 50, "stock": 5},
    {"id": "dragon_egg", "emoji": "🥚", "name": "Dragon Egg", "price": 500, "stock": 1},
]


async def _init_inventory():
    """Publish initial inventory items with key_mode=if_new (idempotent)."""
    for item in INVENTORY_ITEMS:
        data = {"item": dict(item)}
        await centrifugo.map_publish("inventory:main", item["id"], data, key_mode="if_new")
    logger.info("Inventory initialized")


async def handle_inventory_buy(data: dict, user: str, client: str) -> dict:
    item_id = data.get("itemId", "")
    quantity = data.get("quantity", 1)

    # Artificial delay to demonstrate CAS contention.
    await asyncio.sleep(2)

    for attempt in range(5):
        if attempt > 0:
            await asyncio.sleep(0.1)
        state = await centrifugo.map_read_state("inventory:main", key=item_id)
        result_data = state.get("result", {})
        entries = result_data.get("entries", [])
        if not entries:
            return {"success": False, "message": "Item not found"}

        entry = entries[0]
        item_data = decode_entry_data(entry["data"])
        item = item_data.get("item", {})
        stock = item.get("stock", 0)

        if stock < quantity:
            return {"success": False, "message": "Out of stock"}

        item["stock"] = stock - quantity
        new_data = {
            "item": item,
            "transaction": {
                "action": "purchase",
                "message": f"{user or client} bought {quantity}x {item['name']}",
            },
        }

        result = await centrifugo.map_publish(
            "inventory:main", item_id, new_data,
            version=entry["offset"],
            version_epoch=result_data.get("epoch", ""),
        )
        r = result.get("result", {})
        if not r.get("suppressed"):
            return {"success": True, "message": f"Purchased {quantity}x {item['name']}", "attempts": attempt + 1}
        if r.get("suppress_reason") == "position_mismatch":
            continue
        return {"success": False, "message": f"Suppressed: {r.get('suppress_reason')}"}

    return {"success": False, "message": "Failed after 5 CAS attempts"}


async def handle_inventory_restock(data: dict, user: str, client: str) -> dict:
    item_id = data.get("itemId", "")
    quantity = data.get("quantity", 1)

    for attempt in range(5):
        if attempt > 0:
            await asyncio.sleep(0.1)
        state = await centrifugo.map_read_state("inventory:main", key=item_id)
        result_data = state.get("result", {})
        entries = result_data.get("entries", [])
        if not entries:
            return {"success": False, "message": "Item not found"}

        entry = entries[0]
        item_data = decode_entry_data(entry["data"])
        item = item_data.get("item", {})
        item["stock"] = item.get("stock", 0) + quantity
        new_data = {
            "item": item,
            "transaction": {
                "action": "restock",
                "message": f"Restocked {quantity}x {item['name']}",
            },
        }

        result = await centrifugo.map_publish(
            "inventory:main", item_id, new_data,
            version=entry["offset"],
            version_epoch=result_data.get("epoch", ""),
        )
        r = result.get("result", {})
        if not r.get("suppressed"):
            return {"success": True, "message": f"Restocked {quantity}x {item['name']}"}
        if r.get("suppress_reason") == "position_mismatch":
            continue
        return {"success": False, "message": f"Suppressed: {r.get('suppress_reason')}"}

    return {"success": False, "message": "Failed after 5 CAS attempts"}


# ===================================================================
# Polls (PostgreSQL transactional)
# ===================================================================
@app.post("/api/poll/vote")
async def poll_vote(request: Request):
    data = await request.json()
    user_id = data.get("userId", "")
    option_id = data.get("optionId", "")

    if not pool:
        return JSONResponse({"error": "PostgreSQL not configured"}, status_code=503)

    try:
        result = await _record_poll_vote(option_id, user_id)
        return JSONResponse(result)
    except Exception as e:
        logger.exception("poll vote error")
        return JSONResponse({"success": False, "message": str(e)})


async def _record_poll_vote(option_id: str, user_id: str | None = None) -> dict:
    """Record a vote. If user_id is provided, dedup via poll:votes."""
    # Extract poll_id from option_id (format: "{pollId}_opt_{N}").
    parts = option_id.rsplit("_opt_", 1)
    poll_id = parts[0] if len(parts) == 2 else "unknown"

    async with pool.acquire() as conn:
        async with conn.transaction():
            # 1. Dedup (only for real users, not bots).
            if user_id:
                vote_key = f"{poll_id}:{option_id}:{user_id}"
                dedup = await conn.fetchrow(
                    """
                    SELECT * FROM cf_map_publish(
                        p_channel := 'poll:votes',
                        p_key := $1,
                        p_data := $2::jsonb,
                        p_key_mode := 'if_new'
                    )
                    """,
                    vote_key, json.dumps({"voted": True}),
                )
                if dedup["suppressed"]:
                    return {"success": False, "message": "already voted"}

            # 2. Read current score with lock.
            row = await conn.fetchrow(
                "SELECT score FROM cf_map_state WHERE channel = 'poll:results' AND key = $1 FOR UPDATE",
                option_id,
            )
            if not row:
                return {"success": False, "message": "option not found"}

            new_score = (row["score"] or 0) + 1

            # 3. Publish updated score.
            await conn.fetchrow(
                """
                SELECT * FROM cf_map_publish(
                    p_channel := 'poll:results',
                    p_key := $1,
                    p_data := (SELECT data FROM cf_map_state WHERE channel = 'poll:results' AND key = $1)::jsonb,
                    p_score := $2
                )
                """,
                option_id, new_score,
            )

    return {"success": True}


# ===================================================================
# Board (PostgreSQL direct)
# ===================================================================
@app.post("/api/board/create")
async def board_create(request: Request):
    data = await request.json()
    task_id = "task:" + str(uuid.uuid4())[:8]
    now = int(time.time() * 1000)
    task_data = {
        "title": data.get("title", "Untitled"),
        "status": data.get("status", "todo"),
        "priority": data.get("priority", "medium"),
        "assignee": data.get("assignee", ""),
        "author": data.get("author", "Anonymous"),
        "color": data.get("color", "#888"),
        "created": now,
    }
    await pg_map_publish("board:main", task_id, task_data)
    return JSONResponse({"taskId": task_id})


@app.post("/api/board/update")
async def board_update(request: Request):
    data = await request.json()
    task_id = data.get("taskId", "")
    task_data = {
        "title": data.get("title", ""),
        "status": data.get("status", "todo"),
        "priority": data.get("priority", "medium"),
        "assignee": data.get("assignee", ""),
        "author": data.get("author", ""),
        "color": data.get("color", "#888"),
        "created": data.get("created", 0),
    }
    await pg_map_publish("board:main", task_id, task_data)
    return JSONResponse({"success": True})


@app.post("/api/board/delete")
async def board_delete(request: Request):
    data = await request.json()
    task_id = data.get("taskId", "")
    await pg_map_remove("board:main", task_id)
    return JSONResponse({"success": True})


# ===================================================================
# Visualizer (Centrifugo HTTP API)
# ===================================================================
@app.post("/api/viz/populate")
async def viz_populate(request: Request):
    data = await request.json()
    count = min(data.get("count", 10), 100000)
    for i in range(count):
        key = f"item_{i}"
        value = {"index": i, "value": random.random(), "ts": time.time()}
        await centrifugo.map_publish("visualizer:main", key, value)
    return JSONResponse({"populated": count})


@app.post("/api/viz/publish")
async def viz_publish(request: Request):
    data = await request.json()
    key = data.get("key", f"item_{uuid.uuid4().hex[:8]}")
    value = data.get("value", {"ts": time.time()})
    await centrifugo.map_publish("visualizer:main", key, value)
    return JSONResponse({"success": True})


@app.post("/api/viz/remove")
async def viz_remove(request: Request):
    data = await request.json()
    key = data.get("key", "")
    await centrifugo.map_remove("visualizer:main", key)
    return JSONResponse({"success": True})


@app.post("/api/viz/clear")
async def viz_clear(request: Request):
    await centrifugo.map_clear("visualizer:main")
    return JSONResponse({"success": True})


@app.get("/api/viz/stats")
async def viz_stats():
    result = await centrifugo.map_stats("visualizer:main")
    return JSONResponse(result.get("result", {}))


# ===================================================================
# Background task: Ticker
# ===================================================================
_SECTORS = ["tech", "ecommerce", "auto", "media", "enterprise"]
_SEED_TICKERS = [
    ("AAPL", 178.50), ("GOOG", 141.20), ("AMZN", 178.90), ("MSFT", 378.50),
    ("TSLA", 248.50), ("META", 505.75), ("NFLX", 628.30), ("NVDA", 875.40),
    ("AMD", 172.60), ("UBER", 78.25), ("SHOP", 78.60), ("CRM", 272.30),
]

def _generate_tickers(n: int) -> list[dict]:
    tickers = []
    for sym, base in _SEED_TICKERS:
        tickers.append({"symbol": sym, "base": base, "sector": random.choice(_SECTORS)})
    for i in range(len(_SEED_TICKERS), n):
        sym = f"T{i:04d}"
        base = round(random.uniform(5.0, 1000.0), 2)
        sector = _SECTORS[i % len(_SECTORS)]
        tickers.append({"symbol": sym, "base": base, "sector": sector})
    return tickers

TICKERS = _generate_tickers(1000)

ticker_prices: dict[str, float] = {}


async def ticker_task():
    # Initialize prices.
    for t in TICKERS:
        ticker_prices[t["symbol"]] = t["base"]

    while True:
        try:
            await asyncio.sleep(0.5)
            n = random.randint(200, 400)
            selected = random.sample(TICKERS, min(n, len(TICKERS)))
            for t in selected:
                sym = t["symbol"]
                price = ticker_prices[sym]
                change = price * random.uniform(-0.05, 0.05)
                price = round(max(1.0, price + change), 2)
                ticker_prices[sym] = price
                spread = round(price * 0.001, 2)
                data = {
                    "bid": price,
                    "ask": round(price + spread, 2),
                    "time": int(time.time() * 1000),
                }
                await centrifugo.map_publish(
                    "tickers:all", sym, data,
                    tags={"sector": t["sector"]},
                )
        except Exception:
            logger.exception("ticker_task error")
            await asyncio.sleep(1)


# ===================================================================
# Background task: Scoreboard
# ===================================================================
MATCHES = [
    {"id": "match_1", "home": "Arsenal", "away": "Chelsea"},
    {"id": "match_2", "home": "Barcelona", "away": "Real Madrid"},
    {"id": "match_3", "home": "Bayern", "away": "Dortmund"},
    {"id": "match_4", "home": "Liverpool", "away": "Man City"},
    {"id": "match_5", "home": "PSG", "away": "Marseille"},
    {"id": "match_6", "home": "Juventus", "away": "AC Milan"},
]

PLAYERS = {
    "Arsenal": ["Saka", "Odegaard", "Rice", "Saliba", "Havertz"],
    "Chelsea": ["Palmer", "Jackson", "Caicedo", "Colwill", "Nkunku"],
    "Barcelona": ["Yamal", "Pedri", "Gavi", "De Jong", "Lewandowski"],
    "Real Madrid": ["Bellingham", "Vinicius", "Rodrygo", "Valverde", "Modric"],
    "Bayern": ["Musiala", "Sane", "Kimmich", "Muller", "Kane"],
    "Dortmund": ["Brandt", "Adeyemi", "Sabitzer", "Hummels", "Reus"],
    "Liverpool": ["Salah", "Nunez", "Szoboszlai", "Mac Allister", "Van Dijk"],
    "Man City": ["Haaland", "De Bruyne", "Foden", "Rodri", "Grealish"],
    "PSG": ["Mbappe", "Dembele", "Vitinha", "Hakimi", "Marquinhos"],
    "Marseille": ["Aubameyang", "Sanchez", "Guendouzi", "Clauss", "Balerdi"],
    "Juventus": ["Vlahovic", "Chiesa", "Locatelli", "Bremer", "Rabiot"],
    "AC Milan": ["Leao", "Giroud", "Pulisic", "Reijnders", "Theo"],
}


def _new_match_state(m: dict) -> dict:
    return {
        "id": m["id"],
        "home_team": m["home"],
        "away_team": m["away"],
        "home_score": 0,
        "away_score": 0,
        "status": "1H",
        "minute": 0,
        "home_poss": 50,
        "away_poss": 50,
        "home_shots": 0,
        "away_shots": 0,
        "home_passes": 0,
        "away_passes": 0,
        "home_corners": 0,
        "away_corners": 0,
        "home_fouls": 0,
        "away_fouls": 0,
        "home_yellow": 0,
        "away_yellow": 0,
        "home_red": 0,
        "away_red": 0,
        "events": [],
    }


def _simulate_tick(state: dict, m: dict):
    """Advance match by ~2 minutes of play."""
    status = state["status"]
    minute = state["minute"]

    if status == "HT" or status == "FT":
        return

    minute += 2
    state["minute"] = minute

    # Transition.
    if status == "1H" and minute >= 45:
        state["status"] = "HT"
        state["minute"] = 45
        return
    if status == "2H" and minute >= 90:
        state["status"] = "FT"
        state["minute"] = 90
        return

    # Simulate events.
    for side in ["home", "away"]:
        team = m[side]
        players = PLAYERS.get(team, ["Player"])

        # Passes (3-8 per tick).
        state[f"{side}_passes"] += random.randint(3, 8)

        # Possession shift.
        shift = random.randint(-3, 3)
        state["home_poss"] = max(30, min(70, state["home_poss"] + shift))
        state["away_poss"] = 100 - state["home_poss"]

        # Shot (~20% chance).
        if random.random() < 0.20:
            state[f"{side}_shots"] += 1

            # Goal (~15% of shots).
            if random.random() < 0.15:
                state[f"{side}_score"] += 1
                player = random.choice(players)
                _add_event(state, "goal", side, player, minute)

        # Corner (~10% chance).
        if random.random() < 0.10:
            state[f"{side}_corners"] += 1
            _add_event(state, "corner", side, random.choice(players), minute)

        # Foul (~8% chance).
        if random.random() < 0.08:
            state[f"{side}_fouls"] += 1
            player = random.choice(players)

            # Yellow card (~30% of fouls).
            if random.random() < 0.30:
                state[f"{side}_yellow"] += 1
                _add_event(state, "yellow", side, player, minute)
            # Red card (~3% of fouls).
            elif random.random() < 0.03:
                state[f"{side}_red"] += 1
                _add_event(state, "red", side, player, minute)


def _add_event(state: dict, event_type: str, team: str, player: str, minute: int):
    state["events"].append({
        "type": event_type,
        "team": team,
        "player": player,
        "minute": minute,
    })
    # Keep last 8 events.
    if len(state["events"]) > 8:
        state["events"] = state["events"][-8:]


async def scoreboard_task():
    match_states = {}
    for m in MATCHES:
        match_states[m["id"]] = _new_match_state(m)

    # Stagger start times: each match starts 2 ticks after the previous.
    start_offsets = {m["id"]: i * 2 for i, m in enumerate(MATCHES)}
    tick = 0

    while True:
        try:
            await asyncio.sleep(1)
            tick += 1

            for m in MATCHES:
                mid = m["id"]
                state = match_states[mid]

                if tick < start_offsets[mid]:
                    continue

                status = state["status"]

                if status == "HT":
                    # Pause at half time for 5 seconds.
                    if not state.get("_ht_start"):
                        state["_ht_start"] = tick
                    elif tick - state["_ht_start"] >= 5:
                        state["status"] = "2H"
                        del state["_ht_start"]
                    continue

                if status == "FT":
                    # Pause at full time for 10 seconds, then reset.
                    if not state.get("_ft_start"):
                        state["_ft_start"] = tick
                    elif tick - state["_ft_start"] >= 10:
                        match_states[mid] = _new_match_state(m)
                        state = match_states[mid]
                    else:
                        continue

                _simulate_tick(state, m)

                # Publish with delta compression.
                publish_data = {k: v for k, v in state.items() if not k.startswith("_")}
                await centrifugo.map_publish(
                    "scoreboard:main", mid, publish_data, delta=True,
                )

        except Exception:
            logger.exception("scoreboard_task error")
            await asyncio.sleep(1)


# ===================================================================
# Background task: Poll manager
# ===================================================================
POLL_QUESTIONS = [
    {
        "question": "What's your favorite programming language?",
        "options": [
            {"label": "Python", "color": "#3776ab"},
            {"label": "JavaScript", "color": "#f7df1e"},
            {"label": "Go", "color": "#00add8"},
            {"label": "Rust", "color": "#dea584"},
        ],
    },
    {
        "question": "Which code editor do you prefer?",
        "options": [
            {"label": "VS Code", "color": "#007acc"},
            {"label": "Neovim", "color": "#57a143"},
            {"label": "JetBrains", "color": "#fe315d"},
            {"label": "Zed", "color": "#084ccf"},
        ],
    },
    {
        "question": "Favorite database?",
        "options": [
            {"label": "PostgreSQL", "color": "#336791"},
            {"label": "Redis", "color": "#dc382d"},
            {"label": "MongoDB", "color": "#47a248"},
            {"label": "SQLite", "color": "#003b57"},
        ],
    },
    {
        "question": "How do you deploy?",
        "options": [
            {"label": "Docker", "color": "#2496ed"},
            {"label": "Kubernetes", "color": "#326ce5"},
            {"label": "Bare Metal", "color": "#888"},
            {"label": "Serverless", "color": "#ff9900"},
        ],
    },
    {
        "question": "Tabs or spaces?",
        "options": [
            {"label": "Tabs", "color": "#e74c3c"},
            {"label": "Spaces", "color": "#3498db"},
            {"label": "Both", "color": "#9b59b6"},
            {"label": "Whatever IDE does", "color": "#95a5a6"},
        ],
    },
    {
        "question": "When do you code best?",
        "options": [
            {"label": "Morning", "color": "#f39c12"},
            {"label": "Afternoon", "color": "#e67e22"},
            {"label": "Night", "color": "#2c3e50"},
            {"label": "All day", "color": "#1abc9c"},
        ],
    },
]


async def poll_manager_task():
    poll_index = 0

    while True:
        try:
            poll = POLL_QUESTIONS[poll_index % len(POLL_QUESTIONS)]
            poll_index += 1
            poll_id = f"poll_{uuid.uuid4().hex[:8]}"

            now_ms = int(time.time() * 1000)
            duration_ms = random.randint(20000, 25000)
            end_time = now_ms + duration_ms

            # Build options with IDs.
            options = []
            for i, opt in enumerate(poll["options"]):
                opt_id = f"{poll_id}_opt_{i}"
                options.append({"id": opt_id, "label": opt["label"], "color": opt["color"]})

            # Publish poll metadata.
            meta = {
                "pollId": poll_id,
                "question": poll["question"],
                "status": "active",
                "options": options,
                "startTime": now_ms,
                "endTime": end_time,
            }
            await pg_map_publish("poll:meta", poll_id, meta)

            # Publish initial option entries.
            for opt in options:
                await pg_map_publish(
                    "poll:results", opt["id"],
                    {"optionId": opt["id"], "label": opt["label"], "color": opt["color"]},
                )

            logger.info("Poll started: %s — %s", poll_id, poll["question"])

            # Bot voting phase.
            elapsed = 0
            while elapsed < duration_ms / 1000:
                wait = random.uniform(2, 4)
                await asyncio.sleep(wait)
                elapsed += wait

                if elapsed >= duration_ms / 1000:
                    break

                # Bot votes for a random option.
                opt = random.choice(options)
                try:
                    await _record_poll_vote(opt["id"], None)
                except Exception:
                    pass

            # Close poll.
            meta["status"] = "closed"
            await pg_map_publish("poll:meta", poll_id, meta)
            logger.info("Poll closed: %s", poll_id)

            await asyncio.sleep(5)

            # Cleanup: remove all entries.
            for opt in options:
                try:
                    await pg_map_remove("poll:results", opt["id"])
                except Exception:
                    pass

            # Remove vote dedup entries.
            rows = await pool.fetch(
                "SELECT key FROM cf_map_state WHERE channel = 'poll:votes' AND key LIKE $1",
                f"{poll_id}:%",
            )
            for row in rows:
                try:
                    await pg_map_remove("poll:votes", row["key"])
                except Exception:
                    pass

            # Remove poll metadata.
            try:
                await pg_map_remove("poll:meta", poll_id)
            except Exception:
                pass

            await asyncio.sleep(2)

        except Exception:
            logger.exception("poll_manager_task error")
            await asyncio.sleep(5)


# ===================================================================
# Startup / Shutdown
# ===================================================================
@app.on_event("startup")
async def startup():
    global pool, http_client, centrifugo

    # Wait for PostgreSQL to be ready.
    for attempt in range(30):
        try:
            pool = await asyncpg.create_pool(DATABASE_URL, min_size=2, max_size=10)
            break
        except Exception:
            logger.info("Waiting for PostgreSQL... (attempt %d)", attempt + 1)
            await asyncio.sleep(2)
    else:
        logger.error("Could not connect to PostgreSQL")
        raise RuntimeError("PostgreSQL unavailable")

    logger.info("Connected to PostgreSQL")

    http_client = httpx.AsyncClient(timeout=10.0)
    centrifugo = CentrifugoAPI(http_client, CENTRIFUGO_API_URL, CENTRIFUGO_API_KEY)

    # Wait for Centrifugo to be ready (it creates the PG schema).
    for attempt in range(30):
        try:
            await centrifugo._call("info", {})
            break
        except Exception:
            logger.info("Waiting for Centrifugo... (attempt %d)", attempt + 1)
            await asyncio.sleep(2)
    else:
        logger.warning("Centrifugo not reachable — starting anyway")

    # Clean up stale poll data from previous runs.
    for ch in ("poll:meta", "poll:results", "poll:votes"):
        try:
            await centrifugo.map_clear(ch)
        except Exception:
            pass

    # Initialize inventory.
    try:
        await _init_inventory()
    except Exception:
        logger.exception("Failed to init inventory (Centrifugo may not be ready)")

    # Start background tasks.
    asyncio.create_task(ticker_task())
    asyncio.create_task(scoreboard_task())
    asyncio.create_task(poll_manager_task())
    logger.info("Background tasks started")


@app.on_event("shutdown")
async def shutdown():
    if http_client:
        await http_client.aclose()
    if pool:
        await pool.close()
