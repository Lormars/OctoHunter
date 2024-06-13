package multiplex

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/lormars/octohunter/common"
)

func Conscan(f common.Atomic, options *common.Opts, concurrency int) {
	request_ch := make(chan *common.Opts)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for options := range request_ch {
				f(options)
			}
		}()
	}
	file, err := os.Open(options.File)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		request_ch <- &common.Opts{
			Hopper:   options.Hopper,
			Target:   line,
			File:     options.File,
			Method:   options.Method,
			Monitor:  options.Monitor,
			Redirect: options.Redirect,
			Cname:    options.Cname,
			Broker:   options.Broker,
			Dork:     options.Dork,
		}
	}
	close(request_ch)
	wg.Wait()
}
