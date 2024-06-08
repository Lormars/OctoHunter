package modules

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"

	"github.com/lormars/octohunter/asset"
	"github.com/lormars/octohunter/common"
)

func GoogleDork(options *common.Opts) {
	if options.Target != "none" {
		singleDork(options)
	} else {
		multiDork(options)
	}
}

func singleDork(options *common.Opts) {
	site := "site:" + options.Target
	for _, dork := range asset.DorkQueries {
		dork = url.QueryEscape(dork)
		url := "https://www.google.com/search?q=" + dork + "+" + site
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
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println(string(body))
		//TODO: regex domain and ban
		pattern := fmt.Sprintf(`(http|https)://[a-zA-Z0-9./?=_~-]*%s/[a-zA-Z0-9./?=_~-]*`, regexp.QuoteMeta(options.Target))
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(string(body), -1)
		for _, match := range matches {
			fmt.Println(match)
		}

	}
}

func multiDork(options *common.Opts) {

}
