# Centrifugo PRO — Native Web Push (VAPID) example

A self-contained demo of native Web Push with Centrifugo PRO: register a browser for push,
then send a notification to it. Native Web Push is FCM-free and works in Chrome, Edge,
Firefox, and Safari (16.4+).

## Layout

```
docker-compose.yml   # centrifugo + postgres + redis + backend + nginx
centrifugo.json      # Centrifugo PRO config (webpush provider enabled)
backend/             # tiny stdlib Go service proxying to the Centrifugo HTTP API
nginx/               # serves the frontend + proxies /api/ to the backend
static/              # frontend (index.html, app.js, service worker)
```

This example requires **Centrifugo PRO** — native Web Push is a PRO feature (docs:
<https://centrifugal.dev/docs/pro/push_notifications>). Everything runs in Docker — a single
`docker compose up` starts Centrifugo PRO, Postgres, Redis, the backend, and nginx.

## Steps

### 1. Generate a VAPID key pair

With Go:

```bash
cd backend && go run . -genvapid
```

Without Go — use the Node `web-push` package (needs Node.js, no install required):

```bash
npx web-push generate-vapid-keys
```

Both print a `Public Key` / `Private Key` pair in the base64url format Centrifugo expects. Copy
them into `centrifugo.json` under `push_notifications.webpush` (`vapid_public_key` /
`vapid_private_key`) and set `subject` to a real `mailto:` or `https:` URL.

### 2. Start the containers

From this directory:

```bash
docker compose up -d
```

This starts Centrifugo PRO (`:8000`), Postgres (`:5432`), Redis (`:6379`), the backend, and
nginx (frontend on `http://localhost:9001`). The backend reaches Centrifugo in-compose at
`http://centrifugo:8000`.

### 3. Open the page and test

Open <http://localhost:9001> in Chrome or Firefox.

1. Click **"1. Enable notifications"** — grants permission, subscribes, and registers the
   device in Centrifugo (you'll see a device id in the log).
2. Click **"2. Send test push"** — the backend calls `send_push_notification` targeting all
   `webpush` devices for the user; the notification should appear.

## Notes

- **Browser endpoints differ** (Chrome → `fcm.googleapis.com`, Firefox → Mozilla autopush,
  Safari → `web.push.apple.com`) but Centrifugo speaks the standard Web Push protocol to all
  of them — no per-browser configuration needed.
- **Safari** requires the web app to be installed (added to Dock / Home Screen) and Safari
  16.4+ (macOS 13+ / iOS 16.4+). Use Chrome/Firefox for the quickest test.
- **HTTP for localhost is fine** — browsers treat `http://localhost` as a secure context.
- The whole `PushSubscription` JSON is stored as the Centrifugo device **token**; expired
  subscriptions (HTTP 404/410 from the push service) are removed from storage automatically.
- Sending uses a `DeviceFilter` recipient, so timezone-aware delivery, rate limits, templating
  and localization all apply. Raw `webpush_tokens` sends are also supported but skip those
  per-device features.

## Troubleshooting

- **Logs show "Push queued" / "sending to tokens" with no error, but no notification appears.**
  The push was accepted by the browser's push service — the problem is display, not delivery.
  Most often it's OS/browser notification settings:
  - macOS: System Settings → Notifications → your browser → **Allow Notifications** ON, and turn
    off **Focus / Do Not Disturb**.
  - `chrome://settings/content/notifications` → ensure `http://localhost:9001` is **Allowed**.
  To isolate display from delivery, open DevTools → Application → Service Workers and use the
  **Push** test button (e.g. `{"title":"Test","body":"hi"}`); the SW also logs `[sw] push event
  received …` to its console when a real push arrives.
- **`InvalidCharacterError` on subscribe.** The VAPID public key in `centrifugo.json` is the
  placeholder or malformed — generate keys with `go run . -genvapid` and paste them in.
- **Changed `sw.js` but behavior is stale.** Service workers cache; in DevTools → Application →
  Service Workers enable "Update on reload" (or Unregister) and reload.
