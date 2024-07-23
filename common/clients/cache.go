package clients

import (
	"net"
	"sync"
)

type DNSCache struct {
	mu    sync.RWMutex
	cache map[string][]net.IP
}

func NewDNSCache() *DNSCache {
	return &DNSCache{
		cache: make(map[string][]net.IP),
	}
}

func (c *DNSCache) LookupIP(host string) ([]net.IP, error) {
	c.mu.RLock()
	if ips, found := c.cache[host]; found {
		c.mu.RUnlock()
		return ips, nil
	}
	c.mu.RUnlock()

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[host] = ips
	c.mu.Unlock()

	return ips, nil
}
