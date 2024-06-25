package smuggle

import (
	"net/http"
	"sync"
)

var Client *http.Client
var mu sync.Mutex

var HostMutexes = make(map[string]*sync.Mutex)

func GetHostMutex(host string) *sync.Mutex {
	mu.Lock()
	defer mu.Unlock()
	_, exists := HostMutexes[host]
	if !exists {
		HostMutexes[host] = &sync.Mutex{}
	}
	return HostMutexes[host]
}
