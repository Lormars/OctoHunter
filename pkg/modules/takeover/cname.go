package takeover

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/requester/pkg/runner"
)

var records []common.TakeoverRecord

var skip []string = []string{"incapdns", "ctripgslb", "gitlab", "impervadns", "sendgrid.net", "akamaiedge"}

func CNAMETakeover(ctx context.Context, wg *sync.WaitGroup, options *common.Opts) {
	defer wg.Done()
	parseSignature("asset/fingerprints.json")
	if options.Target == "none" {
		multiplex.Conscan(ctx, Takeover, options, options.CnameFile, "cname", 10)
	} else {
		Takeover(options)
	}
}

func Takeover(opts *common.Opts) {
	logger.Debugln("Takeover module running")
	domain := opts.Target
	hasCname, cname, _ := checker.HasCname(domain)
	if hasCname {
		for _, s := range skip {
			if strings.Contains(cname, s) {
				logger.Debugln("Skipping ", domain, " because of ", s, " is in skipping list.")
				return
			}
		}
		checkSig(domain, opts)
	}

}

func checkSig(domain string, opts *common.Opts) bool {
	var dnsError error
	var temp_cname string

	temp_domain := domain
	count := 0
	for {
		temp_cname, dnsError = checker.FindImmediateCNAME(temp_domain)
		if temp_cname == "" || (dnsError != nil && !errors.Is(dnsError, common.ErrNXDOMAIN)) {
			logger.Debugf("Oops, something bad happened on checkSig with temp_cname: %s and dnsError: %v\n", temp_cname, dnsError)
			return false
		}
		if temp_cname == temp_domain || count > 10 { //prevent circular cname
			logger.Debugf("Circular CNAME detected on %s\n", domain)
			break
		}

		temp_domain = temp_cname
		count++

	}
	//just for elb...
	if strings.Contains(temp_cname, "elb.") && strings.Contains(temp_cname, "amazonaws.com") {
		logger.Debugf("Skipping %s because it's an ELB\n", temp_cname)
		return false
	}
	for _, s := range skip {
		if strings.Contains(temp_cname, s) {
			logger.Debugf("Skipping %s because of %s is in skipping list.\n", temp_cname, s)
			return false
		}
	}
	//check if NXDomain
	if dnsError != nil {
		if errors.Is(dnsError, common.ErrNXDOMAIN) {
			for _, record := range records {
				if record.Nxdomain && record.Vulnerable {
					for _, sig := range record.Cname {
						if strings.Contains(temp_cname, sig) {
							msg := "[CNAME Confirmed] " + domain + " | Cname: " + temp_cname + " | Service: " + record.Service
							color.Red(msg)
							if opts.Module.Contains("broker") {
								common.OutputP.PublishMessage(msg)
							}
							return true
						}
					}
				}
			}
		}
		return false
		//if not NXDomain
	} else {
		protocols := []string{"http://", "https://"}
		for _, protocol := range protocols {
			url := protocol + domain
			config, err := runner.NewConfig(url)
			if err != nil {
				logger.Debugf("Error creating runner config for %s: %v\n", url, err)
				continue
			}
			resp, err := runner.Run(config)
			if err != nil {
				//when the status is noerror but there is no ip address...
				logger.Debugf("Error getting response from %s: %v\n", url, err)
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
							if opts.Module.Contains("broker") {
								common.OutputP.PublishMessage(msg)
							}
							return true
						}
					}
				}

				if record.Vulnerable {

					if record.Fingerprint != "" && strings.Contains(resp.Body, record.Fingerprint) {
						msg := "[CNAME Potential] " + url + " | Cname: " + temp_cname + " | Service: " + record.Service
						color.Red(msg)
						if opts.Module.Contains("broker") {
							common.OutputP.PublishMessage(msg)
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
		logger.Errorln("Error opening file: ", fileName)
		return
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		logger.Errorln("Error reading file: ", fileName)
		return
	}

	err = json.Unmarshal(byteValue, &records)
	if err != nil {
		logger.Errorln("Error parsing file", err)
		return
	}
}
