package limiter

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/JohnnyKahiu/speedsales_login/pkg/database"
)

// ErrLimitReached is returned by Check when the caller has exhausted its
// token bucket. Compare against it with errors.Is — do not compare errors
// with == against a freshly constructed error value, which never matches.
var ErrLimitReached = errors.New("limit reached")

// limiters maps a route to its bucket shape. Routes not listed here fall
// back to the "/" bucket via configFor, which also guarantees Check never
// dereferences a missing map entry.
var limiters = map[string]BucketConfig{
	"/login":       {MaxTokens: 15, RefillWindow: time.Minute, RefillSizePerWindow: 3},
	"/login/reset": {MaxTokens: 15, RefillWindow: 3 * time.Minute, RefillSizePerWindow: 1},
	"/":            {MaxTokens: 10, RefillWindow: time.Second, RefillSizePerWindow: 1},
}

func configFor(path string) BucketConfig {
	if cfg, ok := limiters[path]; ok {
		return cfg
	}
	return limiters["/"]
}

// Limiter identifies one rate-limit check. TokenID is whoever is being
// limited — a username, a session token, or a client IP. Callers should
// never leave it empty: every caller sharing "" would share a single
// bucket, letting one anonymous client exhaust it for everyone else.
// Path selects which BucketConfig applies.
type Limiter struct {
	TokenID string
	Path    string
}

// Check consumes one token from the bucket identified by (Path, TokenID),
// refilling it first if enough time has passed. The full refill-and-consume
// cycle runs as a single Redis Lua script (bucketScript), so it can't race
// with a concurrent Check for the same identifier.
//
// If Redis itself is unreachable, Check fails open: it logs the error and
// allows the request through, rather than taking the whole service's login
// path down because the rate limiter's own dependency is unavailable. In
// that case the returned error is the underlying Redis error, not
// ErrLimitReached — callers should only treat `allowed == false` as
// "send a 429", not the presence of a non-nil error.
func (l *Limiter) Check(ctxt context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctxt, 5*time.Second)
	defer cancel()

	id := l.TokenID
	if id == "" {
		id = "anon"
	}
	key := fmt.Sprintf("ratelimit:%s:%s", l.Path, id)

	cfg := configFor(l.Path)
	now := float64(time.Now().UnixNano()) / float64(time.Second)

	res, err := bucketScript.Run(ctx, database.RdbCon, []string{key},
		now,
		cfg.MaxTokens,
		cfg.RefillWindow.Seconds(),
		cfg.RefillSizePerWindow,
		int64(cfg.idleTTL().Seconds()),
	).Slice()
	if err != nil {
		log.Printf("rate limiter: redis error, failing open   path=%s id=%s err=%v", l.Path, id, err)
		return true, err
	}
	if len(res) != 2 {
		return true, fmt.Errorf("rate limiter: unexpected script result %#v", res)
	}

	allowed, _ := res[0].(int64)
	if allowed == 1 {
		return true, nil
	}
	return false, ErrLimitReached
}
