package modules

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/lormars/octohunter/asset"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

var errbased = `<%'${{/#{@}}%>{{`
var nonerrbased = []string{
	`p ">[[${{1}}]]`, `<%=1%>@*#{1}`, `{##}/*{{.}}*/`,
}

// var signatures = []string{
// 	`p ">[[$1]]`, `{##}/**/`, `p ">[[$]]`, `<a>p`, `p ">[[${1}]]`, `<p>">[[${{1}}]]</p>`,
// 	`1@*#{1}`, `<%=1%>@*1`, `<%=1%>`, `p ">1`, `&lt;%=1%&gt;@*#{1}`, `{##}`,
// }

func CheckSSTI(input *common.XssInput) {

	logger.Debugf("Checking SSTI for %s for param %s\n", input.Url, input.Param)

	parsedURL, err := url.Parse(input.Url)
	if err != nil {
		return
	}

	queries := parsedURL.Query()
	queries.Set(input.Param, errbased)
	parsedURL.RawQuery = queries.Encode()

	common.AddToCrawlMap(parsedURL.String(), "ssti", 200) //TODO: can be accurate
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		logger.Warnf("Error creating request: %v", err)
	}

	resp, err := checker.CheckServerCustom(req, clients.Clients.GetRandomClient("h0", false, true))
	if err != nil {
		return
	}

	if resp.StatusCode >= 500 {
		msg := fmt.Sprintf("[SSTI Errbased] %s in %s", input.Param, input.Url)
		if common.SendOutput {
			common.OutputP.PublishMessage(msg)
		}
		notify.SendMessage(msg)
		return
	}

	type result struct {
		index int
		body  string
		err   error
	}

	results := make(chan result, len(nonerrbased))
	var wg sync.WaitGroup

	copyQueries := func(original url.Values) url.Values {
		copy := make(url.Values)
		for k, vs := range original {
			for _, v := range vs {
				copy.Add(k, v)
			}
		}
		return copy
	}

	sstiSuspect := make(map[string][]string)
	for index, nonerr := range nonerrbased {
		wg.Add(1)
		go func(index int, nonerr string) {
			defer wg.Done()
			localQueries := copyQueries(queries)
			localQueries.Set(input.Param, nonerr)
			parsedURL.RawQuery = localQueries.Encode()

			req, err := http.NewRequest("GET", parsedURL.String(), nil)
			if err != nil {
				logger.Warnf("Error creating request: %v", err)
			}
			resp, err := checker.CheckServerCustom(req, clients.Clients.GetRandomClient("h0", false, true))
			if err != nil {
				results <- result{index, "", err}
			}
			results <- result{index: index, body: resp.Body, err: nil}

		}(index, nonerr)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
	for res := range results {
		if res.err != nil {
			return
		}
		if len(sstiSuspect) == 0 {
			for key, values := range asset.SSTIPoly {
				var toCheck string
				if values[res.index] == "Unmodified" {
					toCheck = nonerrbased[res.index]
				} else if values[res.index] != "Error" {
					toCheck = values[res.index]
				} else {
					continue
				}
				if strings.Contains(res.body, toCheck) {
					sstiSuspect[key] = values
				}
			}
		} else {
			for key, values := range sstiSuspect {
				var toCheck string
				if values[res.index] == "Unmodified" {
					toCheck = nonerrbased[res.index]
				} else if values[res.index] != "Error" {
					toCheck = values[res.index]
				} else {
					continue
				}
				if !strings.Contains(res.body, toCheck) {
					delete(sstiSuspect, key)
				}
			}
		}

	}
	if len(sstiSuspect) > 0 {
		var allSuspects string
		for key := range sstiSuspect {
			allSuspects += key + " or "
		}
		msg := fmt.Sprintf("[SSTI NonErrbased] %s in %s possibly using %s", input.Param, input.Url, allSuspects)
		if common.SendOutput {
			common.OutputP.PublishMessage(msg)
		}
		notify.SendMessage(msg)
		return
	}
}
