package comparer

import "github.com/lormars/octohunter/common"

func CompareResponse(resp1, resp2 *common.ServerResult) (bool, string) {
	if resp1.StatusCode != resp2.StatusCode {
		return false, "status"
	}
	if resp1.Body != resp2.Body {
		return false, "body"
	}
	if CompareHeaders(resp1.Headers, resp2.Headers) == false {
		return false, "headers"
	}
	return true, ""
}
