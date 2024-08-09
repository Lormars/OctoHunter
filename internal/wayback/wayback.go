package wayback

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
)

// modified from tomnomnom's waybackurls

func GetWaybackURLs(domain string) {

	// logger.Warnf("Waybackurls for %s", domain)

	startTime := time.Now()
	var wg sync.WaitGroup
	wg.Add(2)

	allURLs := make(map[string]bool)

	go func() {
		defer wg.Done()
		urls, err := getWaybackURLs(domain)
		if err != nil {
			return
		}

		for _, u := range urls {
			allURLs[u] = true
		}
	}()

	go func() {
		defer wg.Done()
		urls, err := getCommonCrawlURLs(domain)
		if err != nil {
			return
		}

		for _, u := range urls {
			allURLs[u] = true
		}
	}()

	wg.Wait()

	semaphore := make(chan struct{}, 10)
	// fmt.Println(allURLs)
	for u := range allURLs {
		if !cacher.CanScan(u, "divider") || !cacher.CheckCache(u, "wayback") {
			continue
		}
		wg.Add(1)
		semaphore <- struct{}{}
		go func(u string) {
			defer wg.Done()
			defer func() {
				<-semaphore
			}()
			req, err := clients.NewRequest("GET", u, nil, clients.Wayback)
			if err != nil {
				logger.Warnf("Error creating request: %v", err)
				return
			}
			resp, err := checker.CheckServerCustom(req, clients.Clients.GetRandomClient("h0", true, true))
			if err != nil {
				logger.Warnf("Error getting response from %s: %v", u, err)
				return
			}
			// logger.Warnf("[Wayback Debug] %s - %d", u, resp.StatusCode)
			common.AddToCrawlMap(u, "wayback", resp.StatusCode)
			common.DividerP.PublishMessage(resp)
		}(u)
	}
	wg.Wait()

	elapsed := time.Since(startTime)
	if elapsed < 4*time.Second {
		time.Sleep(4*time.Second - elapsed)
	}
}

func getWaybackURLs(domain string) ([]string, error) {
	subsWildcard := "*."

	res, err := http.Get(
		fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=%s%s/*&output=json&collapse=urlkey", subsWildcard, domain),
	)
	if err != nil {
		logger.Warnf("Error getting waybackurls: %v", err)
		return nil, err
	}

	raw, err := io.ReadAll(res.Body)

	res.Body.Close()
	if err != nil {
		logger.Warnf("Error reading waybackurls: %v", err)
		return nil, err
	}

	var wrapper [][]string
	err = json.Unmarshal(raw, &wrapper)
	if err != nil {
		logger.Warnf("Error unmarshalling waybackurls: %v", err)
		return nil, err
	}

	out := make([]string, 0, len(wrapper))

	skip := true
	for _, urls := range wrapper {
		// The first item is always just the string "original",
		// so we should skip the first item
		if skip {
			skip = false
			continue
		}

		out = append(out, urls[2])
	}

	return out, nil

}

func getCommonCrawlURLs(domain string) ([]string, error) {
	subsWildcard := "*."

	res, err := http.Get(
		fmt.Sprintf("http://index.commoncrawl.org/CC-MAIN-2018-22-index?url=%s%s/*&output=json", subsWildcard, domain),
	)
	if err != nil {
		logger.Warnf("Error getting commoncrawl urls: %v", err)
		return nil, err
	}

	defer res.Body.Close()
	sc := bufio.NewScanner(res.Body)

	out := make([]string, 0)

	for sc.Scan() {

		wrapper := struct {
			URL       string `json:"url"`
			Timestamp string `json:"timestamp"`
		}{}
		err = json.Unmarshal([]byte(sc.Text()), &wrapper)

		if err != nil {
			logger.Warnf("Error unmarshalling commoncrawl urls: %v", err)
			continue
		}

		out = append(out, wrapper.URL)
	}

	return out, nil

}
