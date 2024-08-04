package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/lormars/octohunter/internal/logger"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
	"golang.org/x/net/proxy"
)

type H0Transport struct {
	h1Transport *http.Transport
	h2Transport *http2.Transport
}

// RoundTrip implements the RoundTripper interface
func (t *H0Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme == "https" {
		logger.Debugf("Attempting HTTP/2 for %s\n", req.URL.String())
		resp, err := t.h2Transport.RoundTrip(req)
		if err == nil {
			// logger.Warnf("HTTP/2 request succeeded for %s\n", req.URL.String())
			return resp, nil
		}
		// logger.Warnf("HTTP/2 request failed: %v, falling back to HTTP/1.1 for %s\n", err, req.URL.String())
	}
	return t.h1Transport.RoundTrip(req)
}

// createCombinedTransport creates a transport that supports both HTTP/2 and HTTP/1.1
func CreateH0Transport(proxy string) *H0Transport {
	h2Transport := CreateCustomh2Transport(proxy)
	h1Transport := CreateCustomh1Transport(proxy)

	return &H0Transport{
		h1Transport: h1Transport,
		h2Transport: h2Transport,
	}
}

func dial(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		logger.Debugf("Error dialing: %v\n", err)
		return nil, err
	}
	return conn, nil
}

func dialProxy(ctx context.Context, network, addr, proxyStr string) (net.Conn, error) {
	auth := &proxy.Auth{
		User:     os.Getenv("PROXY_USER"),
		Password: os.Getenv("PROXY_PASS"),
	}
	dialer, err := proxy.SOCKS5("tcp", proxyStr, auth, proxy.Direct)
	if err != nil {
		logger.Debugf("Error dialing: %v\n", err)
		return nil, err
	}

	conn, err := dialer.Dial(network, addr)
	if err != nil {
		logger.Debugf("Error dialing: %v\n", err)
		return nil, err
	}

	return conn, nil
}

func handshake(ctx context.Context, host, protocol string, conn net.Conn) (tlsConn net.Conn, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Debugf("Recovered from panic: %v\n", r)
			tlsConn = nil
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()
	tlsConn, err = dohandshake(ctx, host, protocol, conn)
	return
}

func dohandshake(ctx context.Context, host, protocol string, conn net.Conn) (net.Conn, error) {
	_, ok := ctx.Value("browser").(bool)
	if ok {
		logger.Debugf("browsering")
		config := &tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
			NextProtos: []string{protocol},
		}
		tlsConn := tls.Client(conn, config)
		err := tlsConn.Handshake()
		logger.Debugf("Handshake done\n")
		if err != nil {
			logger.Debugf("Error handshaking: %v\n", err)
			return nil, err
		}
		state := tlsConn.ConnectionState()
		logger.Debugf("Negotiated Protocol: %s", state.NegotiatedProtocol) // Log the negotiated protocol

		return tlsConn, nil
	} else {
		config := &utls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
			NextProtos: []string{protocol},
		}
		tlsConn := utls.UClient(conn, config, utls.HelloRandomizedALPN)
		err := tlsConn.Handshake()
		// logger.Infof("Handshake done\n")
		if err != nil {
			logger.Debugf("Error handshaking: %v\n", err)
			return nil, err
		}
		state := tlsConn.ConnectionState()
		logger.Debugf("Negotiated Protocol: %s", state.NegotiatedProtocol) // Log the negotiated protocol

		return tlsConn, nil
	}
}
