package fuzzer

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/notify"
)

var prefixes chan string
var suffixes chan string
var subdomains chan string
var tasks chan Fuzz3Part

func init() {
	apifuzzerinit()
}

func apifuzzerinit() {
	prefixes = make(chan string, 1000)
	suffixes = make(chan string, 1000)
	subdomains = make(chan string, 1000)
	tasks = make(chan Fuzz3Part, 1000)

	collectedPrexies := make(map[string]bool)
	collectedSuffixes := make(map[string]bool)
	collectedSubdomains := make(map[string]bool)

	var mu sync.Mutex

	for i := 0; i < 1000; i++ {
		go apiWorker(tasks)
	}
	for i := 0; i < 1000; i++ {
		go func() {
			for {
				select {
				case prefix := <-prefixes:
					mu.Lock()
					if !collectedPrexies[prefix] {
						collectedPrexies[prefix] = true
						for suffix := range collectedSuffixes {
							for subdomain := range collectedSubdomains {
								tasks <- Fuzz3Part{Part1: subdomain, Part2: prefix, Part3: suffix}
							}
						}
					}
					mu.Unlock()
				case suffix := <-suffixes:
					mu.Lock()
					if !collectedSuffixes[suffix] {
						collectedSuffixes[suffix] = true
						for prefix := range collectedPrexies {
							for subdomain := range collectedSubdomains {
								tasks <- Fuzz3Part{Part1: subdomain, Part2: prefix, Part3: suffix}
							}
						}
					}
					mu.Unlock()
				case subdomain := <-subdomains:
					mu.Lock()
					if !collectedSubdomains[subdomain] {
						collectedSubdomains[subdomain] = true
						for prefix := range collectedPrexies {
							for suffix := range collectedSuffixes {
								tasks <- Fuzz3Part{Part1: subdomain, Part2: prefix, Part3: suffix}
							}
						}
					}
					mu.Unlock()

				}
			}
		}()
	}
}

func apiWorker(tasks chan Fuzz3Part) {
	for task := range tasks {
		if !interesting(task.Part3) {
			continue
		}
		time.Sleep(100 * time.Millisecond)
		//I just find it hard to believe that any api endpoint would be in http...
		reconstructed := "https://" + task.Part1 + "/" + task.Part2 + "/" + task.Part3
		//check cache to avoid fuzz the original input api endpoint
		if !cacher.CheckCache(reconstructed, "fuzzapi") {
			continue
		}
		// logger.Warnf("[Fuzz API Debug] reconstructed is: %s", reconstructed)
		req, err := http.NewRequest("GET", reconstructed, nil)
		if err != nil {
			continue
		}
		resp, err := checker.CheckServerCustom(req, clients.NormalClient)
		if err != nil {
			continue
		}
		if resp.StatusCode == 404 || resp.StatusCode == 401 || resp.StatusCode == 403 {
			continue
		}
		//check content type to make sure we find new API endpoints
		contentType := resp.Headers.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			common.DividerP.PublishMessage(resp)
			continue
		}

		common.AddToCrawlMap(resp.Url, "fuzz", resp.StatusCode)
		msg := fmt.Sprintf("[Fuzz API] Found new endpoint: %s with SC %d", resp.Url, resp.StatusCode)
		if common.SendOutput {
			common.OutputP.PublishMessage(msg)
		}
		notify.SendMessage(msg)
		//if work, check path traversal first
		common.PathTraversalP.PublishMessage(reconstructed)

		if resp.Body != "" {
			common.QuirksP.PublishMessage(resp)
		}
	}
}

func FuzzAPI(urlStr string) {

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return
	}
	hostname := parsedURL.Hostname()
	fullPath := parsedURL.Path
	fileName := path.Base(fullPath)
	dirPath := path.Dir(fullPath)
	subdomains <- hostname

	//trim all leading or trailing / for clarity
	if fileName != "." && fileName != "/" && fileName != "" {
		fileName = strings.Trim(fileName, "/")
		suffixes <- fileName
	}

	if dirPath != "." && dirPath != "/" && dirPath != "" {
		dirPath = strings.Trim(dirPath, "/")
		prefixes <- dirPath
	}

}

func interesting(part3 string) bool {
	if !strings.HasSuffix(part3, ".js") && !strings.HasSuffix(part3, ".css") {
		return true
	}
	return false
}
