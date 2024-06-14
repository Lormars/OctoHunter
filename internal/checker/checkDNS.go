package checker

import (
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/miekg/dns"
)

// Usage: finds the immediate cname record for a given hostname
// Returns:
// - The immediate cname record if found
// - The original hostname is no cname record is found
// - errors indicating failed query or NXDomain
func FindImmediateCNAME(hostname string) (string, error) {
	c := new(dns.Client)
	m := new(dns.Msg)

	question := hostname + "."
	m.SetQuestion(question, dns.TypeCNAME)
	m.RecursionDesired = true

	dnsServer := "1.1.1.1:53"
	r, _, err := c.Exchange(m, dnsServer)

	if err != nil {
		//fmt.Printf("DNS query failed: %v\n", err)
		return "", err
	}

	if r.Rcode == dns.RcodeNameError {
		if len(r.Answer) > 0 {
			target := answertoCname(r)
			return target, common.ErrNXDOMAIN

		}
		return hostname, common.ErrNXDOMAIN
	}

	if len(r.Answer) > 0 {
		target := answertoCname(r)
		return target, nil

	}

	return hostname, nil

}

func answertoCname(r *dns.Msg) string {
	for _, ans := range r.Answer {
		if cname, ok := ans.(*dns.CNAME); ok {
			target := cname.Target
			if strings.HasSuffix(target, ".") {
				target = target[:len(target)-1]
			}
			return target
		}
	}
	return ""
}
