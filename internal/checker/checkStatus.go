package checker

import "github.com/lormars/requester/common"

func CheckAccess(resp *common.Response) bool {
	if resp.Status >= 200 && resp.Status < 300 {
		return true
	}
	return false
}

func CheckRedirect(statusCode int) bool {
	if statusCode >= 300 && statusCode < 400 {
		return true
	}
	return false
}

func Check405(resp *common.Response) bool {
	return resp.Status == 405
}

func Check429(resp *common.Response) bool {
	return resp.Status == 429
}
