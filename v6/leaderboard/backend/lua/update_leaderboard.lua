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
redis.call('XADD', stream_key, 'MAXLEN', '~', 10000, '*', 'method', 'publish', 'payload', payload)
return members
