package checker

import (
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/logger"
)

func CheckCacheable(payload string) bool {
	var elapses []time.Duration
	for i := 0; i < 2; i++ {
		req, err := http.NewRequest("GET", payload, nil)
		if err != nil {
			logger.Debugf("Error creating request: %v", err)
			return false
		}
		elapse, _, err := MeasureElapse(req, clients.NoRedirectClient)
		if err == nil {
			elapses = append(elapses, elapse)
		}

	}
	if len(elapses) != 2 {
		return false
	}
	if elapses[0] > 0 && elapses[1] > 0 && (elapses[0]/2) > elapses[1] {
		return true
	}
	return false
}

func MeasureElapse(req *http.Request, client *http.Client) (time.Duration, *common.ServerResult, error) {
	var start time.Time
	var elapse time.Duration
	trace := &httptrace.ClientTrace{
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			start = time.Now()
		},
		GotFirstResponseByte: func() {
			elapse = time.Since(start)
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	resp, err := CheckServerCustom(req, client)
	if err != nil {
		return 0, nil, err
	}
	return elapse, resp, nil

}
