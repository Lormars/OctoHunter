package clients

import (
	"context"
	"crypto/tls"
	"net"

	rchttp2 "github.com/lormars/http2/http2"
	"github.com/lormars/octohunter/internal/logger"
)

func createRcH2DialTLSContext(proxy string) func(context.Context, string, string, *tls.Config) (net.Conn, error) {
	return func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
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
			if proxy != "" {
				conn, err = dialProxy(network, ipAddr, proxy)
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

		if err := conn.(*net.TCPConn).SetNoDelay(false); err != nil {
			logger.Errorf("Error setting TCP_NODELAY: %v\n", err)
			return nil, err
		}

		return handshake(host, "h2", conn)
	}
}

// Custom transport using utls for TLS fingerprinting
func CreateRCh2Transport(proxy string) *rchttp2.Transport {
	transport := &rchttp2.Transport{
		DialTLSContext: createRcH2DialTLSContext(proxy),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return transport
}
