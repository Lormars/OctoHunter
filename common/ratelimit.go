package common

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

type DynamicRateLimiter struct {
	limiter *rate.Limiter
	rps     float64
	burst   int
	updates chan rateParams
	mu      sync.Mutex
}

type rateParams struct {
	rps   float64
	burst int
}

func NewDynamicRateLimiter(rps float64, burst int) *DynamicRateLimiter {
	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	updates := make(chan rateParams)
	drl := &DynamicRateLimiter{
		limiter: limiter,
		rps:     rps,
		burst:   burst,
		updates: updates,
	}

	go drl.listenForUpdates()

	return drl
}

func (drl *DynamicRateLimiter) listenForUpdates() {
	for params := range drl.updates {
		drl.mu.Lock()
		drl.limiter.SetLimit(rate.Limit(params.rps))
		drl.limiter.SetBurst(params.burst)
		drl.rps = params.rps
		drl.burst = params.burst
		drl.mu.Unlock()
	}
}

func (drl *DynamicRateLimiter) Wait(ctx context.Context) error {
	return drl.limiter.Wait(ctx)
}

func (drl *DynamicRateLimiter) Allow() bool {
	return drl.limiter.Allow()
}

func (drl *DynamicRateLimiter) Update(rps float64, burst int) {
	if rps < 2 {
		rps = 2
		burst = 4
	} else if rps > 150 {
		rps = 150
		burst = 200
	}
	drl.updates <- rateParams{rps: rps, burst: burst}
}

func (drl *DynamicRateLimiter) GetRPS() float64 {
	drl.mu.Lock()
	defer drl.mu.Unlock()
	return drl.rps
}

func (drl *DynamicRateLimiter) GetBurst() int {
	drl.mu.Lock()
	defer drl.mu.Unlock()
	return drl.burst
}
