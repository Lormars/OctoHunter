package quirks

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/generator"
	"github.com/lormars/octohunter/internal/matcher"
	"github.com/lormars/octohunter/internal/notify"
	"github.com/lormars/octohunter/internal/parser"
)

//Quirks is a general scanner that scan for intersting http responses.
//It does not mean that the responses are vulnerabilities, but they are interesting.

var result *common.ServerResult

func CheckQuirks(res *common.ServerResult) {
	//there are just so many websites with the same quirks on all the endpoints under a path,
	//so need to cache a little more agressively to cache the first path as well
	// firstPath, err := cacher.GetFirstPath(res.Url)
	// if err != nil {
	// 	if !cacher.CheckCache(res.Url, "quirks") {
	// 		return
	// 	}
	// } else {
	// 	if !cacher.CheckCache(firstPath, "quirks") {
	// 		return
	// 	}
	// }

	// logger.Warnf("Checking quirks for %s\n", res.Url)

	if !cacher.CheckCache(res.Url, "quirks") {
		return
	}

	result = res

	if strings.Contains(result.Url, "/_next/image") {
		msg := fmt.Sprintf("[Quirks] Next.js Image URL in %s", result.Url)
		common.OutputP.PublishMessage(msg)
		notify.SendMessage(msg)
	}

	if strings.HasSuffix(result.Url, ".css") ||
		strings.HasSuffix(result.Url, ".png") ||
		strings.HasSuffix(result.Url, ".jpg") ||
		strings.HasSuffix(result.Url, ".jpeg") ||
		strings.HasSuffix(result.Url, ".gif") ||
		strings.HasSuffix(result.Url, ".svg") ||
		strings.HasSuffix(result.Url, ".ico") {
		return
	}

	//secret miner
	pattern := `(?i)(?:key|api|token|secret|client|passwd|password|auth|access)(?:[0-9a-z\\-_\\t .]{0,20})(?:[\\s|']|[\\s|\"]){0,3}(?:=|>|:{1,3}=|\\|\\|:|<=|=>|:|\\?=)(?:'|\"|\\s|=|\\x60){0,5}([0-9a-z\\-_.=]{10,150})(?:['|\"|\\n|\\r|\\s|\\x60|;]|$)`
	re := regexp.MustCompile(pattern)
	match := re.FindString(result.Body)
	if match != "" {
		msg := fmt.Sprintf("[Quirks] Secret Miner in %s: %s", result.Url, match)
		common.OutputP.PublishMessage(msg)
		notify.SendMessage(msg)
		color.Red(msg)
	}

	//dependency confusion check
	if strings.Contains(result.Body, "package.json") ||
		strings.Contains(result.Body, "requirements.txt") ||
		strings.Contains(result.Body, "Gemfile") ||
		strings.Contains(result.Body, "composer.json") ||
		strings.Contains(result.Url, "package.json") ||
		strings.Contains(result.Url, "requirements.txt") ||
		strings.Contains(result.Url, "Gemfile") ||
		strings.Contains(result.Url, "composer.json") {
		msg := fmt.Sprintf("[Quirks] Dependency Confusion in %s", result.Url)
		common.OutputP.PublishMessage(msg)
		notify.SendMessage(msg)
	}

	//oauth check
	if strings.Contains(result.Url, "client_id") &&
		strings.Contains(result.Url, "redirect_uri") &&
		strings.Contains(result.Url, "response_type") &&
		!strings.Contains(result.Url, "state") {
		msg := fmt.Sprintf("[Quirks] OAuth in URL %s without state parameter", result.Url)
		common.OutputP.PublishMessage(msg)
		notify.SendMessage(msg)
	}

	//no need to wait for this, takes too long, just fire and forget
	if strings.HasSuffix(result.Url, ".js") {
		go CheckJSQuirks(result)
		return
	}

	//too much false positive
	//doubleHTML()

	go checkJSONP()

	go jsonwithHTML()

	// go func() {
	// 	leakenv()
	// }()

	go isdynamic()

	if checker.CheckAccess(result) {
		if result.Body != "" {
			go bodyreflected()
			go common.FuzzUnkeyedP.PublishMessage(result.Url)
		}
	}

}

func bodyreflected() {
	//check whether any param value is reflected in body
	parsedURL, err := url.Parse(result.Url)
	if err != nil {
		return
	}

	params := parsedURL.Query()

	for param, values := range params {
		for _, value := range values {
			if len(value) >= 4 && strings.Contains(result.Body, value) {
				//ssti
				sstiInput := &common.XssInput{
					Url:   result.Url,
					Param: param,
				}
				common.SstiP.PublishMessage(sstiInput)
				//xss
				inBody, location, _ := parser.ExtractSignature(result.Body, value)
				if inBody {
					xssInput := &common.XssInput{
						Url:      result.Url,
						Param:    param,
						Location: location,
					}
					common.XssP.PublishMessage(xssInput)
				}

			}
		}
	}
}

func checkJSONP() {
	parsedURL, err := url.Parse(result.Url)
	if err != nil {
		return
	}
	params, err := url.ParseQuery(parsedURL.RawQuery)
	if err != nil || len(params) == 0 {
		return
	}

	paramsRegex := regexp.MustCompile(`^[a-zA-Z][.\w]{4,}$`)
	start := `(?:^|[^\w'".])`
	end := `\s*[(]`

	for _, values := range params {
		for _, value := range values {
			if paramsRegex.MatchString(value) {
				callbackRegex := regexp.MustCompile(fmt.Sprintf("%s%s%s", start, regexp.QuoteMeta(value), end))
				match := callbackRegex.FindString(result.Body)
				if match != "" {
					msg := "[JSONP Suspect] " + match + " in " + result.Url
					common.OutputP.PublishMessage(msg)
					notify.SendMessage(msg)
					color.Red(msg)
				}
			}
		}
	}

}

func doubleHTML() {
	contentType := result.Headers.Get("Content-Type")
	if contentType == "" {
		return
	}
	if !checker.CheckMimeType(contentType, "text/html") {
		return
	}
	if strings.Count(result.Body, "</html>") > 1 {
		//if result.Depth > 0, it means this url is the result of a crawl
		//then it is worthy to crawl it to get further endpoint
		//if result.Depth = 0, then it must already be crawled by crawler, so no need to crawl it again
		if result.Depth > 0 {
			common.CrawlP.PublishMessage(result)
		}

		msg := fmt.Sprintf("[Quirks] Double HTML in %s", result.Url)
		common.OutputP.PublishMessage(msg)
		notify.SendMessage(msg)
	}
}

func jsonwithHTML() {
	contentType := result.Headers.Get("Content-Type")
	if contentType == "" {
		return
	}
	if !checker.CheckMimeType(contentType, "text/html") {
		return
	}
	if strings.HasPrefix(result.Body, "{") || strings.HasPrefix(result.Body, "[") {
		msg := fmt.Sprintf("[Quirks] JSON with HTML mime in %s", result.Url)
		common.OutputP.PublishMessage(msg)
		notify.SendMessage(msg)
	}
}

func leakenv() {
	if strings.Count(result.Body, "HTTP_") > 2 {
		msg := fmt.Sprintf("[Quirks] HTTP_ ENV leak in %s", result.Url)
		common.OutputP.PublishMessage(msg)
		notify.SendMessage(msg)
	}
}

func isdynamic() {
	if !checker.CheckAccess(result) {
		return
	}
	cacheBuster, err := generator.GenerateSignature()
	if err != nil {
		return
	}

	parsedURL, err := url.Parse(result.Url)
	if err != nil {
		return
	}
	queries := parsedURL.Query()
	queries.Set("cachebuster", cacheBuster)
	parsedURL.RawQuery = queries.Encode()

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return
	}
	resp1, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
	if err != nil {
		return
	}
	resp2, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
	if err != nil {
		return
	}

	//first check if it is dynamically generated by comparing two responses
	same := resp1.Body == resp2.Body
	//then pass to path confusion if there is no cache header or no cache hit
	if !same && (!matcher.HeaderKeyContainsSignature(resp2, "cache") || !matcher.HeaderValueContainsSignature(resp2, "hit")) {
		common.PathConfuseP.PublishMessage(result.Url)
	}

}
