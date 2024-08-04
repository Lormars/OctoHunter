package clients

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/lormars/octohunter/internal/logger"
)

func createCustomH1DialTLSContext(proxy string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
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
		}
		if err != nil {
			logger.Debugf("Error dialing: %v\n", err)
			return nil, err
		}
		return handshake(ctx, host, "http/1.1", conn)
	}
}

func createCustomDialContect(proxy string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
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
		}
		if err != nil {
			logger.Debugf("Error dialing: %v\n", err)
			return nil, err
		}
		return conn, nil
	}
}

func CreateCustomh1Transport(proxy string) *http.Transport {
	transport := &http.Transport{
		DialContext:       createCustomDialContect(proxy),
		DialTLSContext:    createCustomH1DialTLSContext(proxy),
		ForceAttemptHTTP2: false,
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return transport
}

func KeepAliveh1Transport(proxy string) *http.Transport {
	transport := &http.Transport{
		DialContext:         createCustomDialContect(proxy),
		DialTLSContext:      createCustomH1DialTLSContext(proxy),
		ForceAttemptHTTP2:   false,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 1,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return transport
}
