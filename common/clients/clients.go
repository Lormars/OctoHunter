package clients

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lormars/octohunter/asset"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients/proxyP"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/logger"
)

//h0 clients are normal clients, support h2 and h1.
//h1 clients are clients that only support h1.
//h2 clients are clients that only support h2.
//All of these clients use logging round tripper to log requests and responses.
//All of these clients use utls to mimic browser fingerprints.

var (
	totalData        int64 = 0
	mu               sync.Mutex
	resStats         = make(map[string][]*responseStats)
	allRequestsCount = 0
	errRequestsCount = 0
	all429Count      = 0
	UseProxy         = false
	DnsCache         = NewDNSCache()
	Clients          *OctoClients

	rl = 4
)

type responseStats struct {
	statusCode int
	duration   time.Duration
}

type LoggingRoundTripper struct {
	Proxied http.RoundTripper
}

type OctoClients struct {
	clients map[string]*OctoClient
}

type OctoClient struct {
	client    *http.Client
	name      string
	rateLimit map[string]*common.RateLimiterEntry
	mu        sync.Mutex
	proxy     string
}

func init() {
	initClients()
}

func initClients() {
	Clients = &OctoClients{
		clients: make(map[string]*OctoClient),
	}
	proxies := [3]string{"", proxyP.Proxies.Proxies[0], proxyP.Proxies.Proxies[1]} //HACK: hard coded
	var client *OctoClient
	//generate clients
	for _, proxy := range proxies {
		client = NewClient("h1NA", proxy, true, WrapTransport(CreateCustomh1Transport(proxy)))
		Clients.clients[client.name] = client
		client = NewClient("h1KA", proxy, true, WrapTransport(KeepAliveh1Transport(proxy)))
		Clients.clients[client.name] = client
		client = NewClient("h1KA", proxy, false, WrapTransport(KeepAliveh1Transport(proxy)))
		Clients.clients[client.name] = client
		client = NewClient("h1NA", proxy, false, WrapTransport(CreateCustomh1Transport(proxy)))
		Clients.clients[client.name] = client
		client = NewClient("h2", proxy, true, WrapTransport(CreateCustomh2Transport(proxy)))
		Clients.clients[client.name] = client
		client = NewClient("h2", proxy, false, WrapTransport(CreateCustomh2Transport(proxy)))
		Clients.clients[client.name] = client
		client = NewClient("h0", proxy, true, WrapTransport(CreateH0Transport(proxy)))
		Clients.clients[client.name] = client
		client = NewClient("h0", proxy, false, WrapTransport(CreateH0Transport(proxy)))
		Clients.clients[client.name] = client
		client = NewClient("rc", proxy, true, WrapTransport(CreateRCh2Transport(proxy)))
		Clients.clients[client.name] = client
		client = NewClient("rc", proxy, false, WrapTransport(CreateRCh2Transport(proxy)))
		Clients.clients[client.name] = client
	}

}

func NewClient(cType, proxy string, redirect bool, transport http.RoundTripper) *OctoClient {
	var client *http.Client
	var name string
	if redirect {
		name = "Normal"
		client = &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
		}
	} else {
		name = "NoRedirect"
		client = &http.Client{
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: 60 * time.Second,
		}
	}

	if proxy != "" {
		return &OctoClient{
			client:    client,
			name:      name + fmt.Sprintf("%sProxy%s", cType, proxy),
			proxy:     proxy,
			rateLimit: make(map[string]*common.RateLimiterEntry),
		}
	} else {
		return &OctoClient{
			client:    client,
			name:      name + cType + "NoProxy",
			proxy:     proxy,
			rateLimit: make(map[string]*common.RateLimiterEntry),
		}
	}
}

func (oc *OctoClient) RetryableDo(req *http.Request) (*http.Response, error) {
	serializedReq := serializeRequest(req)
	hashed := common.Hash(serializedReq)
	if !cacher.CheckCache(hashed, oc.name) {
		return nil, fmt.Errorf("request already made")
	}
	maxRetries := 3
	retryDelay := 1 * time.Second
	var resp *http.Response
	var err error
	currentHost := req.URL.Hostname()

	randomIndex := rand.Intn(len(asset.Useragent))
	randomAgent := asset.Useragent[randomIndex]
	req.Header.Add("User-Agent", randomAgent) //use ADD as RC would set the user agent
	req.Header.Add("Accept-Charset", "utf-8")
	// Measure request size
	requestSize := int64(0)
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		requestSize = int64(len(body))
		req.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	// Add the size of request headers
	requestSize += int64(len(req.Method) + len(req.URL.String()) + len(req.Proto) + 4) // request line size
	for name, values := range req.Header {
		for _, value := range values {
			requestSize += int64(len(name) + len(value) + 4) // header field size
		}
	}
	for attempt := 0; attempt < maxRetries; attempt++ {

		mu.Lock()
		allRequestsCount++
		common.Sliding.AddRequest(currentHost)
		mu.Unlock()
		start := time.Now()
		resp, err = oc.Do(req)
		duration := time.Since(start)

		if err != nil {
			logger.Debugf("Request failed: %s %s %v\n", req.Method, req.URL.String(), err)
			// If this was not the last attempt, wait before retrying
			if attempt < maxRetries-1 {
				mu.Lock()
				allRequestsCount--
				mu.Unlock()
				time.Sleep(retryDelay)
				continue
			}
			logger.Debugf("Request failed after %d attempts: %s %s %v\n", maxRetries, req.Method, req.URL.String(), err)

			mu.Lock()
			errRequestsCount++
			mu.Unlock()
			return nil, err
		}
		// Measure response size
		responseSize := int64(0)
		if resp.Body != nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			responseSize += int64(len(body))
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		// Add the size of response headers
		responseSize += int64(len(resp.Proto) + 3 + 3 + 2) // status line size
		for name, values := range resp.Header {
			for _, value := range values {
				responseSize += int64(len(name) + len(value) + 4) // header field size
			}
		}
		logger.Debugf("Response size: %d\n", responseSize)
		mu.Lock()
		totalData += requestSize + responseSize
		respStats := &responseStats{
			statusCode: resp.StatusCode,
			duration:   duration,
		}
		resStats[currentHost] = append(resStats[currentHost], respStats)
		mu.Unlock()
		break
	}

	if !strings.Contains(oc.name, "rc") {
		entry, exists := oc.rateLimit[req.URL.Hostname()]
		if exists {
			if resp.StatusCode == 429 {
				entry.Successes = 0
				entry.Failures++
				currentRPS := entry.Ratelimiter.GetRPS()
				if currentRPS > 2 && !entry.Hit && entry.Failures > 10 {
					all429Count++
					entry.Failures = 0
					entry.Hit = true
					newrl := currentRPS - 10
					newbt := entry.Ratelimiter.GetBurst() - 10
					entry.Max = int(currentRPS - 10)
					entry.Ratelimiter.Update(newrl, newbt)
					logger.Debugf("Decrease Rate limit for %s to %f and set max to %d\n", req.URL.Hostname(), newrl, entry.Max)
				}
			} else {
				entry.Successes++
				if entry.Successes > 300 && !entry.Hit { //to make sure the current rate is not too high
					entry.Successes = 0
					if entry.Max > int(entry.Ratelimiter.GetRPS()) {
						newrl := entry.Ratelimiter.GetRPS() + 1
						newbt := entry.Ratelimiter.GetBurst() + 1
						entry.Ratelimiter.Update(newrl, newbt)
						logger.Debugf("Increase Rate limit for %s to %f\n", req.URL.Hostname(), newrl)
					} else {
						entry.Max += 5
					}
				} else if entry.Successes > 1000 {
					entry.Hit = false
				}
			}
		}
	}

	return resp, nil
}

func (oc *OctoClient) Do(req *http.Request) (*http.Response, error) {

	if strings.Contains(oc.name, "rc") {
		return oc.client.Do(req)
	} else {
		oc.mu.Lock()
		ratelimiter, exists := oc.rateLimit[req.URL.Hostname()]
		if !exists {
			ratelimiter = &common.RateLimiterEntry{
				Ratelimiter: common.NewDynamicRateLimiter(float64(2), rl),
				LastUsed:    time.Now(),
				Max:         12,
				Hit:         false,
			}
			oc.rateLimit[req.URL.Hostname()] = ratelimiter
		}
		oc.mu.Unlock()
		err := ratelimiter.Ratelimiter.Wait(req.Context())
		if err != nil {
			oc.mu.Lock()
			if ratelimiter.Max > int(ratelimiter.Ratelimiter.GetRPS()) {
				newrl := ratelimiter.Ratelimiter.GetRPS() + 1
				newbt := ratelimiter.Ratelimiter.GetBurst() + 1
				ratelimiter.Ratelimiter.Update(newrl, newbt)
				logger.Debugf("Increase Rate limit for %s to %f\n", req.URL.Hostname(), newrl)
			}
			oc.mu.Unlock()
		}
		return oc.client.Do(req)
	}
}

func (ocs *OctoClients) GetRandomClient(ctype string, redirect, proxy bool) *OctoClient {
	var prefix string
	if redirect {
		prefix = "Normal" + ctype
		if proxy && UseProxy {
			prefix += "Proxy"
		} else {
			prefix += "NoProxy"
		}
	} else {
		prefix = "NoRedirect" + ctype
		if proxy && UseProxy {
			prefix += "Proxy"
		} else {
			prefix += "NoProxy"
		}
	}
	// Collect keys that start with the prefix
	var keys []string
	for k := range ocs.clients {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			keys = append(keys, k)
		}
	}

	// Check if there are any matching keys
	if len(keys) == 0 {
		logger.Warnf("No client found with the prefix %s\n", prefix)
		return nil // No keys found with the given prefix
	}

	// Select a random key
	randomKey := keys[rand.Intn(len(keys))]

	// Retrieve and return the value
	return ocs.clients[randomKey]
}

// A custom Roundtrip that can log, rate limit and cleanup rate limiters.
// The ratelimiter works host-wise, so each host has its own rate limiter.
// The rate limiting depends on the option rl, which is the rate limit per second.
// The cleanup interval is 60 seconds and the max idle time is 120 seconds.
func (lrt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := lrt.Proxied.RoundTrip(req)
	return resp, err
}

func WrapTransport(transport http.RoundTripper) http.RoundTripper {
	lrt := &LoggingRoundTripper{
		Proxied: transport,
	}
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

func Get429Count() int {
	mu.Lock()
	defer mu.Unlock()
	return all429Count
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
	return fmt.Sprintf("Bad: %.2f. Err: %.2f", percentageBad, percentageErr)
}

func serializeRequest(req *http.Request) string {
	var buffer bytes.Buffer
	buffer.WriteString(req.Method)
	buffer.WriteString(req.URL.String())
	buffer.WriteString(req.Proto)
	for name, values := range req.Header {
		buffer.WriteString(name)
		for _, value := range values {
			buffer.WriteString(value)
		}
	}

	if req.Body != nil {
		// Reset the request body so it can be read again
		body, _ := io.ReadAll(req.Body)
		buffer.WriteString(string(body))
		req.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	return buffer.String()
}
