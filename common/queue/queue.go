package queue

import (
	"container/heap"
	"fmt"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
)

type Queue struct {
	Host     string
	Request  Request
	RespChan chan []Response
}

type Response struct {
	Resp *http.Response
	Err  error
}

type Request struct {
	Host   string
	Client *clients.OctoClient
	Reqs   []*http.Request
}

type BrokerQueue struct {
	Host string
	ch   chan bool
}

type PriorityQueueItem interface {
	*Queue | *BrokerQueue
}

type PriorityQueue[T PriorityQueueItem] struct {
	items []T
	less  func(i, j int) bool
}

func (pq PriorityQueue[T]) Len() int {
	uniqueHosts := make(map[string]struct{})
	for _, item := range pq.items {
		var host string
		switch v := any(item).(type) {
		case *Queue:
			host = v.Host
		case *BrokerQueue:
			host = v.Host
		default:
			continue
		}
		if _, exists := uniqueHosts[host]; !exists {
			uniqueHosts[host] = struct{}{}
		}
	}
	return len(uniqueHosts)
}

func (pq PriorityQueue[T]) Less(i, j int) bool {
	return pq.less(i, j)
}

func (pq PriorityQueue[T]) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

func (pq *PriorityQueue[T]) Push(x interface{}) {
	item := x.(T)
	mu.Lock()
	pq.items = append(pq.items, item)
	mu.Unlock()
}

func (pq *PriorityQueue[T]) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	pq.items = old[0 : n-1]
	return item
}

var (
	queueLock   sync.Mutex
	pq          *PriorityQueue[*Queue]       //for request queue
	bq          *PriorityQueue[*BrokerQueue] //for broker queue
	workerCount int32
	mu          sync.Mutex
)

func NewQueuePriorityQueue() *PriorityQueue[*Queue] {
	return &PriorityQueue[*Queue]{
		items: []*Queue{},
		less: func(i, j int) bool {
			return common.Sliding.GetRequestCount(pq.items[i].Request.Host) < common.Sliding.GetRequestCount(pq.items[j].Request.Host)
		},
	}
}

func NewQueueBrokerQueue() *PriorityQueue[*BrokerQueue] {
	return &PriorityQueue[*BrokerQueue]{
		items: []*BrokerQueue{},
		less: func(i, j int) bool {
			return common.BrokerSliding.GetRequestCount(bq.items[i].Host) < common.BrokerSliding.GetRequestCount(bq.items[j].Host)
		},
	}
}

func AddToQueue(host string, reqs []*http.Request, client *clients.OctoClient) chan []Response {
	queueLock.Lock()
	defer queueLock.Unlock()
	respChan := make(chan []Response)
	heap.Push(pq, &Queue{Request: Request{Host: host, Client: client, Reqs: reqs}, Host: host, RespChan: respChan})
	return respChan
}

func AddToBrokerQueue(host string) chan bool {
	queueLock.Lock()
	defer queueLock.Unlock()
	respChan := make(chan bool)
	heap.Push(bq, &BrokerQueue{Host: host, ch: respChan})
	return respChan
}

func getNextRequest() *Queue {
	queueLock.Lock()
	defer queueLock.Unlock()

	if pq.Len() == 0 {
		return nil
	}

	return heap.Pop(pq).(*Queue)
}

func getNextBrokerMessage() *BrokerQueue {
	queueLock.Lock()
	defer queueLock.Unlock()

	if bq.Len() == 0 {
		return nil
	}

	return heap.Pop(bq).(*BrokerQueue)
}

func init() {
	pq = NewQueuePriorityQueue()
	bq = NewQueueBrokerQueue()
	heap.Init(bq)
	heap.Init(pq)
	go dispatch()
	go dispatchBroker()
}

// This function is used to dispatch the request queue.
// The sleep time must be greater than the dispatchBroker function.
func dispatch() {

	wg := sync.WaitGroup{}

	reqChannel := make(chan *Queue, 200)
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for req := range reqChannel {
				var responses []Response
				if req != nil && req.Request.Client != nil {
					atomic.AddInt32(&workerCount, 1)
					for _, r := range req.Request.Reqs {

						// currentHostName := r.URL.Hostname()

						if r.Header.Get("Connection") == "" {
							r.Close = true
						}
						// if _, exists := common.NeedBrowser[currentHostName]; exists {
						// 	// logger.Warnf("Need browser for %s", currentHostName)
						// 	resp, err := common.RequestWithBrowser(r, req.Request.Client)
						// 	responses = append(responses, Response{Resp: resp, Err: err})
						// 	continue
						// }
						resp, err := req.Request.Client.RetryableDo(r)
						responses = append(responses, Response{Resp: resp, Err: err})
					}
				} else {
					responses = []Response{Response{Resp: nil, Err: fmt.Errorf("Client or Req is nil")}}
				}
				req.RespChan <- responses
				close(req.RespChan)
				atomic.AddInt32(&workerCount, -1)
			}
			wg.Done()
		}()
	}

	for {
		req := getNextRequest()
		if req == nil {
			time.Sleep(time.Second) // Sleep briefly if no requests are available
			continue
		}

		select {
		case reqChannel <- req:
			// Request successfully sent to the channel
		default:
			// Channel is full, wait for a short period before trying again
			time.Sleep(time.Millisecond * 100)
		}

		mu.Lock()
		currentLen := pq.Len()
		mu.Unlock()

		// Use float division to avoid integer truncation
		sleepDuration := time.Duration(math.Max(1.0/(float64(currentLen)+1), 0.1)) * time.Second
		time.Sleep(sleepDuration)
	}
}

// This function is used to dispatch the broker queue.
// The sleep time must be less than the dispatch function.
func dispatchBroker() {
	for {
		//fmt.Println("length: ", bq.Len())
		broker := getNextBrokerMessage()
		if broker != nil {
			broker.ch <- true
			close(broker.ch)
		}
		time.Sleep(time.Duration(1/(bq.Len()+1)) * 500 * time.Millisecond)
	}
}
func GetConcurrentRequests() int32 {
	activeWorkers := atomic.LoadInt32(&workerCount)
	return activeWorkers
}
