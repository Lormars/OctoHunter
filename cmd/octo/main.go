package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/parser"
	"github.com/lormars/octohunter/pkg/modules"
	"github.com/lormars/octohunter/pkg/modules/takeover"
)

func main() {
	options := parser.Parse_Options()

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	if options.Broker {
		common.Init()
		defer common.Close()
	}

	if options.Monitor {
		modules.Monitor(options)

	}

	if options.Hopper {
		modules.CheckHop(options)
	}

	if options.Dork {
		modules.GoogleDork(options)
	}

	if options.Method {
		modules.CheckMethod(options)
	}

	if options.Cname {
		takeover.CNAMETakeover(options)
	}

}
