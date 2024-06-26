package clients

import (
	"net/http"
	"time"

	"github.com/lormars/octohunter/internal/logger"
	"golang.org/x/net/http2"
)

type H0Transport struct {
	h1Transport *http.Transport
	h2Transport *http2.Transport
}

// RoundTrip implements the RoundTripper interface
func (t *H0Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == "https" {
		// Attempt HTTP/2 first
		ctx := req.Context()
		address := req.URL.Host + ":443"
		tlsConn, err := t.h2Transport.DialTLSContext(ctx, "tcp", address, nil)
		if err == nil {
			http2ClientConn, err := t.h2Transport.NewClientConn(tlsConn)
			if err == nil {
				return http2ClientConn.RoundTrip(req)
			}
		}
	}
	logger.Debugf("Falling back to HTTP/1.1 for %s\n", req.URL.String())

	// Fallback to HTTP/1.1
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
	Timeout:   120 * time.Second,
}

var NoRedirectClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: loggingH0Transport,
	Timeout:   120 * time.Second,
}
