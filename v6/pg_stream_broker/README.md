# PG Stream Broker + getState Demo

Kitchen orders dashboard where **app state lives in the app's own database**
and **real-time notifications piggy-back on the same transaction** as the
state write. Centrifugo uses its PostgreSQL stream broker; the app uses
Centrifugo's built-in `cf_stream_publish` SQL function inside its own
transaction, so state mutation and publication commit atomically.

## The pattern

```python
async with conn.transaction():
    await conn.execute("INSERT INTO orders ...", ...)
    await pg_stream_publish(conn, "orders:kitchen", {...})   # same tx
```

If the INSERT fails, nothing is published. If the INSERT commits, the
publication is guaranteed to be delivered — no outbox worker, no 2PC,
no race between the DB row and the notification.

## Client side

The browser uses a regular **stream subscription** with a position-only
`getState` callback:

```js
const sub = centrifuge.newSubscription('orders:kitchen', {
    getState: async () => {
        const { entries, offset, epoch } = await fetch('/api/orders/state')
            .then(r => r.json());
        renderInitial(entries);      // app DB is the source of truth
        return { offset, epoch };    // return just the stream position
    }
});
sub.on('publication', ctx => applyDelta(ctx.data));
sub.subscribe();
```

`getState` runs when the subscription starts and again after a failed
recovery. It reads the stream top position FIRST (inside a repeatable-read
transaction), then the app's rows — this guarantees the returned position
is a lower bound, so any publications committed after the read arrive as
`publication` events via stream catch-up.

## Run

```bash
docker compose up
```

Then open <http://localhost:9001>.

## Prerequisites

- Docker Compose.

## Why this layout

- **App DB is the source of truth.** Queries use your native schema with
  joins, indexes, and RLS. The broker doesn't duplicate state.
- **Transactional consistency.** One commit, one truth — no split-brain
  between DB and notification channel.
- **Stream catch-up covers gaps.** Between `getState`'s position read and
  the subscribe, any committed publications are delivered in order.
- **No map broker needed.** Stream broker + getState gives you the full
  "real-time replication of app state" story without a duplicate state
  table in the broker.

See also: `examples/v6/map_demo` for a collection-shaped use case where
the broker itself stores the key-value state.
