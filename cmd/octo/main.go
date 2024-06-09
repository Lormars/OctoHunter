package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/parser"
	"github.com/lormars/octohunter/pkg/modules"
)

func main() {
	options := parser.Parse_Options()

	if options.Broker {
		common.Init()
		defer common.Close()
	}

	if options.Hopper {
		modules.CheckHop(options)
	}

	if options.Dork {
		modules.GoogleDork(options)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	log.Printf("Received signal: %s", sig)
	log.Println("Exiting...")
}
