package clients

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/lormars/octohunter/internal/logger"
)

// Custom dialer for utls
func customh1DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {

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
	return handshake(ctx, host, "http/1.1", conn)
}

func customDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
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
	}
	if err != nil {
		logger.Warnf("Error dialing: %v\n", err)
		return nil, err
	}
	return conn, nil
}

func CreateCustomh1Transport() *http.Transport {
	transport := &http.Transport{
		DialContext:       customDialContext,
		DialTLSContext:    customh1DialTLSContext,
		ForceAttemptHTTP2: false,
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return transport
}

func KeepAliveh1Transport() *http.Transport {
	transport := &http.Transport{
		DialContext:         customDialContext,
		DialTLSContext:      customh1DialTLSContext,
		ForceAttemptHTTP2:   false,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 1,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return transport
}

var customh1Transport = CreateCustomh1Transport()
var loggingh1Transport = WrapTransport(customh1Transport)

var Normalh1Client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	},
	Transport: loggingh1Transport,
}

var NoRedirecth1Client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: loggingh1Transport,
}
