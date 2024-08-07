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
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	dispatcher "github.com/lormars/octohunter/dispatcher"
	"github.com/lormars/octohunter/internal/bench"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/parser"
)

func main() {

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	// Directory to check and delete
	dirPath := "output"

	// Check if the directory exists
	if _, err := os.Stat(dirPath); err == nil {
		// Directory exists, delete it and all its contents
		err = os.RemoveAll(dirPath)
		if err != nil {
			log.Fatalf("Error deleting directory: %v", err)
			return
		}
	}

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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan

	cmap.Save("output")
	logger.Infof("Received signal: %s. Shutting down gracefully...\n", sig)
	for _, producer := range producers {
		close(producer.ShutdownChan)
	}

	logger.Infof("All producers shut down. Exiting...\n")

	common.Close()
	logger.Infof("All connections closed. Exiting...\n")

	if err := cmd.Process.Kill(); err != nil {
		log.Fatalf("Error killing node server: %v", err)
	}
	logger.Infoln("Exiting...")
}
