package limiter

import (
	"time"

	"github.com/go-redis/redis/v8"
)

// BucketConfig describes a token bucket's shape: capacity, and how fast it
// refills. It carries no per-caller state — state (current tokens, last
// refill time) lives only in Redis, keyed per (path, identifier), so
// changing these limits takes effect immediately for every caller instead
// of being masked by whatever config an existing Redis entry was created
// with (see rateLimiter.go's configFor).
type BucketConfig struct {
	MaxTokens           int64
	RefillWindow        time.Duration
	RefillSizePerWindow int64
}

// idleTTL is how long an untouched bucket key is kept in Redis: long enough
// to cover a couple of full refill cycles, so a client that goes quiet for a
// bit doesn't get penalized by stale state on return, but short enough that
// abandoned keys don't linger indefinitely.
func (c BucketConfig) idleTTL() time.Duration {
	if c.RefillSizePerWindow <= 0 || c.RefillWindow <= 0 {
		return time.Hour
	}
	windowsToFull := (c.MaxTokens + c.RefillSizePerWindow - 1) / c.RefillSizePerWindow
	ttl := c.RefillWindow * time.Duration(windowsToFull) * 2
	if ttl < time.Minute {
		return time.Minute
	}
	return ttl
}

// bucketScript atomically reads, refills, and consumes a token from the
// bucket stored at KEYS[1], in a single Redis round trip. Running the whole
// read-refill-consume-write cycle server-side as one Lua script is what
// makes it race-free: two concurrent requests for the same identifier can't
// both read the same starting balance and both "win", the way a Go-side
// GET / mutate / SET sequence would allow.
//
// ARGV: now (unix seconds), max_tokens, refill_window_seconds,
// refill_size_per_window, ttl_seconds.
// Returns {allowed (0/1), tokens_remaining}.
var bucketScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local max_tokens = tonumber(ARGV[2])
local refill_window = tonumber(ARGV[3])
local refill_size = tonumber(ARGV[4])
local ttl = tonumber(ARGV[5])

local data = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(data[1])
local last_refill = tonumber(data[2])

if tokens == nil then
  tokens = max_tokens
  last_refill = now
end

if refill_window > 0 then
  local elapsed = now - last_refill
  if elapsed > 0 then
    local windows = math.floor(elapsed / refill_window)
    if windows > 0 then
      tokens = math.min(max_tokens, tokens + windows * refill_size)
      last_refill = last_refill + (windows * refill_window)
    end
  end
end

local allowed = 0
if tokens > 0 then
  tokens = tokens - 1
  allowed = 1
end

redis.call('HSET', key, 'tokens', tokens, 'last_refill', last_refill)
redis.call('EXPIRE', key, ttl)

return {allowed, tokens}
`)
