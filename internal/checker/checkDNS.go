package checker

import (
	"net"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/miekg/dns"
)

func HasCname(hostname string) (bool, string, error) {
	cname, err := net.LookupCNAME(hostname)
	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			if dnsErr.Err == "no such host" {
				//no idea why, but sometimes a query would both return NXDomain and a cname...
				return true, "weird situation", nil
			}
			return false, "", nil
		}
		return false, "", err

	}
	cname = strings.TrimSuffix(cname, ".")

	if cname != hostname {
		return true, cname, nil

	}

	return false, "", nil

}

// Usage: finds the immediate cname record for a given hostname
// Why this? Because net.LookupCNAME does not return the immediate cname record and
// sometimes we need to know the immediate cname record to treat weird situation when NXDomain and cname both exist.
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
			target = strings.TrimSuffix(target, ".")
			return target
		}
	}
	return ""
}
