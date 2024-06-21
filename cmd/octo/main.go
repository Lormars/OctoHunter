package main

import (
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	dispatcher "github.com/lormars/octohunter/Dispatcher"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/bench"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/parser"
	"github.com/lormars/octohunter/pkg/modules"
	"github.com/lormars/octohunter/tools/controller"
)

func main() {

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	options, logLevel, cacheTime, mu := parser.Parse_Options()
	if mu {
		go bench.PrintMemUsage(options)
	}
	logger.SetLogLevel(logger.ParseLogLevel(logLevel))
	cacher.SetCacheTime(cacheTime)
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	if options.Module.Contains("broker") {
		common.Init()
	}

	if options.Module.Contains("dispatcher") {
		go dispatcher.Input(options)
	}

	if options.Module.Contains("monitor") {
		modules.Monitor(options)
	} else {
		moduleManager := controller.NewModuleManager()
		modules.Startup(moduleManager, options)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	logger.Infof("Received signal: %s. Shutting down gracefully...\n", sig)

	common.Close()
	logger.Infoln("Exiting...")
}
