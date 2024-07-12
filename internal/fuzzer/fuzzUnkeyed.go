package fuzzer

import (
	"bufio"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/generator"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/matcher"
)

var prefix = "vbwpzub"

func init() {
	file, err := os.Open("list/unkeyedParam")
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

	file2, err := os.Open("list/unkeyedHeader")
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

func FuzzUnkeyed(urlStr string) {
	if !cacher.CheckCache(urlStr, "unkeyed") {
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
				var sigMap = make(map[string]string)
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
						sigMap[prefix+signature] = UnkeyedParam[paramIndex]
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
				req.Header.Set(header, prefix+signature)
				sigMap[prefix+signature] = header

				logger.Warnf("[DEBUG] Checking %s", parsedURL.String())
				logger.Warnf("[DEBUG] Headers: %v", req.Header)

				resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
				if err != nil {
					continue
				}
				for sig, param := range sigMap {
					if strings.Contains(resp.Body, sig) || matcher.HeaderValueContainsSignature(resp, sig) {
						logger.Warnf("Unkeyed parameter found: %s on %s", param, urlStr)
					} else if strings.Contains(resp.Body, prefix) || matcher.HeaderValueContainsSignature(resp, prefix) {
						logger.Warnf("Unkeyed prefix found on %s", urlStr)
					}

				}

			}
		}()
	}
}
