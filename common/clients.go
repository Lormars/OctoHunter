package common

import (
	"crypto/tls"
	"net/http"
	"time"
)

var NormalClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	},
	Timeout: 10 * time.Second,
}

var NoRedirectClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Timeout: 10 * time.Second,
}

// disable auto upgrade to http2, force http/1.1
var transport = &http.Transport{
	TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
}

var NoRedirectHTTP1Client = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: transport,
	Timeout:   10 * time.Second,
}
