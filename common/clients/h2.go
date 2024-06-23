package clients

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/lormars/octohunter/internal/logger"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

func customh2DialTLSContext(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}
	conn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		logger.Debugf("Error dialing: %v\n", err)
		return nil, err
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		logger.Debugf("Error splitting host and port: %v\n", err)
		return nil, err
	}
	config := &utls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
	}
	tlsConn := utls.UClient(conn, config, utls.HelloRandomized)
	err = tlsConn.Handshake()
	if err != nil {
		logger.Debugf("Error handshaking: %v\n", err)
		return nil, err
	}
	return tlsConn, nil
}

// Custom transport using utls for TLS fingerprinting
func CreateCustomh2Transport() (*http2.Transport, error) {
	transport := &http2.Transport{
		DialTLSContext:     customh2DialTLSContext,
		DisableCompression: true,
		AllowHTTP:          false,
	}

	return transport, nil
}

var customh2Transport, _ = CreateCustomh2Transport()
var loggingh2Transport = WrapTransport(customh2Transport)
var Normalh2Client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	},
	Transport: loggingh2Transport,
	Timeout:   30 * time.Second,
}

var NoRedirecth2Client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: loggingh2Transport,
	Timeout:   30 * time.Second,
}
