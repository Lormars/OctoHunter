package clients

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/lormars/octohunter/internal/logger"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/proxy"
)

// Custom dialer for utls
func customh1DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {
	proxyStr, _ := ctx.Value("proxy").(string)
	// if ok {
	// 	fmt.Println("Using proxy: ", proxyStr)
	// } else {
	// 	fmt.Println("No proxy")
	// }
	// dialer := &net.Dialer{
	// 	Timeout: 30 * time.Second,
	// }
	auth := &proxy.Auth{
		User:     os.Getenv("PROXY_USER"),
		Password: os.Getenv("PROXY_PASS"),
	}
	dialer, err := proxy.SOCKS5("tcp", proxyStr, auth, proxy.Direct)
	if err != nil {
		logger.Warnf("Error dialing: %v\n", err)
		return nil, err
	}
	conn, err := dialer.Dial(network, addr)
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

func customDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	proxyStr, _ := ctx.Value("proxy").(string)
	// if ok {
	// 	fmt.Println("Using proxy: ", proxyStr)
	// } else {
	// 	fmt.Println("No proxy")
	// }
	// dialer := &net.Dialer{
	// 	Timeout: 30 * time.Second,
	// }
	auth := &proxy.Auth{
		User:     os.Getenv("PROXY_USER"),
		Password: os.Getenv("PROXY_PASS"),
	}
	dialer, err := proxy.SOCKS5("tcp", proxyStr, auth, proxy.Direct)
	if err != nil {
		logger.Warnf("Error dialing: %v\n", err)
		return nil, err
	}
	conn, err := dialer.Dial(network, addr)
	if err != nil {
		logger.Debugf("Error dialing: %v\n", err)
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
	Timeout:   120 * time.Second,
}

var NoRedirecth1Client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: loggingh1Transport,
	Timeout:   120 * time.Second,
}
