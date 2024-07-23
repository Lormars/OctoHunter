package clients

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/lormars/octohunter/internal/logger"
	"golang.org/x/net/http2"
	"golang.org/x/net/proxy"
)

type H0Transport struct {
	h1Transport *http.Transport
	h2Transport *http2.Transport
}

// RoundTrip implements the RoundTripper interface
func (t *H0Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == "https" {
		logger.Debugf("Attempting HTTP/2 for %s\n", req.URL.String())
		resp, err := t.h2Transport.RoundTrip(req)
		if err == nil {
			//logger.Warnf("HTTP/2 request succeeded for %s\n", req.URL.String())
			return resp, nil
		}
		// logger.Warnf("HTTP/2 request failed: %v, falling back to HTTP/1.1 for %s\n", err, req.URL.String())
	}
	return t.h1Transport.RoundTrip(req)
}

// createCombinedTransport creates a transport that supports both HTTP/2 and HTTP/1.1
func createH0Transport() (*H0Transport, error) {
	h2Transport, err := CreateCustomh2Transport()
	if err != nil {
		logger.Debugf("Error creating h2 transport: %v\n", err)
		return nil, err
	}
	h1Transport := CreateCustomh1Transport()

	return &H0Transport{
		h1Transport: h1Transport,
		h2Transport: h2Transport,
	}, nil
}

// Create clients with the combined transport
var h0Transport, _ = createH0Transport()
var loggingH0Transport = WrapTransport(h0Transport)
var NormalClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	},
	Transport: loggingH0Transport,
}

var NoRedirectClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: loggingH0Transport,
}

func dial(ctx context.Context, network, addr string) (net.Conn, error) {
	var conn net.Conn
	var err error
	if UseProxy {
		proxyStr, _ := ctx.Value("proxy").(string)
		auth := &proxy.Auth{
			User:     os.Getenv("PROXY_USER"),
			Password: os.Getenv("PROXY_PASS"),
		}
		dialer, err := proxy.SOCKS5("tcp", proxyStr, auth, proxy.Direct)
		if err != nil {
			logger.Warnf("Error dialing: %v\n", err)
			return nil, err
		}

		conn, err = dialer.Dial(network, addr)
		if err != nil {
			logger.Debugf("Error dialing: %v\n", err)
			return nil, err
		}
	} else {
		dialer := &net.Dialer{
			Timeout: 10 * time.Second,
		}
		conn, err = dialer.DialContext(ctx, network, addr)
		if err != nil {
			logger.Debugf("Error dialing: %v\n", err)
			return nil, err
		}
	}
	return conn, nil
}
