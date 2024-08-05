package takeover

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
)

var records []common.TakeoverRecord
var skip []string = []string{"incapdns", "ctripgslb", "gitlab", "impervadns", "sendgrid.net", "akamaiedge"}

func init() {
	parseSignature("asset/fingerprints.json")
}

func Takeover(domainStr string) {

	// logger.Debugln("Takeover module running")
	domain := domainStr
	hasCname, cname, _ := checker.HasCname(domain)
	if hasCname {
		for _, s := range skip {
			if strings.Contains(cname, s) {
				logger.Debugln("Skipping ", domain, " because of ", s, " is in skipping list.")
				return
			}
		}
		checkSig(domain)
	}

}

func checkSig(domain string) bool {
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
			// logger.Debugf("Circular CNAME detected on %s\n", domain)
			break
		}
		//salesforce site, check salesforce
		if strings.Contains(temp_cname, ".force.com") || strings.Contains(temp_cname, ".siteforce.com") {
			common.SalesforceP.PublishMessage(temp_cname)
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
	for _, record := range records {
		if record.Vulnerable {
			for _, sig := range record.Cname {
				if strings.Contains(temp_cname, sig) {
					msg := "[CNAME Confirmed] " + domain + " | Cname: " + temp_cname + " | Service: " + record.Service
					if common.SendOutput {
						common.OutputP.PublishMessage(msg)
					}
					return true
				}
			}

			//comment out due to high false positive
			// var serverResults []*common.ServerResult
			// httpResult, httpsResult, errHttp, errHttps := checker.CheckHTTPAndHTTPSServers(temp_cname)
			// if errHttp == nil {
			// 	serverResults = append(serverResults, httpResult)
			// }
			// if errHttps == nil {
			// 	serverResults = append(serverResults, httpsResult)
			// }

			// for _, result := range serverResults {
			// 	if record.Fingerprint != "" && strings.Contains(result.Body, record.Fingerprint) {
			// 		msg := "[CNAME Potential] " + result.Url + " | Cname: " + temp_cname + " | Service: " + record.Service
			// 		color.Red(msg)
			// 		if common.SendOutput {
			// 			common.OutputP.PublishMessage(msg)
			// 		}
			// 		notify.SendMessage(msg)
			// 		return true
			// 	}
			// }
		}
	}
	return false

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
