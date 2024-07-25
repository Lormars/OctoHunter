package clients

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	rchttp2 "github.com/lormars/http2/http2"
	"github.com/lormars/octohunter/internal/logger"
)

func rch2DialTLSContext(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		logger.Debugf("Error splitting host and port: %v\n", err)
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

	if err := conn.(*net.TCPConn).SetNoDelay(false); err != nil {
		logger.Errorf("Error setting TCP_NODELAY: %v\n", err)
		return nil, err
	}

	return handshake(ctx, host, "h2", conn)
}

// Custom transport using utls for TLS fingerprinting
func CreateRCh2Transport() (*rchttp2.Transport, error) {
	transport := &rchttp2.Transport{
		DialTLSContext: rch2DialTLSContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return transport, nil
}

var rch2Transport, _ = CreateRCh2Transport()
var loggingRCh2Transport = WrapTransport(rch2Transport)

var NormalRCClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	},
	Transport: loggingRCh2Transport,
}

var NoRedirectRCClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: loggingRCh2Transport,
}
