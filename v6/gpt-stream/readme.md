# 🧠 MCP Chat with GPT Streaming

A full-stack real-time chat app powered by:

- **Centrifugo** for realtime message delivery
- **FastAPI** (async) backend for handling chat and AI requests
- **OpenAI** for LLM responses
- **Streaming token delivery** via temporary Centrifugo channels using the MCP (Message Communication Protocol)

---

## 🚀 Features

- ✅ Realtime chat using Centrifugo
- ✅ `/ask` questions sent to OpenAI and streamed token-by-token
- ✅ Frontend subscribes to a unique UUID-based stream channel per question
- ✅ Stream ends cleanly with a `done: true` signal
- ✅ Fully MCP-compliant message format

---

## 🧰 Tech Stack

- **FastAPI** – async backend with streaming support
- **Centrifugo** – pub/sub websocket server
- **OpenAI API** – LLM responses (via GPT-3.5/4)
- **httpx** – async publishing to Centrifugo
- **Docker Compose** – one command to run everything

---

## 🗂 Directory Structure

```
project-root/
├── backend/
│ ├── app.py # FastAPI backend with streaming
│ ├── Dockerfile
│ └── requirements.txt
├── frontend/
│ └── index.html # HTML UI using Centrifuge.js
├── centrifugo/
│ └── config.json # Centrifugo config
├── .env # Your API keys (gitignored)
├── docker-compose.yml
```


---

## ⚙️ Setup

1. Copy `.env.example` to `.env` and fill in:

```env
CENTRIFUGO_API_KEY=your_centrifugo_api_key
OPENAI_API_KEY=your_openai_key
```

Then:

```
docker compose up --build
```

Visit: http://localhost:3000

## 🧪 Example

Type:

```bash
/ask What is the capital of Japan?
```

Watch GPTBot stream:

```bash
GPTBot: The capital of Japan is Tokyo.
```
