package modules

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/getter"
	"github.com/lormars/octohunter/internal/multiplex"
)

var payload []string = []string{"main", "admin", "dashboard", "user", "profile", "account", "settings", "portal", "home", "auth", "manage", "control", "panel", "secure", "access", "member", "myaccount", "private", "cpanel"}

func CheckRedirect(ctx context.Context, wg *sync.WaitGroup, opts *common.Opts) {
	defer wg.Done()
	if opts.Target != "none" {
		singleRedirectCheck(opts)
	} else {
		multiplex.Conscan(ctx, singleRedirectCheck, opts, opts.RedirectFile, "redirect", 10)
	}
}

func singleRedirectCheck(opts *common.Opts) {
	finalURL, err := getFinalURL(opts.Target)
	if err != nil {
		return
	}
	if strings.Contains(finalURL.Path, "login") {
		for _, p := range payload {
			testPath := strings.Replace(finalURL.Path, "login", p, -1)
			newUrl := finalURL.Scheme + "://" + finalURL.Host + testPath
			newFinalURL, err := getFinalURL(newUrl)
			if err != nil {
				continue
			}
			if newFinalURL.Path == finalURL.Path {
				length, err := getter.GetHeader(newUrl, "Content-Length")
				if err != nil {
					continue
				}
				var length_i int64
				_, err = fmt.Sscan(length, &length_i)
				if err != nil {
					continue
				}
				if length_i > 100 {
					msg := fmt.Sprintf("[Redirect] from %s to %s\n", newUrl, finalURL.String())
					color.Red(msg)
					if opts.Module.Contains("broker") {
						common.PublishMessage(msg)
					}
				}
			}
		}
	}
}

func getFinalURL(initialURL string) (*url.URL, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	resp, err := client.Get(initialURL)
	if err != nil {
		return nil, err
	}

	finalURL := resp.Request.URL
	if resp.Request.URL.String() != initialURL {
		finalURL, err = url.Parse(resp.Request.URL.String())
		if err != nil {
			return nil, err
		}
	}

	return finalURL, nil

}
