import time
import random
import redis


def main():
    r = redis.Redis(host='redis', port=6379)

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
