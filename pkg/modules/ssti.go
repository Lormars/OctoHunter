package modules

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/lormars/octohunter/asset"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
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

	if !cacher.CheckCache(input.Url, "ssti") {
		return
	}

	common.AddToCrawlMap(input.Url, "ssti", 200) //TODO: can be accurate

	logger.Debugf("Checking SSTI for %s for param %s\n", input.Url, input.Param)

	parsedURL, err := url.Parse(input.Url)
	if err != nil {
		return
	}

	queries := parsedURL.Query()
	queries.Set(input.Param, errbased)
	parsedURL.RawQuery = queries.Encode()

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		logger.Warnf("Error creating request: %v", err)
	}

	resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
	if err != nil {
		return
	}

	if resp.StatusCode >= 500 {
		msg := fmt.Sprintf("[SSTI Errbased] %s in %s", input.Param, input.Url)
		if common.SendOutput {
			common.OutputP.PublishMessage(msg)
		}
		notify.SendMessage(msg)
	} else {
		sstiSuspect := make(map[string][]string)
		for index, nonerr := range nonerrbased {
			queries.Set(input.Param, nonerr)
			parsedURL.RawQuery = queries.Encode()

			req, err := http.NewRequest("GET", parsedURL.String(), nil)
			if err != nil {
				logger.Warnf("Error creating request: %v", err)
			}
			resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
			if err != nil {
				continue
			}

			if len(sstiSuspect) == 0 {
				if index != 0 {
					return
				}
				for key, values := range asset.SSTIPoly {
					var toCheck string
					if values[index] == "Unmodified" {
						toCheck = nonerr
					} else if values[index] != "Error" {
						toCheck = values[index]
					} else {
						continue
					}
					if strings.Contains(resp.Body, toCheck) {
						sstiSuspect[key] = values
					}
				}
			} else {
				for key, values := range sstiSuspect {
					var toCheck string
					if values[index] == "Unmodified" {
						toCheck = nonerr
					} else if values[index] != "Error" {
						toCheck = values[index]
					} else {
						delete(sstiSuspect, key)
						continue
					}
					if !strings.Contains(resp.Body, toCheck) {
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

}
