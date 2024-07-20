package fuzzer

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/generator"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/matcher"
	"github.com/lormars/octohunter/internal/notify"
	"github.com/lormars/octohunter/internal/parser"
)

var prefix = "vbwpzub"

func init() {
	file, err := os.Open("asset/unkeyedParam")
	if err != nil {
		panic("Error opening file")
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		param := scanner.Text()
		if strings.TrimSpace(param) != "" {
			UnkeyedParam = append(UnkeyedParam, param)
		}
	}
	logger.Infof("UnkeyedParam Wordlist loaded")

	file.Close()

	file2, err := os.Open("asset/unkeyedHeader")
	if err != nil {
		panic("Error opening file")
	}
	defer file2.Close()
	scanner2 := bufio.NewScanner(file2)
	for scanner2.Scan() {
		header := scanner2.Text()
		if strings.TrimSpace(header) != "" {
			UnkeyedHeader = append(UnkeyedHeader, header)
		}
	}
	logger.Infof("UnkeyedHeader Wordlist loaded")
}

// It fuzzes for parameters and headers that are reflected in the response (either in header or body) for cacheable pages.
func FuzzUnkeyed(urlStr string) {
	if !cacher.CheckCache(urlStr, "unkeyed") {
		return
	}

	//check if the page is cacheable
	//cannot do it here anymore since we are also checking for ssti
	// cacheable := checker.CheckCacheable(urlStr)

	paramLength := len(UnkeyedParam)
	headerLength := len(UnkeyedHeader)
	var mu sync.Mutex
	paramIndex := 0
	headerIndex := 0
	foundLocations := make(map[string]bool)

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return
	}

	found := 0
	for i := 0; i < 5; i++ {
		go func() {

			for {
				mu.Lock()
				if paramIndex >= paramLength && headerIndex >= headerLength {
					mu.Unlock()
					break
				}

				parsedURL := *parsed
				var header string
				var sigMap = make(map[string][]string)
				//get next five params
				for j := 0; j < 5; j++ {
					if paramIndex < paramLength {
						signature, err := generator.GenerateSignature()
						if err != nil {
							logger.Errorf("Error generating signature: %v\n", err)
							mu.Unlock()
							return
						}
						queryParams := parsedURL.Query()
						queryParams.Set(UnkeyedParam[paramIndex], prefix+signature)
						parsedURL.RawQuery = queryParams.Encode()
						sigMap[prefix+signature] = []string{UnkeyedParam[paramIndex], "param"}
						paramIndex++
					}
				}

				//get next 1 header
				if headerIndex < headerLength {
					header = UnkeyedHeader[headerIndex]
					headerIndex++
				}

				//reset found to 0 if it is not disabled. (meaning if it is not already reported that there is a persistent parameter issue)
				found = updateFound(found, 0)

				mu.Unlock()
				req, err := http.NewRequest("GET", parsedURL.String(), nil)
				if err != nil {
					logger.Warnf("Error creating request: %v", err)
					continue
				}
				signature, err := generator.GenerateSignature()
				if err != nil {
					logger.Errorf("Error generating signature: %v\n", err)
					return
				}

				specialHeader := false
				if header != "" {
					if strings.Contains(header, "~") {
						specialHeader = true
						parts := strings.Split(header, "~")
						value := fmt.Sprintf(parts[1], prefix+signature+".com")
						req.Header.Set(parts[0], value)
						sigMap[prefix+signature+".com"] = []string{header, "header"}
						// logger.Warnf("[DEBUG] Special header: %s with value %s", parts[0], value)
					} else {
						req.Header.Set(header, prefix+signature)
						sigMap[prefix+signature] = []string{header, "header"}
					}
				}

				logger.Debugf("[DEBUG] Checking %s", parsedURL.String())
				logger.Debugf("[DEBUG] Headers: %v", req.Header)

				resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
				if err != nil {
					continue
				}

				if specialHeader {
					if resp.StatusCode >= 300 {
						msg := fmt.Sprintf("[Fuzz Unkeyed] Special header found: %s on %s with sc %d", header, urlStr, resp.StatusCode)
						if common.SendOutput {
							common.OutputP.PublishMessage(msg)
						}
						notify.SendMessage(msg)
					}
				}

				for sig, param := range sigMap {
					//if value is reflected in response body
					if strings.Contains(resp.Body, sig) {
						mu.Lock()
						found = updateFound(found, 1)
						mu.Unlock()
						if param[1] == "header" {
							//check if this header is unkeyed
							req.Header.Del(header)
							resp, err = checker.CheckServerCustom(req, clients.NoRedirectClient)
							if err != nil {
								continue
							}
							//if the signature is still there, then it is unkeyed and cached
							//can directly report it, as it is already interesting to have a header reflected in the body
							if strings.Contains(resp.Body, sig) {
								msg := fmt.Sprintf("[Fuzz Unkeyed] Unkeyed header found: %s on %s", param[0], urlStr)
								if common.SendOutput {
									common.OutputP.PublishMessage(msg)
								}
								notify.SendMessage(msg)
							}
						} else if param[1] == "param" {

							//the location is either "attribute", "tag", or "both".
							//found is map of string and bool, and string represents the tag/attribute the signature is found in
							inBody, location, found := parser.ExtractSignature(resp.Body, sig)
							mu.Lock()
							//check again if the signature is reflected in body, and
							//if the location (attr/tag) is not already found by other params
							//this is necessary to avoid duplicate reports of params reflected in the same location
							if inBody && !common.IsSuperset(foundLocations, found) {
								logger.Debugf("[Fuzz Unkeyed Debug] parameter found: %s on %s on %s", param[0], urlStr, location)
								common.MergeMaps(foundLocations, found)
								mu.Unlock()
								xssInput := &common.XssInput{
									Url:      urlStr,
									Param:    param[0],
									Location: location,
								}
								// if cacheable {
								//publish the xssInput to the broker
								common.XssP.PublishMessage(xssInput) //this only checks for possible xss. It has nothing to do checking if the param itself is unkeyed though.
								// }

								//ssti check
								sstiInput := &common.XssInput{
									Url:   urlStr,
									Param: param[0],
								}
								common.SstiP.PublishMessage(sstiInput)

							} else {
								mu.Unlock() //a little ugly
							}

						}

						//if value is reflected in response header
					} else if matcher.HeaderValueContainsSignature(resp, sig) {
						mu.Lock()
						found = updateFound(found, 1)
						mu.Unlock()
						if param[1] == "param" {
							//if param is in header value, then can check if CRLF injection is possible
							common.SplittingP.PublishMessage(resp)
						}
					}
				}
				mu.Lock()
				notfound := found == 0
				mu.Unlock()
				if notfound {
					//this is interesting, as it means that the exact signature is not found, but the prefix is found,
					//which means that the a signature from previous requests is cached
					if strings.Contains(resp.Body, prefix) {
						hostname := parsed.Hostname()
						if !cacher.CheckCache(hostname, "unkeyedPrefix") {
							mu.Lock()
							found = -1
							mu.Unlock()
							continue
						}
						msg := fmt.Sprintf("[Fuzz Unkeyed] Prefix %s found on %s", prefix, urlStr)
						if common.SendOutput {
							common.OutputP.PublishMessage(msg)
						}
						notify.SendMessage(msg)
						mu.Lock()
						found = -1 //just disable it in current url to prevent information flood
						mu.Unlock()
					}
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}
}

func updateFound(source, target int) int {
	if source != -1 {
		source = target
	}
	return source
}
