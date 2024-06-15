package takeover

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/requester/pkg/runner"
)

var records []common.TakeoverRecord

var skip []string = []string{"incapdns", "ctripgslb", "gitlab", "impervadns", "sendgrid.net", "akamaiedge"}

func CNAMETakeover(options *common.Opts) {
	parseSignature("asset/fingerprints.json")

	if options.Target == "none" {
		multiplex.Conscan(takeover, options, 100)
	} else {
		takeover(options)
	}
}

func takeover(opts *common.Opts) {
	var domain string
	var cname string

	line := opts.Target
	parts := strings.Split(line, " ")
	if len(parts) >= 3 {
		domain = parts[0]
		cname = parts[2]
		cname = strings.Replace(cname, "[", "", -1)
		cname = strings.Replace(cname, "]", "", -1)
	}

	for _, s := range skip {
		if strings.Contains(cname, s) {
			//fmt.Println("skipped")
			return
		}
	}
	checkSig(domain, opts)

}

func checkSig(domain string, opts *common.Opts) bool {
	var dnsError error
	var temp_cname string

	temp_domain := domain
	for {
		temp_cname, dnsError = checker.FindImmediateCNAME(temp_domain)
		if temp_cname == "" || (dnsError != nil && !errors.Is(dnsError, common.ErrNXDOMAIN)) {
			//fmt.Printf("Oops, something bad happened on checkSig with temp_cname: %s and dnsError: %v", temp_cname, dnsError)
			return false
		}
		if temp_cname == temp_domain {
			break
		}

		temp_domain = temp_cname

	}
	//just for elb...
	if strings.Contains(domain, "tesla") {
		msg := fmt.Sprintf("DEBUG: ", temp_cname, domain)
		common.PublishMessage(msg)
	}
	if strings.Contains(temp_cname, "elb.") && strings.Contains(temp_cname, "amazonaws.com") {
		return false
	}
	for _, s := range skip {
		if strings.Contains(temp_cname, s) {
			return false
		}
	}

	if dnsError != nil {
		if errors.Is(dnsError, common.ErrNXDOMAIN) {
			for _, record := range records {
				if record.Nxdomain && record.Vulnerable {
					for _, sig := range record.Cname {
						if strings.Contains(temp_cname, sig) {
							msg := "[CNAME Confirmed] " + domain + " | Cname: " + temp_cname + " | Service: " + record.Service
							color.Red(msg)
							if opts.Broker {
								common.PublishMessage(msg)
							}
							return true
						}
					}
				}
			}
		}
		return false
	} else {
		protocols := []string{"http://", "https://"}
		for _, protocol := range protocols {
			url := protocol + domain
			config, err := runner.NewConfig(url)
			if err != nil {
				continue
			}
			resp, err := runner.Run(config)
			if err != nil {
				continue
			}
			for _, record := range records {

				for _, sig := range record.Cname {
					if strings.Contains(temp_cname, sig) {
						if !record.Vulnerable {
							return false
						} else {
							msg := "[CNAME Confirmed] " + domain + " | Cname: " + temp_cname + " | Service: " + record.Service
							color.Red(msg)
							if opts.Broker {
								common.PublishMessage(msg)
							}
							return true
						}
					}
				}

				if record.Vulnerable {

					if record.Fingerprint != "" && strings.Contains(resp.Body, record.Fingerprint) {
						msg := "[CNAME Potential] " + url + " | Fingerprint: " + record.Fingerprint + " | Service: " + record.Service
						color.Red(msg)
						if opts.Broker {
							common.PublishMessage(msg)
						}
						return true
					}

				}
			}
		}

		return false

	}

}

func parseSignature(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(byteValue, &records)
	if err != nil {
		fmt.Println(err)
		return
	}
}
