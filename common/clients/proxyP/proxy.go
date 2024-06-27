package proxyP

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/lormars/octohunter/common/clients/lightsailP"
	"github.com/lormars/octohunter/internal/logger"
)

type ProxyT struct {
	Proxies []string
	Mu      *sync.Mutex
}

var Proxies = ProxyT{
	Proxies: ParseProxies(),
	Mu:      &sync.Mutex{},
}

func ParseProxies() []string {
	fileName := "list/proxy"
	file, err := os.Open(fileName)
	if err != nil {
		logger.Errorf("Error opening file: %v\n", err)
		return nil
	}

	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		name, location := strings.Split(line, " ")[0], strings.Split(line, " ")[1]
		ip := lightsailP.GetIP(name, location)
		ipWithPort := ip + ":1080"
		//fmt.Println(ipWithPort)
		lines = append(lines, ipWithPort)
	}
	if err := scanner.Err(); err != nil {
		logger.Errorf("Error scanning file: %v\n", err)
		return nil
	}
	return lines
}

func DeleteProxy(proxy string) {
	Proxies.Mu.Lock()
	defer Proxies.Mu.Unlock()
	for i, p := range Proxies.Proxies {
		if p == proxy {
			Proxies.Proxies = append(Proxies.Proxies[:i], Proxies.Proxies[i+1:]...)
		}
	}
	logger.Warnf("Proxies after deletion: %v\n", Proxies.Proxies)
}

func AddProxy(proxy string) {
	Proxies.Mu.Lock()
	defer Proxies.Mu.Unlock()
	Proxies.Proxies = append(Proxies.Proxies, proxy)
	logger.Warnf("Proxies after addition: %v\n", Proxies.Proxies)
}
