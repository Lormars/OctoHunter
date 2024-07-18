package common

import (
	"sync"

	cmap "github.com/lormars/crawlmap/pkg"
)

var crawlmu sync.Mutex

func AddToCrawlMap(urlStr, origin string, statusCode int) {
	cmapInput := &cmap.NodeInput{
		Url:        urlStr,
		StatusCode: statusCode,
		Origin:     origin,
	}
	crawlmu.Lock()
	cmap.AddNode(cmapInput)
	crawlmu.Unlock()
}

func GetOriginMap() map[string][]string {
	originMap := cmap.ReturnOrigin()
	return originMap
}
