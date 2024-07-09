package fuzzer

import (
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
)

var prefixes chan string
var suffixes chan string
var subdomains chan string
var tasks chan Fuzz3Part

func init() {
	apifuzzerinit()
}

func apifuzzerinit() {
	prefixes = make(chan string, 100)
	suffixes = make(chan string, 100)
	subdomains = make(chan string, 100)
	tasks = make(chan Fuzz3Part, 100)

	collectedPrexies := make(map[string]bool)
	collectedSuffixes := make(map[string]bool)
	collectedSubdomains := make(map[string]bool)

	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		go apiWorker(tasks)
	}

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

func apiWorker(tasks chan Fuzz3Part) {
	for task := range tasks {
		//i just find it hard to believe that any api endpoint would be in http...
		reconstructed := "https://" + task.Part1 + "/" + task.Part2 + "/" + task.Part3
		//check cache to avoid fuzz the original input api endpoint
		if !cacher.CheckCache(reconstructed, "fuzzapi") {
			return
		}
		logger.Warnf("[Fuzz API Debug] reconstructed is: %s", reconstructed)

	}
}

func FuzzAPI(urlStr string) {
	if !cacher.CheckCache(urlStr, "fuzzapi") {
		return
	}
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
	if fileName != "." || fileName != "/" || fileName != "" {
		fileName = strings.Trim(fileName, "/")
		suffixes <- fileName
	}

	if dirPath != "." || dirPath != "/" || dirPath != "" {
		dirPath = strings.Trim(dirPath, "/")
		prefixes <- dirPath
	}

}
