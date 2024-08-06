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
	Hosts          map[string]HostDetail
}

type HostDetail struct {
	Number  int
	Runtime []time.Duration
}

var ProducerBenches = map[string]*ProducerBench{}

func GetOutput() map[string]BenchOutput {
	outputs := make(map[string]BenchOutput)
	GlobalMu.Lock()
	for _, pb := range ProducerBenches {
		bo, exists := outputs[pb.Producer]
		if !exists {
			bo = BenchOutput{
				ProducerNumber: 0,
				Hosts:          make(map[string]HostDetail),
			}
		}
		bo.ProducerNumber++
		hostDetail := bo.Hosts[pb.Hostname]
		hostDetail.Number++
		hostDetail.Runtime = append(hostDetail.Runtime, time.Since(pb.Time))
		bo.Hosts[pb.Hostname] = hostDetail
		outputs[pb.Producer] = bo
	}
	GlobalMu.Unlock()
	return outputs
}

func (hd HostDetail) Average() time.Duration {
	var total time.Duration
	for _, t := range hd.Runtime {
		total += t
	}
	return total / time.Duration(len(hd.Runtime))
}
