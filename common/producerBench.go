package common

import (
	"time"
)

type ProducerBench struct {
	Producer string
	Hostname string
	Time     time.Time
}

type BenchOutput struct {
	ProducerNumber int
	Hosts          map[string]int
}

var ProducerBenches = []ProducerBench{}

func GetOutput() map[string]BenchOutput {
	now := time.Now()
	SecondsAgo := now.Add(-10 * time.Second)
	outputs := make(map[string]BenchOutput)
	var filtered []ProducerBench
	GlobalMu.Lock()
	for _, pb := range ProducerBenches {
		if pb.Time.After(SecondsAgo) {
			bo, exists := outputs[pb.Producer]
			if !exists {
				bo = BenchOutput{
					ProducerNumber: 0,
					Hosts:          make(map[string]int),
				}
			}
			bo.ProducerNumber++
			bo.Hosts[pb.Hostname]++
			outputs[pb.Producer] = bo
			filtered = append(filtered, pb)
		}
	}
	ProducerBenches = filtered
	GlobalMu.Unlock()
	return outputs
}
