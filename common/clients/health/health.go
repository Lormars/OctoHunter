package health

import (
	"math"
	"sync"
	"time"

	"github.com/lormars/octohunter/common/clients/lightsailP"
	"github.com/lormars/octohunter/common/clients/proxyP"
	"github.com/lormars/octohunter/internal/logger"
)

type ElementType int

const (
	good ElementType = iota
	bad
)

const bufferSize = 500

type FixedQueue struct {
	data      []ElementType
	start     int
	end       int
	size      int
	cap       int
	goodCount int
	badCount  int
	lock      sync.Mutex
}

// NewFixedQueue creates a new FixedQueue with the given capacity
func NewFixedQueue(capacity int) *FixedQueue {
	return &FixedQueue{
		data: make([]ElementType, capacity),
		cap:  capacity,
	}
}

// Enqueue adds an element to the queue. If the queue is full, it overwrites the oldest element.
func (q *FixedQueue) Enqueue(elementType ElementType) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.size == q.cap {
		oldElement := q.data[q.start]
		if oldElement == good {
			q.goodCount--
		} else {
			q.badCount--
		}
		// Queue is full, overwrite the oldest element
		q.start = (q.start + 1) % q.cap
		q.size--
	}

	q.data[q.end] = elementType
	q.end = (q.end + 1) % q.cap
	q.size++

	if elementType == good {
		q.goodCount++
	} else {
		q.badCount++
	}
}

// Dequeue removes and returns the oldest element from the queue. Returns nil if the queue is empty.
func (q *FixedQueue) Dequeue() *ElementType {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.size == 0 {
		return nil
	}

	element := q.data[q.start]
	q.data[q.start] = 0 // Optional: Clear the slot
	q.start = (q.start + 1) % q.cap
	q.size--

	if element == good {
		q.goodCount--
	} else if element == bad {
		q.badCount--
	}

	return &element
}

// Size returns the number of elements currently in the queue.
func (q *FixedQueue) Size() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.size
}

// Capacity returns the capacity of the queue.
func (q *FixedQueue) Capacity() int {
	return q.cap
}

// GoodCount returns the number of good elements in the queue.
func (q *FixedQueue) GoodCount() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.goodCount
}

// BadCount returns the number of bad elements in the queue.
func (q *FixedQueue) BadCount() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.badCount
}

type ProxyHealth struct {
	Proxies map[string]*FixedQueue
	mu      sync.Mutex
}

var ProxyHealthInstance *ProxyHealth

func init() {
	ProxyHealthInstance = NewProxyHealth()
	for _, proxy := range proxyP.Proxies.Proxies {
		ProxyHealthInstance.AddProxy(proxy)
	}

	go func() {
		for {
			ProxyHealthInstance.Monitor()
			time.Sleep(1 * time.Second)
		}
	}()
}

func NewProxyHealth() *ProxyHealth {
	return &ProxyHealth{
		Proxies: make(map[string]*FixedQueue),
	}
}

func (ph *ProxyHealth) AddProxy(proxy string) {
	ph.mu.Lock()
	defer ph.mu.Unlock()

	ph.Proxies[proxy] = NewFixedQueue(bufferSize)
}

func (ph *ProxyHealth) AddGood(proxy string) {
	ph.mu.Lock()
	defer ph.mu.Unlock()

	ph.Proxies[proxy].Enqueue(good)
}

func (ph *ProxyHealth) AddBad(proxy string) {
	ph.mu.Lock()
	defer ph.mu.Unlock()

	ph.Proxies[proxy].Enqueue(bad)
}

func (ph *ProxyHealth) GetHealth(proxy string) float64 {
	ph.mu.Lock()
	defer ph.mu.Unlock()

	total := ph.Proxies[proxy].GoodCount() + ph.Proxies[proxy].BadCount()
	if total == 0 {
		return float64(1)
	}
	return float64(ph.Proxies[proxy].GoodCount()) / float64(total)
}

//var test = true

// This monitors the health of proxies based on the most recent 500 requests
func (ph *ProxyHealth) Monitor() {
	ph.mu.Lock()
	defer ph.mu.Unlock()

	// Calculate average health
	var sum float64
	count := len(ph.Proxies)
	for _, health := range ph.Proxies {
		total := health.GoodCount() + health.BadCount()
		if total == 0 || total < bufferSize {
			return
		} else {
			sum += float64(health.GoodCount()) / float64(total)
		}
	}
	average := sum / float64(count)

	// Calculate standard deviation
	var sumOfSquares float64
	for _, health := range ph.Proxies {
		total := health.GoodCount() + health.BadCount()
		var healthRatio float64
		if total == 0 {
			return
		} else {
			healthRatio = float64(health.GoodCount()) / float64(total)
		}
		sumOfSquares += (healthRatio - average) * (healthRatio - average)
	}
	stdDev := math.Sqrt(sumOfSquares / float64(count))
	// logger.Debugln(stdDev)
	// Print proxies with health lower than (average - 2*stdDev)
	for proxy, health := range ph.Proxies {
		total := health.GoodCount() + health.BadCount()
		var healthRatio float64
		if total == 0 {
			return
		} else {
			healthRatio = float64(health.GoodCount()) / float64(total)
		}
		//fmt.Println("Health ratio: ", healthRatio, "for proxy: ", proxy)
		if healthRatio < (average - 2*stdDev) {
			// if test {
			logger.Warnf("Proxy %s has abnormal health: %.2f\n", proxy, healthRatio)
			logger.Warnf("Getting new IP for proxy %s\n", proxy)
			// Get new IP for the proxy
			proxyP.DeleteProxy(proxy)
			newIP := lightsailP.ReGetIp(proxy)
			newIPWithPort := newIP + ":1080"
			proxyP.AddProxy(newIPWithPort)
			ph.Proxies[newIPWithPort] = NewFixedQueue(bufferSize)
			delete(ph.Proxies, proxy)

		}
	}
}
