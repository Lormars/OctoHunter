package fuzzer

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

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
	if !checker.CheckCacheable(urlStr) {
		return
	}

	paramLength := len(UnkeyedParam)
	headerLength := len(UnkeyedHeader)
	var mu sync.Mutex
	paramIndex := 0
	headerIndex := 0

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return
	}
	for i := 0; i < 10; i++ {
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
				if header != "" {
					req.Header.Set(header, prefix+signature)
					sigMap[prefix+signature] = []string{header, "header"}
				}

				logger.Debugf("[DEBUG] Checking %s", parsedURL.String())
				logger.Warnf("[DEBUG] Headers: %v", req.Header)

				resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
				if err != nil {
					continue
				}
				for sig, param := range sigMap {

					if strings.Contains(resp.Body, sig) {
						if param[1] == "header" {
							//check if this header is unkeyed
							req.Header.Del(header)
							resp, err = checker.CheckServerCustom(req, clients.NoRedirectClient)
							if err != nil {
								continue
							}
							//if the signature is still there, then it is unkeyed and cached
							if strings.Contains(resp.Body, sig) {
								msg := fmt.Sprintf("[Fuzz Unkeyed] Unkeyed header found: %s on %s", param[0], urlStr)
								common.OutputP.PublishMessage(msg)
								notify.SendMessage(msg)
							}
						} else if param[1] == "param" {
							inBody, location := parser.ExtractSignature(resp.Body, sig)
							if inBody {
								logger.Warnf("[Fuzz Unkeyed Debug] parameter found: %s on %s on %s", param[0], urlStr, location)
							}
						}
					} else if matcher.HeaderValueContainsSignature(resp, sig) {
						if param[1] == "param" {
							//if param is in header value, then can check if CRLF injection is possible
							common.SplittingP.PublishMessage(resp)
						}
					}

					// if strings.Contains(resp.Body, sig) || matcher.HeaderValueContainsSignature(resp, sig) {
					// 	logger.Warnf("Unkeyed parameter found: %s on %s", param[0], urlStr)
					// }

				}
				// if strings.Contains(resp.Body, prefix) || matcher.HeaderValueContainsSignature(resp, prefix) {
				// 	msg := fmt.Sprintf("[Unkeyed] Unkeyed prefix found on %s", urlStr)
				// }

			}
		}()
	}
}
