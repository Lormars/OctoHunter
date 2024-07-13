package modules

import "github.com/lormars/octohunter/internal/cacher"

// Xss checkes for possible xss vulnerabilities in the given url
// It takes the url and a list of injection as input
// The list of injection must be in the form of [target param/header, param/header type]
// TODO: hold for now
func Xss(urlStr string, injection []string) {
	if len(injection) < 2 {
		return
	}

	for_cache := urlStr + injection[0] + injection[1]
	if !cacher.CheckCache(for_cache, "xss") {
		return
	}

}
