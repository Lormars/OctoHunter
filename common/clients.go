package common

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/lormars/octohunter/internal/logger"
)

type LoggingRoundTripper struct {
	Proxied http.RoundTripper
}

func (lrt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	logger.Debugf("Making request: %s %s\n", req.Method, req.URL.String())

	resp, err := lrt.Proxied.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		logger.Debugf("Request failed: %s %s %v (%v)\n", req.Method, req.URL.String(), err, duration)
	} else {
		logger.Debugf("Response: %s %s %d (%v)\n", req.Method, req.URL.String(), resp.StatusCode, duration)
	}

	return resp, err
}

func WrapTransport(transport http.RoundTripper) http.RoundTripper {
	return &LoggingRoundTripper{Proxied: transport}
}

var loggingTransport = WrapTransport(http.DefaultTransport)

var NormalClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	},
	Transport: loggingTransport,
	Timeout:   10 * time.Second,
}

var NoRedirectClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: loggingTransport,
	Timeout:   10 * time.Second,
}

// disable auto upgrade to http2, force http/1.1
var http1transport = &http.Transport{
	TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
}

var NoRedirectHTTP1Client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: WrapTransport(http1transport),
	Timeout:   10 * time.Second,
}
