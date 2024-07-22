package comparer

import (
	"fmt"

	"github.com/lormars/octohunter/internal/getter"
)

func AreSiblingDomains(url1, url2 string) (bool, error) {

	rootDomain1, err := getter.GetDomain(url1)
	if err != nil {
		return false, fmt.Errorf("could not get root domain for URL1: %w", err)
	}

	rootDomain2, err := getter.GetDomain(url2)
	if err != nil {
		return false, fmt.Errorf("could not get root domain for URL2: %w", err)
	}

	return rootDomain1 == rootDomain2, nil
}
