package modules

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/lormars/octohunter/asset"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/octohunter/internal/proxy"
)

func GoogleDork(ctx context.Context, wg *sync.WaitGroup, options *common.Opts) {
	defer wg.Done()
	if options.Target != "none" {
		singleDork(options)
	} else {
		multiplex.Conscan(ctx, singleDork, options, options.DorkFile, "dork", 1)
	}
}

func singleDork(options *common.Opts) {
	site := "site:" + options.Target
	for _, dork := range asset.DorkQueries {

		randomDuration := time.Duration(rand.Intn(5)+1) * time.Second
		time.Sleep(randomDuration)

		dork = url.QueryEscape(dork)
		query := dork + "+" + site
		url := "https://www.google.com/search?q=" + query
		//fmt.Print(url + "\n")
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		randomIndex := rand.Intn(len(asset.Useragent))
		randomAgent := asset.Useragent[randomIndex]
		req.Header.Set("User-Agent", randomAgent)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		resp.Body.Close()
		if strings.Contains(string(body), "Our systems have detected unusual traffic from your computer network.") {
			fmt.Println("Google detected unusual traffic, try aws api gateway")
			var bypassed bool
			for i := 0; i < 3; i++ {
				bypassed, body = proxy.AwsProxy(query)
				if bypassed {
					fmt.Println("Bypassed Google Captcha with AWS API Gateway")
					break
				}
			}

			if !bypassed {
				time.Sleep(1 * time.Minute)
				continue
			}
		}
		pattern := fmt.Sprintf(`(http|https)://[a-zA-Z0-9./?=_~-]*%s/[a-zA-Z0-9./?=_~-]*`, regexp.QuoteMeta(options.Target))
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(string(body), -1)
		for _, match := range matches {
			fmt.Println(match)
			msg := fmt.Sprintf("[Dork]: %s, Match: %s", query, match)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
		}

	}
}
