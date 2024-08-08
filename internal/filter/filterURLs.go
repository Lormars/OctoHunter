package filter

import (
	"math/rand"
	"net/url"
	"regexp"
	"strings"
)

type urlGroup struct {
	urls []string
}

func GroupAndFilterURLs(urls []string) map[string]string {
	groups := groupURLS(urls)
	filteredURLs := filterURLs(groups)
	return filteredURLs
}

func groupURLS(urls []string) map[string]*urlGroup {
	groups := make(map[string]*urlGroup)
	for _, rawURL := range urls {
		u, err := url.Parse(rawURL)
		if err != nil {
			continue
		}

		structure := getURLStructure(u)
		if group, exists := groups[structure]; exists {
			group.urls = append(group.urls, rawURL)
		} else {
			groups[structure] = &urlGroup{
				urls: []string{rawURL},
			}
		}
	}

	return groups
}

func getURLStructure(u *url.URL) string {
	path := u.Path

	path = replaceNumericSegments(path)

	queryParams := u.Query()
	queryParamNames := make([]string, 0, len(queryParams))
	for param := range queryParams {
		queryParamNames = append(queryParamNames, param)
	}
	structure := u.Hostname() + path + "?" + strings.Join(queryParamNames, "&")
	return structure
}

func replaceNumericSegments(path string) string {
	re := regexp.MustCompile(`\d+`)
	return re.ReplaceAllString(path, "{num}")
}

func filterURLs(groups map[string]*urlGroup) map[string]string {
	filteredURLs := make(map[string]string)
	for structure, group := range groups {
		randomIndex := rand.Intn(len(group.urls))
		filteredURLs[structure] = group.urls[randomIndex]
	}
	return filteredURLs
}
