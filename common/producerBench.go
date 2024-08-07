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
	Average        float32
	Hosts          map[string]HostDetail
}

type HostDetail struct {
	Number  int
	Runtime []time.Duration
}

type ProducerAverage struct {
	Average float32
	count   int
}

var ProducerBenches = map[string]*ProducerBench{}
var ProducerAverages = map[string]*ProducerAverage{}

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
		ave, exists := ProducerAverages[pb.Producer]
		if exists {
			bo.Average = ave.Average
		} else {
			bo.Average = 0
		}
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

func DeleteFrom(output map[string]*ProducerBench, sig string) {
	producer := output[sig].Producer
	timeElapsed := time.Since(output[sig].Time)
	producerAverage, exists := ProducerAverages[producer]
	if !exists {
		producerAverage = &ProducerAverage{
			Average: 0,
			count:   1,
		}
		ProducerAverages[producer] = producerAverage
	}
	producerAverage.Average = updateAverage(producerAverage.Average, producerAverage.count, float32(timeElapsed.Seconds()))
	producerAverage.count++
	delete(output, sig)
}

func updateAverage(average float32, count int, newValue float32) float32 {
	return average + (newValue-average)/float32(count)
}
