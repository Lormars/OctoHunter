package comparer

import (
	"github.com/lormars/requester/common"
)

func CompareResponse(resp1, resp2 *common.Response) (bool, string) {
	if resp1.Status != resp2.Status {
		return false, "status"
	}
	if resp1.Body != resp2.Body {
		return false, "body"
	}
	if CompareHeaders(resp1.Header, resp2.Header) == false {
		return false, "headers"
	}
	return true, ""
}
