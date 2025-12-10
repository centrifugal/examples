#!/usr/bin/env python3
"""
Publisher script that periodically sends data to Centrifugo using server API.
The data has small incremental changes to demonstrate delta compression effectiveness.
"""

import os
import time
import json
import random
import requests
from datetime import datetime

# Configuration from environment variables
CENTRIFUGO_API_URL = os.getenv('CENTRIFUGO_API_URL', 'http://localhost:8000/api/publish')
CENTRIFUGO_API_KEY = os.getenv('CENTRIFUGO_API_KEY', 'my_api_key')
CHANNEL = 'updates:data'
PUBLISH_INTERVAL = 0.1  # seconds

# Sample data that will be gradually modified
state = {
    "timestamp": "",
    "counter": 0,
    "status": "active",
    "metrics": {
        "cpu": 0.0,
        "memory": 0.0,
        "disk": 0.0,
        "network": {
            "rx": 0,
            "tx": 0
        }
    },
    "events": [],
    "config": {
        "max_connections": 1000,
        "timeout": 30,
        "debug": False,
        "features": {
            "feature_a": True,
            "feature_b": False,
            "feature_c": True
        }
    }
}


def publish_to_centrifugo(data):
    """Publish data to Centrifugo channel via HTTP API."""
    headers = {
        'Content-Type': 'application/json',
        'X-API-Key': CENTRIFUGO_API_KEY
    }

    payload = {
        'channel': CHANNEL,
        'data': data
    }

    try:
        response = requests.post(CENTRIFUGO_API_URL, json=payload, headers=headers)
        response.raise_for_status()
        result = response.json()

        if 'error' in result:
            print(f"Error from Centrifugo: {result['error']}")
            return False
        else:
            print(f"Published successfully (offset: {result.get('result', {}).get('offset', 'N/A')})")
            return True
    except requests.exceptions.RequestException as e:
        print(f"Failed to publish: {e}")
        return False


def update_state():
    """Update state with small incremental changes."""
    state["timestamp"] = datetime.utcnow().isoformat() + "Z"
    state["counter"] += 1

    # Update metrics with small changes
    state["metrics"]["cpu"] = round(random.uniform(10, 90), 2)
    state["metrics"]["memory"] = round(random.uniform(30, 80), 2)
    state["metrics"]["disk"] = round(random.uniform(40, 95), 2)
    state["metrics"]["network"]["rx"] += random.randint(100, 1000)
    state["metrics"]["network"]["tx"] += random.randint(50, 800)

    # Randomly add an event (but keep list limited)
    if random.random() > 0.7:
        event = {
            "type": random.choice(["info", "warning", "error"]),
            "message": f"Event at {state['counter']}",
            "timestamp": state["timestamp"]
        }
        state["events"].insert(0, event)
        # Keep only last 3 events
        state["events"] = state["events"][:3]

    # Occasionally toggle a feature
    if random.random() > 0.9:
        feature = random.choice(["feature_a", "feature_b", "feature_c"])
        state["config"]["features"][feature] = not state["config"]["features"][feature]


def main():
    print(f"Starting publisher...")
    print(f"Publishing to: {CENTRIFUGO_API_URL}")
    print(f"Channel: {CHANNEL}")
    print(f"Interval: {PUBLISH_INTERVAL}s")
    print("-" * 60)

    # Wait a bit for Centrifugo to be ready
    time.sleep(3)

    while True:
        update_state()
        print(f"\n[{datetime.now().strftime('%H:%M:%S')}] Publishing update #{state['counter']}")
        print(f"Data size: {len(json.dumps(state))} bytes")

        success = publish_to_centrifugo(state)

        if not success:
            print("Retrying in 5 seconds...")
            time.sleep(5)
        else:
            time.sleep(PUBLISH_INTERVAL)


if __name__ == '__main__':
    main()
