package comparer

import (
	"github.com/lormars/requester/common"
)

func CompareResponse(resp1, resp2 *common.Response) bool {
	if resp1.Status != resp2.Status {
		return false
	}
	if resp1.Body != resp2.Body {
		return false
	}
	if CompareHeaders(resp1.Header, resp2.Header) == false {
		return false
	}
	return true
}
