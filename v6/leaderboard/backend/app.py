import time
import random
import redis

# Connect to Redis (using the service name from docker-compose)
r = redis.Redis(host='redis', port=6379, decode_responses=True)

# Load the Lua script
with open('lua/update_leaderboard.lua', 'r') as f:
    lua_script = f.read()

# Register the Lua script
update_leaderboard = r.register_script(lua_script)

# List of sample leaders
leaders = ["Alice", "Bob", "Charlie", "David", "Eve"]

while True:
    leader = random.choice(leaders)
    increment = random.randint(1, 10)
    channel = "leaderboard"
    result = update_leaderboard(
        keys=["leaderboard", "leaderboard-state", "leaderboard-stream"],
        args=[leader, increment, channel]
    )
    time.sleep(0.2)
