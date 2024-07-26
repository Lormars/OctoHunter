package clients

import (
	"context"
	"net"
	"sync"
	"time"
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

var resolver = &net.Resolver{
	PreferGo: true,
	Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{
			Timeout: 30 * time.Second,
		}
		return d.DialContext(ctx, network, "1.1.1.1:53")
	},
}

func (c *DNSCache) LookupIP(host string) ([]net.IP, error) {
	c.mu.RLock()
	if ips, found := c.cache[host]; found {
		c.mu.RUnlock()
		// logger.Warnf("returnning")
		return ips, nil
	}
	c.mu.RUnlock()

	allIPs, err := resolver.LookupIP(context.Background(), "ip", host)
	if err != nil {
		return nil, err
	}

	// Filter out IPv6 addresses
	var ipv4s []net.IP
	for _, ip := range allIPs {
		if ip.To4() != nil {
			ipv4s = append(ipv4s, ip)
			// logger.Warnf("Found IPv4 address: %s\n", ip)
		}
	}

	c.mu.Lock()
	c.cache[host] = ipv4s
	c.mu.Unlock()
	// logger.Warnf("returnning")
	return ipv4s, nil
}
