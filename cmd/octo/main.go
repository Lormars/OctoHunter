package main

import (
	"log"
	_ "net/http/pprof"

	"github.com/joho/godotenv"
	dispatcher "github.com/lormars/octohunter/Dispatcher"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/parser"
	"github.com/lormars/octohunter/pkg/modules"
	"github.com/lormars/octohunter/tools/controller"
)

func main() {

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	options, logLevel := parser.Parse_Options()
	logger.SetLogLevel(logger.ParseLogLevel(logLevel))
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	moduleManager := controller.NewModuleManager()

	if options.Module.Contains("broker") {
		common.Init()
		defer common.Close()
	}

	if options.Module.Contains("dispatcher") {
		dispatcher.Input(options)
	}

	if options.Module.Contains("monitor") {
		modules.Monitor(options)
	} else {
		modules.Startup(moduleManager, options)
	}

	select {}
}
