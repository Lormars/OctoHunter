package clients

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/lormars/octohunter/asset"
	"github.com/lormars/octohunter/internal/logger"
	"golang.org/x/time/rate"
)

//h0 clients are normal clients, support h2 and h1.
//h1 clients are clients that only support h1.
//h2 clients are clients that only support h2.
//All of these clients use logging round tripper to log requests and responses.
//All of these clients use utls to mimic browser fingerprints.

var (
	rl              = 2
	cleanupInterval = 60 * time.Second
	maxIdleTime     = 120 * time.Second
)

type rateLimiterEntry struct {
	ratelimiter *rate.Limiter
	lastUsed    time.Time
}

type LoggingRoundTripper struct {
	Proxied      http.RoundTripper
	ratelimiters map[string]*rateLimiterEntry
	mu           sync.Mutex
}

func SetRateLimiter(r int) {
	rl = r
}

var AllClients = map[string]*http.Client{
	"Normalh1Client":     Normalh1Client,
	"NoRedirecth1Client": NoRedirecth1Client,
	"Normalh2Client":     Normalh2Client,
	"NoRedirecth2Client": NoRedirecth2Client,
}

// A custom Roundtrip that can log, rate limit and cleanup rate limiters.
// The ratelimiter works host-wise, so each host has its own rate limiter.
// The rate limiting depends on the option rl, which is the rate limit per second.
// The cleanup interval is 60 seconds and the max idle time is 120 seconds.
func (lrt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	currentHost := req.URL.Host
	lrt.mu.Lock()
	entry, exists := lrt.ratelimiters[currentHost]
	if !exists {
		entry = &rateLimiterEntry{
			ratelimiter: rate.NewLimiter(rate.Every(1*time.Second), rl),
			lastUsed:    time.Now(),
		}
		lrt.ratelimiters[currentHost] = entry

	} else {
		entry.lastUsed = time.Now()
	}

	lrt.mu.Unlock()
	logger.Debugf("Rate limit for %s at %s\n", currentHost, time.Now())
	err := entry.ratelimiter.Wait(req.Context())
	if err != nil {
		logger.Warnf("Rate limit Error for %s: %v\n", currentHost, err)
		return nil, err
	}

	start := time.Now()

	logger.Debugf("Making request: %s %s at %s\n", req.Method, currentHost, start)

	randomIndex := rand.Intn(len(asset.Useragent))
	randomAgent := asset.Useragent[randomIndex]
	req.Header.Set("User-Agent", randomAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	//fmt.Println(req.Header)
	resp, err := lrt.Proxied.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		logger.Debugf("Request failed: %s %s %v (%v)\n", req.Method, req.URL.String(), err, duration)
	} else {
		logger.Debugf("Response: %s %s %d (%v)\n", req.Method, req.URL.String(), resp.StatusCode, duration)
	}
	return resp, err
}

func (lrt *LoggingRoundTripper) cleanupRateLimiters() {
	for {
		time.Sleep(cleanupInterval)
		now := time.Now()
		lrt.mu.Lock()
		for host, entry := range lrt.ratelimiters {
			if now.Sub(entry.lastUsed) > maxIdleTime {
				logger.Debugf("Cleaning up rate limiter for %s\n", host)
				delete(lrt.ratelimiters, host)
			}

		}
		lrt.mu.Unlock()
	}
}

func WrapTransport(transport http.RoundTripper) http.RoundTripper {
	lrt := &LoggingRoundTripper{
		Proxied:      transport,
		ratelimiters: make(map[string]*rateLimiterEntry),
	}
	go lrt.cleanupRateLimiters()
	return lrt
}
