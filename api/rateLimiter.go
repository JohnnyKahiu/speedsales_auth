package api

import (
	"context"
	"time"
)

type Limiter struct {
	MaxTokens          int64
	RefilWindow        time.Duration
	RefilSizePerWindow int
	TokenID            string
	ResourcePath       string
	LastRefil          time.Time
}

func (arg *Limiter) allowRequest(ctx context.Context) bool {
	const aLLOW = true
	const dENY = false

	return dENY
}

func (arg *Limiter) Refil(ctx context.Context) {

}
