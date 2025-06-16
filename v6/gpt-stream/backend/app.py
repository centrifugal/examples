from fastapi import FastAPI
from pydantic import BaseModel
from openai import OpenAI
import httpx
import os

app = FastAPI()

client = None
if os.getenv("OPENAI_API_KEY"):
    client = OpenAI(api_key=os.getenv("OPENAI_API_KEY"))

CENTRIFUGO_HTTP_API_URL = "http://centrifugo:8000/api"
CENTRIFUGO_HTTP_API_KEY = "secret"

class Command(BaseModel):
    text: str
    channel: str


@app.post("/api/execute")
async def api_execute(cmd: Command):
    await handle_command(cmd)
    return {}


class StreamMessage(BaseModel):
    text: str
    done: bool


async def handle_command(cmd: Command):
    text = cmd.text
    channel = cmd.channel

    if not client:
        await publish_message(
            channel,
            StreamMessage(text=f"⚠️ Error: OPENAI_API_KEY env is not set", done=True).model_dump()
        )

    try:
        response = client.chat.completions.create(
            model="gpt-3.5-turbo",
            messages=[{"role": "user", "content": text}],
            stream=True,
        )
        for chunk in response:
            token = chunk.choices[0].delta.content or ""
            if token:
                await publish_message(
                    channel,
                    StreamMessage(text=token, done=False).model_dump()
                )
        await publish_message(
            channel,
            StreamMessage(text=token, done=True).model_dump()
        )
    except Exception as e:
        await publish_message(
            channel,
            StreamMessage(text=f"⚠️ Error: {e}", done=True).model_dump()
        )


async def publish_message(channel, stream_message):
    payload = {
        "channel": channel,
        "data": stream_message
    }

    headers = {
        "X-API-Key": f"{CENTRIFUGO_HTTP_API_KEY}",
        "Content-Type": "application/json"
    }

    async with httpx.AsyncClient() as http_client:
        await http_client.post(
            f"{CENTRIFUGO_HTTP_API_URL}/publish", json=payload, headers=headers)
