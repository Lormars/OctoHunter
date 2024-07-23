package clients

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/lormars/octohunter/internal/logger"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

func customh2DialTLSContext(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		logger.Warnf("Error splitting host and port: %v\n", err)
		return nil, err
	}

	ips, err := DnsCache.LookupIP(host)
	if err != nil {
		logger.Warnf("Error looking up IP: %v\n", err)
		return nil, err
	}

	var conn net.Conn
	for _, ip := range ips {
		ipAddr := net.JoinHostPort(ip.String(), port)
		conn, err = dial(ctx, network, ipAddr)
		if err == nil {
			break
		}
		logger.Warnf("Error dialing IP %v: %v\n", ipAddr, err)
	}

	if err != nil {
		return nil, err
	}

	config := &utls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		NextProtos: []string{"h2"},
	}
	tlsConn := utls.UClient(conn, config, utls.HelloRandomizedALPN)
	err = tlsConn.Handshake()
	if err != nil {
		logger.Warnf("Error handshaking: %v\n", err)
		return nil, err
	}
	return tlsConn, nil
}

// Custom transport using utls for TLS fingerprinting
func CreateCustomh2Transport() (*http2.Transport, error) {
	transport := &http2.Transport{
		DialTLSContext:     customh2DialTLSContext,
		DisableCompression: false,
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
}

var NoRedirecth2Client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: loggingh2Transport,
}
