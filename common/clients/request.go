package clients

import (
	"io"
	"net/http"
)

type OctoRequest struct {
	Producer int
	Request  *http.Request
}

func NewRequest(method string, urlStr string, body io.Reader, ptype int) (*OctoRequest, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	return &OctoRequest{Producer: ptype, Request: req}, nil
}

const (
	Redirect = iota
	Method
	Hopper
	Divider
	Crawl
	Salesforce
	Splitting
	Cl0
	Quirks
	Rc
	Cors
	Pathconfuse
	Fuzz4034
	Pathtraversal
	Fuzzapi
	Fuzzunkeyed
	Xss
	Ssti
	Graphql
	Mime
	Fuzzpath
	Wayback
	In
	Misc
)

var ProducerWillExplode = make(map[int]bool)

func init() {
	ProducerWillExplode[Redirect] = false
	ProducerWillExplode[Method] = false
	ProducerWillExplode[Hopper] = false
	ProducerWillExplode[Divider] = false
	ProducerWillExplode[Crawl] = true
	ProducerWillExplode[Salesforce] = false
	ProducerWillExplode[Splitting] = true
	ProducerWillExplode[Cl0] = false
	ProducerWillExplode[Quirks] = true
	ProducerWillExplode[Rc] = true
	ProducerWillExplode[Cors] = false
	ProducerWillExplode[Pathconfuse] = true
	ProducerWillExplode[Fuzz4034] = true
	ProducerWillExplode[Pathtraversal] = true
	ProducerWillExplode[Fuzzapi] = true
	ProducerWillExplode[Fuzzunkeyed] = true
	ProducerWillExplode[Xss] = false
	ProducerWillExplode[Graphql] = true
	ProducerWillExplode[Mime] = false
	ProducerWillExplode[Fuzzpath] = true
	ProducerWillExplode[Wayback] = false
	ProducerWillExplode[In] = false
	ProducerWillExplode[Misc] = false

}
