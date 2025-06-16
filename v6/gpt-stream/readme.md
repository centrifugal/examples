# 🧠 Chat with GPT Streaming

## 🧰 Tech Stack

- **FastAPI** – async backend with streaming support
- **Centrifugo** – pub/sub websocket server
- **OpenAI API** – LLM responses (via GPT-3.5/4)
- **httpx** – async publishing to Centrifugo
- **Docker Compose** – one command to run everything

---

## ⚙️ Setup

1. Create `.env` file in example root and fill in:

```env
CENTRIFUGO_HTTP_API_KEY="secret"
OPENAI_API_KEY="<YOUR_OPEN_AI_TOKEN>"
```

Then:

```
docker compose up --build
```

Visit: http://localhost:9000
