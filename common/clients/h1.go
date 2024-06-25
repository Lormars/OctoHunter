package clients

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/lormars/octohunter/internal/logger"
	utls "github.com/refraction-networking/utls"
)

// Custom dialer for utls
func customh1DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {
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
	config := &utls.Config{ServerName: host}
	tlsConn := utls.UClient(conn, config, utls.HelloRandomized)
	err = tlsConn.Handshake()
	if err != nil {
		logger.Debugf("Error handshaking: %v\n", err)
		return nil, err
	}
	return tlsConn, nil
}

func CreateCustomh1Transport() *http.Transport {
	transport := &http.Transport{
		DialTLSContext:    customh1DialTLSContext,
		ForceAttemptHTTP2: false,
	}

	return transport
}

func KeepAliveh1Transport() *http.Transport {
	transport := &http.Transport{
		DialTLSContext:      customh1DialTLSContext,
		ForceAttemptHTTP2:   false,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 1,
	}

	return transport
}

var customh1Transport = CreateCustomh1Transport()
var loggingh1Transport = WrapTransport(customh1Transport)

var keepAliveh1Transport = WrapTransport(KeepAliveh1Transport())

var Normalh1Client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	},
	Transport: loggingh1Transport,
	Timeout:   30 * time.Second,
}

var NoRedirecth1Client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: loggingh1Transport,
	Timeout:   30 * time.Second,
}

var KeepAliveh1Client = &http.Client{
	Transport: keepAliveh1Transport,
	Timeout:   30 * time.Second,
}
