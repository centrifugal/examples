from fastapi import FastAPI
import httpx
import os
import jwt
import time
import asyncio
import json

app = FastAPI()

CENTRIFUGO_HTTP_API_URL = "http://centrifugo:8000/api"
CENTRIFUGO_HTTP_API_KEY = "secret"
CENTRIFUGO_TOKEN_SECRET = "secret"
USER_ID = "user123"

def generate_centrifugo_token():
    now = int(time.time())
    exp = now + 3600
    
    payload = {
        "sub": USER_ID,
        "iat": now,
        "exp": exp,
        "channels": [USER_ID]
    }
    
    token = jwt.encode(payload, CENTRIFUGO_TOKEN_SECRET, algorithm="HS256")
    return token

@app.post("/api/token")
async def api_execute():
    token = generate_centrifugo_token()
    return {
        'token': token
    }

async def publish_to_centrifugo(channel: str, data: dict):
    headers = {
        "Authorization": f"apikey {CENTRIFUGO_HTTP_API_KEY}",
        "Content-Type": "application/json"
    }
    
    payload = {
        "method": "publish",
        "params": {
            "channel": channel,
            "data": data
        }
    }
    
    async with httpx.AsyncClient() as client:
        try:
            response = await client.post(CENTRIFUGO_HTTP_API_URL, json=payload, headers=headers)
            response.raise_for_status()
        except Exception as e:
            print(f"Error publishing to Centrifugo: {e}")

async def send_time_periodically():
    while True:
        current_time = int(time.time())
        await publish_to_centrifugo(USER_ID, {"unix_time": current_time})
        await asyncio.sleep(1)

@app.on_event("startup")
async def startup_event():
    asyncio.create_task(send_time_periodically())
