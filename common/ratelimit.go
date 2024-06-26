package common

import (
	"context"

	"golang.org/x/time/rate"
)

// thanks to Sourav Choudhary for this code

type DynamicRateLimiter struct {
	limiter *rate.Limiter
	rps     float64
	burst   int
	updates chan rateParams
}

type rateParams struct {
	rps   float64
	burst int
}

func NewDynamicRateLimiter(rps float64, burst int) *DynamicRateLimiter {
	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	updates := make(chan rateParams)
	go func() {
		for params := range updates {
			limiter.SetLimit(rate.Limit(params.rps))
			limiter.SetBurst(params.burst)
		}
	}()
	return &DynamicRateLimiter{
		limiter: limiter,
		rps:     rps,
		burst:   burst,
		updates: updates,
	}
}

func (drl *DynamicRateLimiter) Wait(ctx context.Context) error {
	return drl.limiter.Wait(ctx)
}

func (drl *DynamicRateLimiter) Allow() bool {
	return drl.limiter.Allow()
}

func (drl *DynamicRateLimiter) Update(rps float64, burst int) {
	drl.updates <- rateParams{rps: rps, burst: burst}
}

func (drl *DynamicRateLimiter) GetRPS() float64 {
	return drl.rps
}

func (drl *DynamicRateLimiter) GetBurst() int {
	return drl.burst
}
