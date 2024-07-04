package checker

import (
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/logger"
)

func CheckCacheable(payload string) bool {

	var start time.Time
	var elapses []time.Duration
	for i := 0; i < 2; i++ {
		req, err := http.NewRequest("GET", payload, nil)
		if err != nil {
			logger.Debugf("Error creating request: %v", err)
			return false
		}
		trace := &httptrace.ClientTrace{
			WroteRequest: func(info httptrace.WroteRequestInfo) {
				start = time.Now()
			},
			GotFirstResponseByte: func() {
				elapses = append(elapses, time.Since(start))
			},
		}
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
		_, err = CheckServerCustom(req, clients.NoRedirectClient)
		if err != nil {
			logger.Debugf("Error getting response from %s: %v\n", payload, err)
			return false
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
