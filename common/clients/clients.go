package clients

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/lormars/octohunter/asset"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients/health"
	"github.com/lormars/octohunter/common/clients/proxyP"
	"github.com/lormars/octohunter/internal/logger"
)

//h0 clients are normal clients, support h2 and h1.
//h1 clients are clients that only support h1.
//h2 clients are clients that only support h2.
//All of these clients use logging round tripper to log requests and responses.
//All of these clients use utls to mimic browser fingerprints.

var (
	rl                     = 4
	cleanupInterval        = 60 * time.Second
	maxIdleTime            = 120 * time.Second
	totalData        int64 = 0
	concurrentReq          = make(chan struct{}, 100)
	mu               sync.Mutex
	resStats         = make(map[string][]*responseStats)
	allRequestsCount = 0
	errRequestsCount = 0
	UseProxy         = false
)

type responseStats struct {
	statusCode int
	duration   time.Duration
}

type rateLimiterEntry struct {
	ratelimiter *common.DynamicRateLimiter
	lastUsed    time.Time
}

type LoggingRoundTripper struct {
	Proxied http.RoundTripper
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

var ratelimiters = make(map[string]map[string]*rateLimiterEntry)

// A custom Roundtrip that can log, rate limit and cleanup rate limiters.
// The ratelimiter works host-wise, so each host has its own rate limiter.
// The rate limiting depends on the option rl, which is the rate limit per second.
// The cleanup interval is 60 seconds and the max idle time is 120 seconds.
func (lrt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var proxy string
	var ok bool
	maxRetries := 2
	retryDelay := 1 * time.Second
	acquireSemaphore := func() {
		concurrentReq <- struct{}{}
		logger.Debugf("acquired")
	}

	releaseSemaphore := func() {
		<-concurrentReq
		logger.Debugf("released")
	}

	// Retry loop
	for attempt := 0; attempt < maxRetries; attempt++ {
		//first check if the request has a proxy set
		proxy, ok = req.Context().Value("proxy").(string)
		if ok {
			//fmt.Println("Using proxy: ", proxy)
		} else {
			//if not, randomly select a proxy from the list
			proxyP.Proxies.Mu.Lock()
			proxy = proxyP.Proxies.Proxies[rand.Intn(len(proxyP.Proxies.Proxies))]
			proxyP.Proxies.Mu.Unlock()
			ctx := context.WithValue(req.Context(), "proxy", proxy)
			req = req.WithContext(ctx)
		}

		currentHost := req.URL.Hostname()
		var entry *rateLimiterEntry
		var proxyExists bool
		//does not apply rate limit when testing for race condition
		_, okrace := req.Context().Value("race").(string)
		if !okrace {
			mu.Lock()
			//the rate limiter is host & proxy specific
			proxyentry, exists := ratelimiters[currentHost]
			if !exists {
				proxyentry = make(map[string]*rateLimiterEntry)
				ratelimiters[currentHost] = proxyentry
			}
			entry, proxyExists = proxyentry[proxy]
			if !proxyExists {
				entry = &rateLimiterEntry{
					ratelimiter: common.NewDynamicRateLimiter(float64(2), rl),
					lastUsed:    time.Now(),
				}
				proxyentry[proxy] = entry

			} else {
				entry.lastUsed = time.Now()
			}

			logger.Debugf("Rate limit for %s at %f\n", currentHost, entry.ratelimiter.GetRPS())

			mu.Unlock()

			err := entry.ratelimiter.Wait(req.Context())
			if err != nil {
				logger.Debugf("Rate limit Error for %s: %v\n", currentHost, err)
				//mostly it is because exceeding context deadline
				mu.Lock()
				newrl := entry.ratelimiter.GetRPS() + 1
				newbt := entry.ratelimiter.GetBurst() + 1
				entry.ratelimiter.Update(newrl, newbt)
				logger.Debugf("Increase Rate limit for %s to %f\n", currentHost, newrl)
				mu.Unlock()
				return nil, err
			}
		}

		start := time.Now()

		logger.Debugf("Making request: at %s\n", req.URL.String())

		randomIndex := rand.Intn(len(asset.Useragent))
		randomAgent := asset.Useragent[randomIndex]
		req.Header.Add("User-Agent", randomAgent) //use ADD as RC would set the user agent
		req.Header.Add("Accept-Charset", "utf-8")

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

		acquireSemaphore()
		// Measure concurrent requests

		proxiedCtx, cancel := context.WithTimeout(req.Context(), 10*time.Second)
		req = req.WithContext(proxiedCtx)

		mu.Lock()
		allRequestsCount++
		common.Sliding.AddRequest(currentHost)
		mu.Unlock()
		resp, err := lrt.Proxied.RoundTrip(req)
		cancel()
		releaseSemaphore()
		duration := time.Since(start)

		if err != nil {
			logger.Debugf("Request failed: %s %s %v (%v)\n", req.Method, req.URL.String(), err, duration)
			// If this was not the last attempt, wait before retrying
			if attempt < maxRetries-1 {
				mu.Lock()
				allRequestsCount--
				mu.Unlock()
				time.Sleep(retryDelay)
				continue
			}

			mu.Lock()
			health.ProxyHealthInstance.AddBad(proxy)
			errRequestsCount++
			mu.Unlock()
			return nil, err
		}

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
		respStats := &responseStats{
			statusCode: resp.StatusCode,
			duration:   duration,
		}
		resStats[currentHost] = append(resStats[currentHost], respStats)

		if resp.StatusCode == 429 {
			if !okrace {
				if entry != nil {
					newrl := entry.ratelimiter.GetRPS() - 1
					newbt := entry.ratelimiter.GetBurst() - 1
					entry.ratelimiter.Update(newrl, newbt)
					logger.Debugf("Decrease Rate limit for %s to %f\n", currentHost, newrl)
				}
			}
		} else if resp.StatusCode == 403 {
			health.ProxyHealthInstance.AddBad(proxy)
		} else {
			health.ProxyHealthInstance.AddGood(proxy)

			newrl := entry.ratelimiter.GetRPS() + 1
			newbt := entry.ratelimiter.GetBurst() + 1
			entry.ratelimiter.Update(newrl, newbt)
			logger.Debugf("Increase Rate limit for %s to %f\n", currentHost, newrl)
		}
		logger.Debugf("Current ratelimit for %s: %f\n", currentHost, entry.ratelimiter.GetRPS())
		mu.Unlock()

		logger.Debugf("Response: %s %s %d (%v)\n", req.Method, req.URL.String(), resp.StatusCode, duration)

		return resp, nil
	}

	// If we exit the loop, it means all retries failed
	return nil, fmt.Errorf("request failed after %d attempts", maxRetries)
}

func cleanupRateLimiters() {
	for {
		time.Sleep(cleanupInterval)
		now := time.Now()
		mu.Lock()
		for host, proxyentry := range ratelimiters {
			for _, entry := range proxyentry {
				if now.Sub(entry.lastUsed) > maxIdleTime {
					logger.Debugf("Cleaning up rate limiter for %s\n", host)
					delete(ratelimiters, host)
				}
			}
		}
		mu.Unlock()
	}
}

func WrapTransport(transport http.RoundTripper) http.RoundTripper {
	lrt := &LoggingRoundTripper{
		Proxied: transport,
	}
	go cleanupRateLimiters()
	return lrt
}

func SetUseProxy(useProxy bool) {
	UseProxy = useProxy
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
	return len(concurrentReq)
}

// This calculates the bad rate and error rate for all hosts for all requests.
func PrintResStats() string {
	mu.Lock()
	defer mu.Unlock()
	// Iterate through each host and count status codes
	count := 0
	all := 0
	for _, stats := range resStats {
		counts := map[string]int{
			"200-300": 0,
			"300-400": 0,
			"400-500": 0,
			"500+":    0,
			"403":     0,
		}
		total := 0

		for _, stat := range stats {
			total++
			switch {
			case stat.statusCode >= 200 && stat.statusCode < 300:
				counts["200-300"]++
			case stat.statusCode >= 300 && stat.statusCode < 400:
				counts["300-400"]++
			case stat.statusCode >= 400 && stat.statusCode < 500:
				if stat.statusCode == 403 {
					counts["403"]++
				} else {
					counts["400-500"]++
				}
			case stat.statusCode >= 500:
				counts["500+"]++
			}
		}
		threshold := 50.0
		percentage403 := float64(counts["403"]) / float64(total) * 100
		if percentage403 > threshold {
			// Print the results for the current host in a single line
			//fmt.Printf("Host: %s | 200-300: %d | 300-400: %d | 400-500 (except 403): %d | 500+: %d | 403: %d\n",
			//	host, counts["200-300"], counts["300-400"], counts["400-500"], counts["500+"], counts["403"])
			count++
		}
		all++
	}
	percentageBad := float64(count) / float64(all) * 100
	percentageErr := float64(errRequestsCount) / float64(allRequestsCount) * 100
	return fmt.Sprintf("Bad rate: %.2f. Err rate: %.2f", percentageBad, percentageErr)
}
