package clients

import (
	"context"
	"net"
	"net/http"

	"github.com/lormars/octohunter/internal/logger"
	utls "github.com/refraction-networking/utls"
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

	config := &utls.Config{
		ServerName: host,
		NextProtos: []string{"http/1.1"},
	}
	tlsConn := utls.UClient(conn, config, utls.HelloRandomizedALPN)
	err = tlsConn.Handshake()
	if err != nil {
		logger.Warnf("Error handshaking: %v\n", err)
		return nil, err
	}
	return tlsConn, nil
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
