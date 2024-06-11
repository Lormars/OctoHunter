package takeover

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/requester/pkg/runner"
)

var records []common.TakeoverRecord

func CNAMETakeover(options *common.Opts) {
	parseSignature("list/fingerprints.json")
	if options.Target == "none" {
		multiplex.Conscan(takeover, options, 10)
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

	checkSig(domain, cname)

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

func checkSig(domain, cname string) bool {
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
					fmt.Printf("Vulnerable to CNAME takeover through fingerprint: %s\n", url)
					return true
				}
				if record.Nxdomain && record.Fingerprint == "NXDOMAIN" && checkNXDomain(cname) {
					fmt.Printf("Vulnerable to CNAME takeover through nxdomain: %s\n", url)
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

func MonitorPreprocess() {
	commands := []struct {
		name string
		args []string
	}{
		{"bash", []string{"-c", "cat list/subdomains list/gunames | grep -v '\\*' | sort | uniq > list/subdomains"}},
		{"dnsx", []string{"-l", "list/subdomains", "-nc", "-cname", "-re", "-o", "list/cnames_raw"}},
		{"bash", []string{"-c", "cat list/cnames_raw | grep -iv 'shop.spacex.com' > list/cnames"}},
	}

	for _, command := range commands {
		err := common.RunCommand(command.name, command.args)
		if err != nil {
			fmt.Printf("Error running command %s: %s\n", command.name, err)
			return
		}
	}
	fmt.Println("Preprocess completed")
}
