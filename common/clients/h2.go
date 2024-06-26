package clients

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/lormars/octohunter/internal/logger"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
	"golang.org/x/net/proxy"
)

func customh2DialTLSContext(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
	proxyStr, _ := ctx.Value("proxy").(string)
	// if ok {
	// 	fmt.Println("h2 Using proxy: ", proxyStr)
	// } else {
	// 	fmt.Println("h2 No proxy")
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
