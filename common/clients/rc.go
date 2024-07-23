package clients

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"time"

	rchttp2 "github.com/lormars/http2/http2"
	"github.com/lormars/octohunter/internal/logger"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/proxy"
)

func rch2DialTLSContext(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
	var conn net.Conn
	var err error
	if UseProxy {
		proxyStr, _ := ctx.Value("proxy").(string)
		auth := &proxy.Auth{
			User:     os.Getenv("PROXY_USER"),
			Password: os.Getenv("PROXY_PASS"),
		}
		dialer, err := proxy.SOCKS5("tcp", proxyStr, auth, proxy.Direct)
		if err != nil {
			logger.Warnf("Error dialing: %v\n", err)
			return nil, err
		}

		conn, err = dialer.Dial(network, addr)
		if err != nil {
			logger.Debugf("Error dialing: %v\n", err)
			return nil, err
		}
	} else {
		dialer := &net.Dialer{
			Timeout: 30 * time.Second,
		}
		conn, err = dialer.DialContext(ctx, network, addr)
		if err != nil {
			logger.Debugf("Error dialing: %v\n", err)
			return nil, err
		}
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

	if err := conn.(*net.TCPConn).SetNoDelay(false); err != nil {
		logger.Errorf("Error setting TCP_NODELAY: %v\n", err)
		return nil, err
	}

	tlsConn := utls.UClient(conn, config, utls.HelloRandomizedALPN)
	err = tlsConn.Handshake()
	if err != nil {
		logger.Debugf("Error handshaking: %v\n", err)
		return nil, err
	}
	return tlsConn, nil
}

// Custom transport using utls for TLS fingerprinting
func CreateRCh2Transport() (*rchttp2.Transport, error) {
	transport := &rchttp2.Transport{
		DialTLSContext: rch2DialTLSContext,
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
