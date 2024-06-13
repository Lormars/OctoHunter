package takeover

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/requester/pkg/runner"
)

var records []common.TakeoverRecord

func CNAMETakeover(options *common.Opts) {
	parseSignature("list/fingerprints.json")
	if options.Target == "none" {
		multiplex.Conscan(takeover, options, 50)
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

	skip := []string{"incapdns", "ctripgslb", "gitlab", "impervadns", "elb.amazonaws"}
	for _, s := range skip {
		if strings.Contains(cname, s) {
			fmt.Println("skipped")
			return
		}
	}
	checkSig(domain, cname, opts)

}

func checkNXDomain(cname string) bool {
	_, err := net.LookupHost(cname)
	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.Err == "no such host" {
			return true
		}
	}
	return false
}

func checkSig(domain, cname string, opts *common.Opts) bool {
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
			if record.Vulnerable {

				if record.Fingerprint != "" && strings.Contains(resp.Body, record.Fingerprint) {
					msg := "[CNAME] " + url + " | Fingerprint: " + record.Fingerprint + " | Service: " + record.Service
					color.Red(msg)
					if opts.Broker {
						common.PublishMessage(msg)
					}
					return true
				}
				if record.Nxdomain && record.Fingerprint == "NXDOMAIN" && checkNXDomain(cname) {
					msg := "[CNAME] " + url + " | Fingerprint: " + record.Fingerprint + " | Service: " + record.Service
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
