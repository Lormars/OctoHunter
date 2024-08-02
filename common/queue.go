package common

import (
	"container/heap"
	"net/http"
	"sync"
	"time"
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
	Client *http.Client
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
	queueLock sync.Mutex
	pq        *PriorityQueue[*Queue]       //for request queue
	bq        *PriorityQueue[*BrokerQueue] //for broker queue
)

func NewQueuePriorityQueue() *PriorityQueue[*Queue] {
	return &PriorityQueue[*Queue]{
		items: []*Queue{},
		less: func(i, j int) bool {
			return Sliding.GetRequestCount(pq.items[i].Request.Host) < Sliding.GetRequestCount(pq.items[j].Request.Host)
		},
	}
}

func NewQueueBrokerQueue() *PriorityQueue[*BrokerQueue] {
	return &PriorityQueue[*BrokerQueue]{
		items: []*BrokerQueue{},
		less: func(i, j int) bool {
			return BrokerSliding.GetRequestCount(bq.items[i].Host) < BrokerSliding.GetRequestCount(bq.items[j].Host)
		},
	}
}

func AddToQueue(host string, reqs []*http.Request, client *http.Client) chan []Response {
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
	for {

		req := getNextRequest()
		if req != nil {
			go func(req *Queue) {
				var responses []Response
				for _, r := range req.Request.Reqs {
					if r.Header.Get("Connection") == "" {
						r.Close = true
					}
					currentHostName := r.URL.Hostname()
					if _, exists := NeedBrowser[currentHostName]; exists {
						// logger.Warnf("Need browser for %s", currentHostName)
						resp, err := RequestWithBrowser(r, req.Request.Client)
						responses = append(responses, Response{Resp: resp, Err: err})
						continue
					}
					resp, err := req.Request.Client.Do(r)
					responses = append(responses, Response{Resp: resp, Err: err})
				}
				req.RespChan <- responses
				close(req.RespChan)
			}(req)
		}
		mu.Lock()
		currentLen := pq.Len()
		mu.Unlock()
		time.Sleep(time.Duration(1/(currentLen+1)) * 1 * time.Second)

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
