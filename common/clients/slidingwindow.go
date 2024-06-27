package clients

import (
	"sync"
	"time"
)

type SlidingWindow struct {
	requests map[string][]time.Time
	mu       sync.Mutex
}

func NewSlidingWindow() *SlidingWindow {
	return &SlidingWindow{
		requests: make(map[string][]time.Time),
	}
}

func (sw *SlidingWindow) AddRequest(host string) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	sw.requests[host] = append(sw.requests[host], now)

}

func (sw *SlidingWindow) cleanup(host string) {
	now := time.Now()
	tenSecondsAgo := now.Add(-10 * time.Second)
	i := 0
	for _, t := range sw.requests[host] {
		if t.After(tenSecondsAgo) {
			break
		}
		i++
	}
	sw.requests[host] = sw.requests[host][i:]
}
func (sw *SlidingWindow) GetRequestCount(host string) int {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.cleanup(host)
	return len(sw.requests[host])
}

func (sw *SlidingWindow) GetHostDiversityScore() float64 {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	tenSecondsAgo := now.Add(-10 * time.Second)
	uniqueHosts := 0
	totalRequests := 0

	for host, timestamps := range sw.requests {
		// Clean up old requests for the host
		i := 0
		for _, t := range timestamps {
			if t.After(tenSecondsAgo) {
				break
			}
			i++
		}
		sw.requests[host] = sw.requests[host][i:]

		// If there are any requests in the last 10 seconds, count the host
		if len(sw.requests[host]) > 0 {
			uniqueHosts++
			totalRequests += len(sw.requests[host])
		}
	}

	if totalRequests == 0 {
		return 0.0
	}

	// Calculate the diversity score
	diversityScore := float64(uniqueHosts) / float64(totalRequests)
	return diversityScore
}
