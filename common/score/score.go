package score

import (
	"math"

	"github.com/lormars/octohunter/common"
)

// weight
const (
	xss           = 10
	ssti          = 10
	redirect      = 10
	pathtraversal = 10
	method        = 10
	hop           = 10
	cors          = 10
	split         = 10
	cl0           = 10
	pathconfuse   = 20
	fuzz          = 10
	quirk         = 5
)

func CalculateScore() {
	originMap := common.GetOriginMap()

	score := make(map[string]int)
	var totalScore int
	var count int
	for domain, origins := range originMap {
		if len(origins) < 20 {
			continue
		}
		score[domain] = 0
		for _, origin := range origins {
			score[domain] += getScore(origin)
		}
		totalScore += score[domain]
		count++
	}

	if len(score) < 20 {
		return
	}

	average := float64(totalScore) / float64(count)

	var variance float64
	for _, s := range score {
		variance += math.Pow(float64(s)-average, 2)
	}
	variance /= float64(count)
	stdDev := math.Sqrt(variance)

	threshold := average + stdDev
	var results []string
	for domain, s := range score {
		if float64(s) > threshold {
			results = append(results, domain)
		}
	}
	// logger.Infof("all origins: %v", originMap)
	// logger.Infof("all domains: %v", score)
	// logger.Infof("High score domains: %v", result)

	for _, result := range results {
		go common.WaybackP.PublishMessage(result)
	}

}

func getScore(origin string) int {
	var score int
	switch origin {
	case "xss":
		score = xss
	case "ssti":
		score = ssti
	case "redirect":
		score = redirect
	case "pathtraversal":
		score = pathtraversal
	case "method":
		score = method
	case "hop":
		score = hop
	case "cors":
		score = cors
	case "split":
		score = split
	case "cl0":
		score = cl0
	case "pathconfuse":
		score = pathconfuse
	case "fuzz":
		score = fuzz
	case "quirk":
		score = quirk
	default:
		score = 0
	}
	return score
}
