package clients

import (
	"bytes"
	"context"
	"io"
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
	rl                    = 2
	cleanupInterval       = 60 * time.Second
	maxIdleTime           = 120 * time.Second
	totalData       int64 = 0
	concurrentReq         = 0
	mu              sync.Mutex
	Proxies         = ParseProxies()
)

type rateLimiterEntry struct {
	ratelimiter *rate.Limiter
	lastUsed    time.Time
}

type LoggingRoundTripper struct {
	Proxied      http.RoundTripper
	ratelimiters map[string]map[string]*rateLimiterEntry
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
	var proxy string
	var ok bool
	//first check if the request has a proxy set
	proxy, ok = req.Context().Value("proxy").(string)
	if ok {
		//fmt.Println("Using proxy: ", proxy)
	} else {
		//if not, randomly select a proxy from the list
		proxy = Proxies[rand.Intn(len(Proxies))]
		ctx := context.WithValue(req.Context(), "proxy", proxy)
		req = req.WithContext(ctx)
	}

	currentHost := req.URL.Host
	lrt.mu.Lock()
	//the rate limiter is host & proxy specific
	proxyentry, exists := lrt.ratelimiters[currentHost]
	if !exists {
		proxyentry = make(map[string]*rateLimiterEntry)
		lrt.ratelimiters[currentHost] = proxyentry
	}
	entry, proxyExists := proxyentry[proxy]
	if !proxyExists {
		entry = &rateLimiterEntry{
			//TODO: dynamic rate limiter
			ratelimiter: rate.NewLimiter(rate.Limit(1), rl),
			lastUsed:    time.Now(),
		}
		proxyentry[proxy] = entry

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

	logger.Debugf("Making request: %s %s/%s at %s\n", req.Method, currentHost, req.URL.Path, start)

	randomIndex := rand.Intn(len(asset.Useragent))
	randomAgent := asset.Useragent[randomIndex]
	req.Header.Set("User-Agent", randomAgent)
	req.Header.Set("Accept-Charset", "utf-8")

	// Measure request size
	requestSize := int64(0)
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		requestSize = int64(len(body))
		req.Body = io.NopCloser(io.MultiReader(bytes.NewReader(body)))
	}
	// Add the size of request headers
	requestSize += int64(len(req.Method) + len(req.URL.String()) + len(req.Proto) + 4) // request line size
	for name, values := range req.Header {
		for _, value := range values {
			requestSize += int64(len(name) + len(value) + 4) // header field size
		}
	}

	// Measure concurrent requests
	mu.Lock()
	concurrentReq++
	mu.Unlock()

	//fmt.Println(req.Header)
	resp, err := lrt.Proxied.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		logger.Debugf("Request failed: %s %s %v (%v)\n", req.Method, req.URL.String(), err, duration)
	} else {
		// Measure response size
		responseSize := int64(0)
		if resp.Body != nil {
			body, _ := io.ReadAll(resp.Body)
			responseSize += int64(len(body))
			resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(body)))
		}
		// Add the size of response headers
		responseSize += int64(len(resp.Proto) + 3 + 3 + 2) // status line size
		for name, values := range resp.Header {
			for _, value := range values {
				responseSize += int64(len(name) + len(value) + 4) // header field size
			}
		}

		mu.Lock()
		totalData += requestSize + responseSize
		mu.Unlock()

		logger.Debugf("Response: %s %s %d (%v)\n", req.Method, req.URL.String(), resp.StatusCode, duration)
	}

	mu.Lock()
	concurrentReq--
	mu.Unlock()
	return resp, err
}

func (lrt *LoggingRoundTripper) cleanupRateLimiters() {
	for {
		time.Sleep(cleanupInterval)
		now := time.Now()
		lrt.mu.Lock()
		for host, proxyentry := range lrt.ratelimiters {
			for _, entry := range proxyentry {
				if now.Sub(entry.lastUsed) > maxIdleTime {
					logger.Debugf("Cleaning up rate limiter for %s\n", host)
					delete(lrt.ratelimiters, host)
				}
			}
		}
		lrt.mu.Unlock()
	}
}

func WrapTransport(transport http.RoundTripper) http.RoundTripper {
	lrt := &LoggingRoundTripper{
		Proxied:      transport,
		ratelimiters: make(map[string]map[string]*rateLimiterEntry),
	}
	go lrt.cleanupRateLimiters()
	return lrt
}

// TODO: clean up these two functions' call in mu
// GetTotalDataTransferred returns the total data transferred in GB.
func GetTotalDataTransferred() float64 {
	mu.Lock()
	defer mu.Unlock()
	return float64(totalData) / (1024 * 1024 * 1024) // Convert bytes to GB
}

func GetConcurrentRequests() int {
	mu.Lock()
	defer mu.Unlock()
	return concurrentReq
}
