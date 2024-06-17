package multiplex

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/cacher"
)

func Conscan(ctx context.Context, f common.Atomic, options *common.Opts, fileName, cacheName string, concurrency int) {
	request_ch := make(chan *common.Opts)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for options := range request_ch {
				select {
				case <-ctx.Done():
					fmt.Print("Context Done!!!\n")
					return
				default:
					f(options)
					cacher.UpdateScanTime(options.Target, cacheName)
				}
			}
		}()
	}
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
Loop:
	for scanner.Scan() {
		line := scanner.Text()
		if !cacher.CanScan(line, cacheName) {
			fmt.Println("Skipping: ", line)
			continue
		}
		select {
		case <-ctx.Done():
			fmt.Println("Loop breaked!!!")
			break Loop
		case request_ch <- &common.Opts{
			Module:       options.Module,
			Target:       line,
			DorkFile:     options.DorkFile,
			HopFile:      options.HopFile,
			MethodFile:   options.MethodFile,
			RedirectFile: options.RedirectFile,
			DnsFile:      options.DnsFile,
		}:
		}
	}
	close(request_ch)
	wg.Wait()
}
