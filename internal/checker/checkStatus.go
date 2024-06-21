package checker

import "github.com/lormars/octohunter/common"

func CheckAccess(resp *common.ServerResult) bool {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
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

func CheckRequestError(statusCode int) bool {
	if statusCode >= 400 && statusCode < 500 {
		return true
	}
	return false
}

func Check405(resp *common.ServerResult) bool {
	return resp.StatusCode == 405
}

func Check429(resp *common.ServerResult) bool {
	return resp.StatusCode == 429
}
