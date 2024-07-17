package main

import (
	"io"
	"log"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	cmap "github.com/lormars/crawlmap/pkg"
	dispatcher "github.com/lormars/octohunter/Dispatcher"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
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

	//disable default logger to get rid of unwanted warning
	log.SetOutput(io.Discard)

	options, config := parser.Parse_Options()
	if config.MemoryUsage {
		go bench.PrintMemUsage(options)
		common.SendOutput = true
	}
	logger.SetLogLevel(logger.ParseLogLevel(config.Loglevel))
	cacher.SetCacheTime(config.CacheTime)
	clients.SetRateLimiter(config.RateLimit)
	clients.SetUseProxy(config.UseProxy)
	err := godotenv.Load()

	cmd := exec.Command("node", "externals/nodejs/server.js")
	if err := cmd.Start(); err != nil {
		log.Fatalf("Error starting node server: %v", err)
	}

	var producers []*common.Producer

	if err != nil {
		log.Println("No .env file found")
	}

	if options.Module.Contains("broker") {
		producers = common.Init(options, config.PurgeBroker)
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
	for _, producer := range producers {
		close(producer.ShutdownChan)
	}

	logger.Infof("All producers shut down. Exiting...\n")

	common.Close()
	logger.Infof("All connections closed. Exiting...\n")

	cmap.Save("output")
	if err := cmd.Process.Kill(); err != nil {
		log.Fatalf("Error killing node server: %v", err)
	}
	logger.Infoln("Exiting...")
}
