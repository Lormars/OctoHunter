package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/lormars/octohunter/internal/parser"
	"github.com/lormars/octohunter/pkg/modules"
	"github.com/lormars/octohunter/tools/controller"
)

func main() {

	options := parser.Parse_Options()

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	moduleManager := controller.NewModuleManager()

	modules.Startup(moduleManager, options)
}
