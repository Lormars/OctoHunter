package comparer

import "net/http"

func CompareHeaders(h1, h2 http.Header) bool {
	if len(h1) != len(h2) {
		return false
	}
	for k, v := range h1 {
		if v2, ok := h2[k]; !ok || len(v) != len(v2) {
			return false
		}
		for i, v1 := range v {
			if v1 != h2[k][i] {
				return false
			}
		}
	}
	return true
}
