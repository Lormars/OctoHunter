package clients

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/lormars/octohunter/internal/logger"
	"golang.org/x/net/http2"
)

func createCustomH2DialTLSContext(proxy string) func(context.Context, string, string, *tls.Config) (net.Conn, error) {
	return func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
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
			if proxy != "" {
				conn, err = dialProxy(ctx, network, ipAddr, proxy)
			} else {
				conn, err = dial(ctx, network, ipAddr)
			}
			if err == nil {
				break
			}
			logger.Debugf("Error dialing IP %v: %v\n", ipAddr, err)
		}

		if err != nil {
			return nil, err
		}
		return handshake(ctx, host, "h2", conn)
	}
}

// Custom transport using utls for TLS fingerprinting
func CreateCustomh2Transport(proxy string) *http2.Transport {
	transport := &http2.Transport{
		DialTLSContext:     createCustomH2DialTLSContext(proxy),
		DisableCompression: false,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return transport
}
