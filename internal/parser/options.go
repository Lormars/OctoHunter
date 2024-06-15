package parser

import (
	"flag"

	"github.com/lormars/octohunter/common"
)

func Parse_Options() *common.Opts {
	var (
		hopper       = flag.Bool("hopper", false, "Enable the hopper")
		dork         = flag.Bool("dork", false, "Enable the dorker")
		broker       = flag.Bool("broker", false, "Enable the broker")
		method       = flag.Bool("method", false, "Enable the HTTP method checker")
		cname        = flag.Bool("cname", false, "Enable the CNAME takeover checker")
		monitor      = flag.Bool("monitor", false, "Enable the monitor")
		redirect     = flag.Bool("redirect", false, "Enable the redirect checker")
		target       = flag.String("target", "none", "The target to scan")
		dnsFile      = flag.String("dnsfile", "none", "The file to scan for subdomain takeover")
		dorkFile     = flag.String("dorkfile", "none", "The file to scan for Google dork")
		methodFile   = flag.String("methodfile", "none", "The file to scan for HTTP method checker")
		redirectFile = flag.String("redirectfile", "none", "The file to scan for redirect checker")
		hopFile      = flag.String("hopfile", "none", "The file to scan for hopper")
	)
	flag.Parse()
	return &common.Opts{
		Hopper:       *hopper,
		Dork:         *dork,
		Broker:       *broker,
		Method:       *method,
		Monitor:      *monitor,
		Redirect:     *redirect,
		Cname:        *cname,
		Target:       *target,
		DnsFile:      *dnsFile,
		DorkFile:     *dorkFile,
		MethodFile:   *methodFile,
		RedirectFile: *redirectFile,
		HopFile:      *hopFile,
	}
}
