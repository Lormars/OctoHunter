package clients

import (
	"container/heap"
	"net/http"
	"sync"
	"time"
)

type Queue struct {
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

type PriorityQueue []*Queue

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return Sliding.GetRequestCount(pq[i].Request.Host) < Sliding.GetRequestCount(pq[j].Request.Host)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Queue)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

var (
	queueLock sync.Mutex
	pq        PriorityQueue
)

func AddToQueue(host string, reqs []*http.Request, client *http.Client) chan []Response {
	queueLock.Lock()
	defer queueLock.Unlock()
	respChan := make(chan []Response)
	heap.Push(&pq, &Queue{Request: Request{Host: host, Client: client, Reqs: reqs}, RespChan: respChan})
	return respChan
}

func getNextRequest() *Queue {
	queueLock.Lock()
	defer queueLock.Unlock()

	if pq.Len() == 0 {
		return nil
	}

	return heap.Pop(&pq).(*Queue)
}

func init() {
	heap.Init(&pq)
	go dispatch()
}

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
					resp, err := req.Request.Client.Do(r)
					responses = append(responses, Response{Resp: resp, Err: err})
				}
				req.RespChan <- responses
			}(req)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
