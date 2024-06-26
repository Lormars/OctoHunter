package parser

import (
	"net/http"
	"net/url"

	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/logger"
)

func ParseUrltoQueue(urlStr string, req *http.Request, client *http.Client) (chan []clients.Response, error) {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		logger.Debugf("Error parsing URL: %v\n", err)
		return nil, err
	}
	currentHost := parsedUrl.Hostname()
	respCh := clients.AddToQueue(currentHost, []*http.Request{req}, client)
	return respCh, nil

}
