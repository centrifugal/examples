# Centrifugo Delta Compression Example

This example demonstrates Centrifugo's delta compression feature using the Fossil delta algorithm. Delta compression significantly reduces bandwidth by sending only the differences between successive messages instead of full payloads.

## Features

- **Delta Compression**: Uses Fossil algorithm to compress message payloads
- **Real-time Updates**: Publisher sends incremental state updates every 2 seconds
- **Visual Feedback**: Web UI shows delta vs full messages with statistics
- **Docker Compose**: Easy setup with all services containerized

## How It Works

1. **Centrifugo Server**: Configured with `updates` namespace that has delta compression enabled
2. **Publisher Service**: Python script that periodically publishes data with small changes to the `updates:data` channel
3. **Web Client**: HTML page with centrifuge-js that subscribes with `delta: "fossil"` option

The first message received by a client is always the full payload. Subsequent messages are sent as deltas (differences from the previous message), which are automatically reconstructed by the client SDK.

## Configuration

The example uses the following Centrifugo configuration for delta compression:

```json
{
  "channel": {
    "namespaces": [
      {
        "name": "updates",
        "allowed_delta_types": ["fossil"],
        "force_positioning": true,
        "history_size": 10,
        "history_ttl": "300s",
        "delta_publish": true
      }
    ]
  }
}
```

Key settings:
- `allowed_delta_types: ["fossil"]` - Enables Fossil delta compression
- `force_positioning: true` - Required for delta compression
- `history_size: 10` - Keeps last 10 messages in history
- `delta_publish: true` - Automatically uses delta for all publications in this namespace

## Requirements

- Docker
- Docker Compose

## Running the Example

1. Start all services:
   ```bash
   docker-compose up
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:8080
   ```

3. You should see:
   - Connection status indicator
   - Latest publication data
   - Statistics showing total, delta, and full messages
   - Message log showing each publication

**Note**: The example uses nginx to serve the HTML and proxy WebSocket connections to Centrifugo. Any changes you make to `index.html` will be visible immediately on browser refresh (no Docker restart required).

4. Open browser Developer Tools (Network tab) to inspect WebSocket frames and see the size difference between full messages and deltas.

## What to Observe

- **First Message**: Will be a full publication (no delta)
- **Subsequent Messages**: Will be delta-compressed (much smaller)
- **Statistics Panel**: Shows the ratio of delta vs full messages
- **Message Log**: Indicates which messages are deltas (green border) vs full (blue border)

## How to Test Delta Compression Effectiveness

1. Open browser DevTools → Network tab
2. Filter by "WS" (WebSocket)
3. Click on the WebSocket connection
4. View the "Messages" tab
5. Compare the size of the first message (full) vs subsequent messages (deltas)

You should see that delta messages are significantly smaller than full messages, especially for the data structure used in this example where only a few fields change between updates.

## Publisher Details

The publisher script (`publisher.py`) sends updates every 2 seconds with:
- Incrementing counter
- Slightly modified metrics (CPU, memory, disk, network)
- Occasionally added events
- Rarely toggled feature flags

This pattern simulates real-world scenarios where most of the data structure remains the same, with only small portions changing - ideal for delta compression.

## Stopping the Example

Press `Ctrl+C` in the terminal where docker-compose is running, or run:

```bash
docker-compose down
```

## Files

- `config.json` - Centrifugo configuration with delta compression enabled
- `docker-compose.yml` - Docker Compose configuration with 3 services (centrifugo, web, publisher)
- `nginx.conf` - nginx configuration for serving static files and proxying WebSocket connections
- `index.html` - Web client with centrifuge-js (hot-reload enabled)
- `publisher.py` - Python script that publishes data via Centrifugo HTTP API
- `requirements.txt` - Python dependencies
- `Dockerfile.publisher` - Dockerfile for publisher service

## Learn More

- [Centrifugo Delta Compression Documentation](https://centrifugal.dev/docs/server/delta_compression)
- [Centrifugo Server API](https://centrifugal.dev/docs/server/server_api)
- [centrifuge-js Documentation](https://github.com/centrifugal/centrifuge-js)
