package multiplex

import (
	"bufio"
	"context"
	"os"
	"runtime"
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/logger"
)

func Conscan(ctx context.Context, f common.Atomic, options *common.Opts, fileName, cacheName string, concurrency int) {
	request_ch := make(chan *common.Opts, concurrency)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for options := range request_ch {
				select {
				case <-ctx.Done():
					logger.Infof("Conscan Done for %s\n", cacheName)
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
		logger.Errorln("Error opening file: ", err)
		return
	}
	defer file.Close()
	lineCount := 0
	gcInterval := 10000
	scanner := bufio.NewScanner(file)
Loop:
	for scanner.Scan() {
		line := scanner.Text()
		if !cacher.CanScan(line, cacheName) {

			continue
		}
		select {
		case <-ctx.Done():
			logger.Infof("Loop Done for %s\n", cacheName)
			break Loop
		case request_ch <- &common.Opts{
			Module:         options.Module,
			Concurrency:    options.Concurrency,
			Target:         line,
			DorkFile:       options.DorkFile,
			HopperFile:     options.HopperFile,
			MethodFile:     options.MethodFile,
			RedirectFile:   options.RedirectFile,
			CnameFile:      options.CnameFile,
			DispatcherFile: options.DispatcherFile,
		}:
			lineCount++
			//logger.Infof("Sending %s to request_ch\n", line)
			if lineCount%gcInterval == 0 {
				logger.Infoln("GC Interval Reached, Running GC")
				runtime.GC()
			}
		}
	}
	close(request_ch)
	wg.Wait()
	runtime.GC()
}
