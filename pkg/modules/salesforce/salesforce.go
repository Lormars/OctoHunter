package salesforce

import (
	"strings"

	"github.com/lormars/octohunter/internal/logger"
)

func SalesforceScan(target string) {

	logger.Debugf("SalesforceScan running on %s\n", target)
	var endpointToTest string //url
	var ok bool

	endpointToTest = target

	//target has two possibilities: one is url, the other is domain
	//if domain, it is from cname, then need to check fingerprint
	//if url, then it is already checked by other services, so skip.
	if !strings.HasPrefix(target, "http") {
		//Comes from cname, need to double check it is running salesforce
		if ok, endpointToTest = Fingerprint("https://" + target); ok {
			logger.Debugf("Salesforce found on %s\n", endpointToTest)
		} else {
			return
		}
	}
	logger.Debugf("Salesforce Pull Custom Objects running on %s\n", endpointToTest)

	err := PullCustomObjects(endpointToTest)
	if err != nil {
		logger.Debugf("Error pulling custom objects from %s: %v\n", endpointToTest, err)
		return
	}
}
