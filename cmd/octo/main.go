package main

import (
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	dispatcher "github.com/lormars/octohunter/Dispatcher"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/parser"
	"github.com/lormars/octohunter/pkg/modules"
	"github.com/lormars/octohunter/tools/controller"
)

func printMemUsage() {
	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// For more info, see: https://golang.org/pkg/runtime/#MemStats
		logger.Debugf("Alloc = %v MiB", bToMb(m.Alloc))
		logger.Debugf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
		logger.Debugf("\tSys = %v MiB", bToMb(m.Sys))
		logger.Debugf("\tNumGC = %v\n", m.NumGC)
		time.Sleep(1 * time.Second)
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func main() {

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	options, logLevel, cacheTime, mu := parser.Parse_Options()
	if mu {
		go printMemUsage()
	}
	logger.SetLogLevel(logger.ParseLogLevel(logLevel))
	cacher.SetCacheTime(cacheTime)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	moduleManager := controller.NewModuleManager()

	if options.Module.Contains("broker") {
		common.Init()
	}

	if options.Module.Contains("dispatcher") {
		go dispatcher.Input(options)
	}

	if options.Module.Contains("monitor") {
		modules.Monitor(options)
	} else {
		modules.Startup(moduleManager, options)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	logger.Infof("Received signal: %s. Shutting down gracefully...\n", sig)

	common.Close()
	logger.Infoln("Exiting...")
}
