# Building a Real-time Leaderboard with Centrifugo, Redis, and React

In this tutorial, we'll build a real-time leaderboard application that updates dynamically as scores change. We'll use Centrifugo for real-time updates, Redis for data storage, and React for the frontend. This is a perfect example of how Centrifugo can be used to create interactive, real-time applications with minimal effort.

![Real-time Leaderboard Demo](https://centrifugal.dev/img/leaderboard_demo.gif)

## What We're Building

Our application will:

1. Simulate score updates for a set of players
2. Store the leaderboard data in Redis
3. Use Centrifugo to push real-time updates to connected clients
4. Display the leaderboard with smooth animations when rankings change

## Prerequisites

To follow this tutorial, you'll need:

- Docker and Docker Compose
- Basic knowledge of Python, JavaScript, and React
- Familiarity with Redis concepts
- Understanding of WebSockets and real-time communication

## Project Structure

Here's the structure of our project:

```
leaderboard/
├── backend/
│   ├── app.py                 # Python backend service
│   ├── Dockerfile             # Backend Docker configuration
│   ├── requirements.txt       # Python dependencies
│   └── lua/
│       └── update_leaderboard.lua  # Redis Lua script
├── centrifugo/
│   └── config.json            # Centrifugo configuration
├── nginx/
│   └── nginx.conf             # Nginx configuration
├── web/
│   ├── public/                # React public assets
│   ├── src/                   # React source code
│   ├── Dockerfile             # Frontend Docker configuration
│   └── package.json           # Frontend dependencies
└── docker-compose.yml         # Docker Compose configuration
```

## Step 1: Setting Up the Project

First, let's create our project structure:

```bash
mkdir -p leaderboard/{backend/lua,centrifugo,nginx,web}
cd leaderboard
```

## Step 2: Setting Up Redis and Centrifugo with Docker Compose

Create a `docker-compose.yml` file:

```yaml
services:
  redis:
    image: redis:7
    ports:
      - "6379:6379"

  centrifugo:
    image: centrifugo/centrifugo
    volumes:
      - ./centrifugo/config.json:/centrifugo/config.json
    command: centrifugo --config=/centrifugo/config.json
    ports:
      - "8000:8000"

  backend:
    build: ./backend
    depends_on:
      - redis
      - centrifugo

  nginx:
    image: nginx:alpine
    restart: always
    ports:
      - 8080:80
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/conf.d/default.conf

  web:
    build: ./web
    volumes:
      - ./web:/app
      - /app/node_modules
    ports:
      - "3000:3000"
    command: sh -c "npm install && npm start"
    depends_on:
      - centrifugo
```

## Step 3: Configuring Centrifugo

Create a `centrifugo/config.json` file:

```json
{
  "client": {
    "insecure": true,
    "allowed_origins": ["*"]
  },
  "consumers": [
    {
      "enabled": true,
      "name": "leaderboard_redis",
      "type": "redis_stream",
      "redis_stream": {
        "address": "redis://redis:6379",
        "streams": ["leaderboard-stream"],
        "consumer_group": "centrifugo",
        "num_workers": 8
      }
    }
  ]
}
```

This configuration:
- Enables insecure WebSocket connections for development
- Sets up a Redis Stream consumer to read from the "leaderboard-stream" stream
- Configures Centrifugo to connect to our Redis instance

## Step 4: Building the Backend Service

### 1. Create the Python requirements file

Create `backend/requirements.txt`:

```
redis==5.2.1
```

### 2. Create the Lua script for updating the leaderboard

Create `backend/lua/update_leaderboard.lua`:

```lua
-- Get or create state hash containing both epoch and version
local leaderboard_key = KEYS[1]
local state_key = KEYS[2]
local stream_key = KEYS[3]

local name = ARGV[1]
local score_inc = tonumber(ARGV[2])
local channel = ARGV[3]

-- Increment leaderboard score
redis.call('ZINCRBY', leaderboard_key, score_inc, name)

-- Get leaderboard data
local members = redis.call('ZREVRANGE', leaderboard_key, 0, -1, 'WITHSCORES')

local epoch = redis.call("HGET", state_key, "epoch")
if not epoch then
    local t = redis.call("TIME")
    epoch = tostring(t[1])
    redis.call("HSET", state_key, "epoch", epoch, "version", 0)
end
-- Always update TTL regardless of whether state is new or existing
redis.call("EXPIRE", state_key, 86400) -- Set TTL (24 hours, adjust as needed)

-- Increment version atomically using HINCRBY
local version = redis.call("HINCRBY", state_key, "version", 1)

local leaders = {}
for i = 1, #members, 2 do
    table.insert(leaders, { name = members[i], score = tonumber(members[i+1]) })
end

-- Prepare payload for Centrifugo publish API command.
local publish_payload = {
  channel = channel,
  data = { leaders = leaders },
  version = version, -- a tip for Centrifugo about state version
  version_epoch = epoch, -- a tip for Centrifugo about state epoch
}

-- Add to stream which is consumed by Centrifugo.
local payload = cjson.encode(publish_payload)
redis.call('XADD', stream_key, 'MAXLEN', 1, '*', 'method', 'publish', 'payload', payload)
return members
```

This Lua script:
1. Increments a player's score in a Redis sorted set
2. Retrieves the current leaderboard data
3. Manages state with epoch and version for tracking changes
4. Formats the data for Centrifugo
5. Adds the data to a Redis stream that Centrifugo consumes

### 3. Create the Python backend service

Create `backend/app.py`:

```python
import time
import random
import redis


def main():
    r = redis.Redis(host='redis', port=6379, decode_responses=True)

    with open('lua/update_leaderboard.lua', 'r') as f:
        lua_script = f.read()

    update_leaderboard = r.register_script(lua_script)

    leader_names = [
        "Alice", "Bob", "Charlie", "David", "Eve",
    ]

    while True:
        leader = random.choice(leader_names)
        increment = random.randint(1, 10)
        channel = "leaderboard"
        update_leaderboard(
            keys=["leaderboard", "leaderboard-state", "leaderboard-stream"],
            args=[leader, increment, channel]
        )
        time.sleep(0.2)


if __name__ == "__main__":
    main()
```

This Python script:
1. Connects to Redis
2. Loads the Lua script for updating the leaderboard
3. Randomly selects a player and increments their score
4. Calls the Lua script to update the leaderboard in Redis
5. Repeats this process every 0.2 seconds

### 4. Create the Dockerfile for the backend

Create `backend/Dockerfile`:

```dockerfile
FROM python:3.9-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["python", "app.py"]
```

## Step 5: Creating the Frontend React Application

### 1. Initialize a new React application

```bash
npx create-react-app web
```

### 2. Install the required dependencies

Navigate to the web directory and install the dependencies:

```bash
cd web
npm install centrifuge motion bootstrap
```

### 3. Create the React application

Update `web/src/App.js`:

```jsx
import React, { useState, useEffect } from 'react';
import { motion } from 'motion/react';
import { Centrifuge } from 'centrifuge';
import 'bootstrap/dist/css/bootstrap.min.css';
import './App.css';

function App() {
  const [state, setState] = useState({
    leaders: [],
    prevOrder: {},
    highlights: {},
  });

  useEffect(() => {
    const centrifuge = new Centrifuge("ws://localhost:8000/connection/websocket");
    const sub = centrifuge.newSubscription("leaderboard", {
      delta: 'fossil',
      since: {}
    });

    sub.on('publication', (message) => {
      const data = message.data;

      setState(prevState => {
        const newHighlights = {};
        const newLeaders = data.leaders.map((leader, index) => {
          let highlightClass = "";
          const prevRank = prevState.prevOrder[leader.name];
          if (prevRank !== undefined) {
            if (prevRank > index) {
              highlightClass = "highlight-up";
            } else if (prevRank < index) {
              highlightClass = "highlight-down";
            }
          }
          if (highlightClass) {
            newHighlights[leader.name] = highlightClass;
          }
          return leader;
        });

        const newOrder = {};
        newLeaders.forEach((leader, index) => {
          newOrder[leader.name] = index;
        });

        return {
          ...prevState,
          leaders: newLeaders,
          prevOrder: newOrder,
          highlights: { ...prevState.highlights, ...newHighlights },
        };
      });
    });

    centrifuge.connect();
    sub.subscribe();

    return () => {
      sub.unsubscribe();
      centrifuge.disconnect();
    };
  }, []);

  return (
    <div className="container mt-5">
      <div className="card">
        <div className="card-header">Real-time Leaderboard with Centrifugo</div>
        <div className="card-body">
          <table className="table table-striped">
            <thead>
              <tr>
                <th scope="col" className="rank-col">Rank</th>
                <th scope="col">Name</th>
                <th scope="col">Score</th>
              </tr>
            </thead>
            <tbody>
              {state.leaders.map((leader, index) => (
                <motion.tr
                  key={leader.name}
                  layout
                  className={state.highlights[leader.name] || ''}
                >
                  <td className="rank-col">{index + 1}</td>
                  <td>{leader.name}</td>
                  <td>{leader.score}</td>
                </motion.tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

export default App;
```

This React component:
1. Connects to Centrifugo via WebSocket
2. Subscribes to the "leaderboard" channel
3. Updates the UI when new leaderboard data is received
4. Uses motion animations to highlight changes in rankings

### 4. Add CSS styling

Update `web/src/App.css`:

```css
@import url('https://fonts.googleapis.com/css2?family=Play:wght@400;700&display=swap');

/* General Body Styling */
body {
  background-color: #f8f9fa;
  font-family: 'Play', 'Roboto', sans-serif;
  font-size: 1.2rem;
}

/* Container adjustments */
.container {
  max-width: 800px;
}

/* Card styling improvements */
.card {
  border: none;
  border-radius: 0.5rem;
  box-shadow: 0 4px 8px rgba(0,0,0,0.1);
}

.card-header {
  background: linear-gradient(45deg, #1fb8ff, #ff68cd);
  color: #fff;
  font-size: 1.5rem;
  text-align: center;
  padding: 1rem;
  border-top-left-radius: 15px;
  border-top-right-radius: 15px;
}

/* Table enhancements */
.table thead th {
  background-color: #e9ecef;
}

.table tbody tr:hover {
  background-color: #f1f1f1;
}

/* Fixed width for the rank column to prevent jumping */
.rank-col {
  min-width: 60px;
  text-align: center;
  font-variant-numeric: tabular-nums;
}

/* Highlight animations */
.highlight-up td {
  animation: highlightUp 1s ease-in-out;
}
.highlight-down td {
  animation: highlightDown 1s ease-in-out;
}
@keyframes highlightUp {
  from { background-color: #a4ff9f; }
  to { background-color: transparent; }
}
@keyframes highlightDown {
  from { background-color: #feabb2; }
  to { background-color: transparent; }
}
```

### 5. Create the Dockerfile for the frontend

Create `web/Dockerfile`:

```dockerfile
FROM node:16-alpine

WORKDIR /app

COPY package.json package-lock.json ./

RUN npm install

COPY . .

EXPOSE 3000

CMD ["npm", "start"]
```

## Step 6: Configuring Nginx

Create `nginx/nginx.conf`:

```nginx
server {
    listen 80;
    server_name localhost;

    location / {
        proxy_pass http://web:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

This Nginx configuration:
1. Listens on port 80
2. Proxies requests to the React application running on port 3000
3. Properly handles WebSocket connections

## Step 7: Running the Application

Now that we have all the components set up, let's run the application:

```bash
docker-compose up
```

This will:
1. Start Redis
2. Start Centrifugo
3. Build and start the backend service
4. Build and start the React frontend
5. Start Nginx

Once everything is running, you can access the application at http://localhost:8080.

## How It Works

Let's understand the data flow in our application:

1. **Backend Service**:
   - Randomly selects a player and increments their score
   - Uses a Lua script to update the leaderboard in Redis
   - Publishes the updated leaderboard to a Redis stream

2. **Centrifugo**:
   - Consumes from the Redis stream
   - Pushes updates to connected clients via WebSockets

3. **Frontend**:
   - Connects to Centrifugo via WebSocket
   - Subscribes to the leaderboard channel
   - Updates the UI in real-time with animations when rankings change

The key to this real-time functionality is Centrifugo's ability to consume from Redis streams and push updates to connected clients. This architecture allows for efficient, scalable real-time updates without the need for complex custom WebSocket implementations.

## Optimizations and Enhancements

Here are some ways you could enhance this application:

1. **Authentication**: Add user authentication to allow players to log in and see their own scores
2. **Custom Scores**: Allow users to submit their own scores instead of using random increments
3. **Multiple Leaderboards**: Support different categories or time periods (daily, weekly, all-time)
4. **Persistence**: Add a database for long-term storage of leaderboard data
5. **Admin Panel**: Create an admin interface for managing leaderboards and players

## Conclusion

In this tutorial, we've built a real-time leaderboard application using Centrifugo, Redis, and React. This demonstrates how Centrifugo can be used to create interactive, real-time applications with minimal effort.

Centrifugo's ability to consume from Redis streams and push updates to connected clients makes it an excellent choice for real-time applications like leaderboards, chat applications, notifications, and more.

By leveraging Centrifugo's real-time capabilities, you can create engaging, interactive applications that provide users with immediate feedback and updates.

## Design Considerations and Benefits

The architecture we've implemented in this tutorial offers several significant advantages for real-time applications:

### 1. Scalability

- **Horizontal Scaling**: Each component (Redis, Centrifugo, backend, frontend) can be scaled independently based on load requirements.
- **Stateless Backend**: The Python backend doesn't maintain connection state with clients, making it easy to scale out.
- **Redis Streams**: Redis streams provide a durable, append-only log that can handle high throughput of events.
- **Centrifugo Clustering**: In production, Centrifugo can be deployed in a cluster configuration for handling millions of simultaneous WebSocket connections.

### 2. Loose Coupling

- **Service Independence**: Each service has a clearly defined responsibility and can be developed, deployed, and scaled independently.
- **Communication via Redis**: Using Redis as the communication backbone means services don't need direct knowledge of each other.
- **Event-Driven Architecture**: The system operates on an event-driven model where changes propagate through the system asynchronously.

### 3. Performance

- **Efficient Data Transfer**: Only the necessary data (leaderboard changes) is sent over the network, minimizing bandwidth usage.
- **Optimized Frontend Updates**: The React application only re-renders the components that have changed, not the entire leaderboard.
- **Lua Scripts**: Using Redis Lua scripts allows for atomic operations and reduces network round-trips between the application and Redis.
- **WebSocket Efficiency**: WebSockets provide a persistent connection that's more efficient than polling for real-time updates.

### 4. Developer Experience

- **Simplified Real-time Implementation**: Centrifugo handles the complex aspects of WebSocket connections, including reconnection logic, heartbeats, and message delivery guarantees.
- **Clear Separation of Concerns**: The architecture clearly separates data storage (Redis), real-time communication (Centrifugo), business logic (backend), and presentation (frontend).
- **Reduced Boilerplate**: No need to implement custom WebSocket servers or complex client-side connection management.

### 5. Reliability

- **Message Durability**: Redis streams ensure that messages aren't lost, even if Centrifugo is temporarily unavailable.
- **Automatic Reconnection**: The Centrifuge client library handles reconnection automatically if the connection is lost.
- **State Versioning**: The implementation includes state versioning (epoch and version) to ensure clients have the most up-to-date information.

### 6. Production Readiness

- **Containerization**: The entire application is containerized, making it easy to deploy to any environment that supports Docker.
- **Nginx as Reverse Proxy**: Using Nginx provides additional capabilities like SSL termination, load balancing, and request filtering.
- **Environment Isolation**: Docker Compose ensures consistent environments across development and production.

## Centrifugo's Unique Advantages

Compared to other real-time technologies, Centrifugo offers several unique advantages that make it particularly well-suited for applications like our leaderboard:

### 1. Language-Agnostic Architecture

- **Standalone Server**: Unlike Socket.IO (Node.js), SignalR (.NET), or Pusher, Centrifugo is a standalone server that can work with any backend language.
- **Multiple Client SDKs**: Official client libraries are available for JavaScript, iOS, Android, and more, allowing for consistent real-time experiences across platforms.
- **API-First Design**: Centrifugo's HTTP API and Redis integration make it easy to integrate with any existing application stack.

### 2. Advanced State Management

- **Built-in State Synchronization**: Centrifugo's "fossil delta" feature allows efficient synchronization of state between server and clients, sending only the changes rather than the entire state.
- **State Recovery**: Clients can recover missed messages after disconnection, ensuring data consistency.
- **Epoch and Version Control**: The built-in versioning system helps prevent conflicts and ensures clients have the most up-to-date information.

### 3. Redis Integration

- **Native Redis Streams Support**: Centrifugo can directly consume from Redis streams without additional middleware, simplifying the architecture.
- **Redis Engine Option**: Beyond streams, Centrifugo can use Redis as a PUB/SUB engine for horizontal scaling.
- **Redis Adapters**: Built-in adapters for Redis presence and history features.

### 4. Performance and Scalability

- **Written in Go**: Centrifugo is built with Go, providing excellent performance and concurrency handling.
- **Minimal Resource Usage**: Efficient memory and CPU utilization compared to Node.js or Python-based alternatives.
- **Connection Multiplexing**: A single Centrifugo server can handle tens of thousands of concurrent WebSocket connections.
- **Horizontal Scaling**: Native support for clustering without sticky sessions, unlike many alternatives.

### 5. Security Features

- **Token-Based Authentication**: JWT-based authentication system that's more secure than many alternatives.
- **Channel Namespaces**: Fine-grained access control with different rules for different channel types.
- **Private Channels**: Support for private channels with dynamic authorization.
- **Connection Expiration**: Automatic connection expiration to prevent unauthorized access.

### 6. Protocol Flexibility

- **Multiple Transport Options**: WebSocket, SockJS, HTTP streaming, and long-polling support, with automatic fallback.
- **GRPC API**: In addition to HTTP API, Centrifugo offers a GRPC API for high-performance server-to-server communication.
- **Bidirectional Communication**: Unlike some pub/sub systems, Centrifugo supports client-to-server commands in addition to server-to-client messages.

### 7. Operational Advantages

- **Metrics and Monitoring**: Built-in Prometheus metrics for monitoring and alerting.
- **Admin Web Interface**: Real-time monitoring and debugging through the admin UI.
- **Detailed Logging**: Configurable logging levels for troubleshooting.
- **Zero Dependencies**: Single binary deployment with no external dependencies (except Redis if used).

These unique advantages make Centrifugo an excellent choice for real-time applications where reliability, scalability, and developer experience are priorities. While other technologies might excel in specific use cases, Centrifugo's combination of features makes it particularly well-suited for applications like our leaderboard that require efficient state synchronization and scalable real-time updates.

This architecture pattern can be extended to many other real-time use cases beyond leaderboards, such as:

- Chat applications
- Collaborative editing tools
- Live dashboards and analytics
- Notification systems
- Multiplayer games
- Live auction or bidding systems

By using Centrifugo with Redis streams, you get a robust, scalable real-time system without the complexity of building and maintaining custom WebSocket infrastructure.
