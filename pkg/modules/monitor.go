package modules

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/auth"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/tools/controller"
)

var moduleManager *controller.ModuleManager

func Monitor(opts *common.Opts) {
	cacher.Init()

	//var username = os.Getenv("CONTROLLER_USERNAME")
	//var password = os.Getenv("CONTROLLER_PASSWORD")
	var username = "user"
	var password = "password"
	var port = os.Getenv("PORT")

	moduleManager = controller.NewModuleManager()

	http.HandleFunc("/start", auth.BasicAuth(startHandler, username, password))
	http.HandleFunc("/stop", auth.BasicAuth(stopHandler, username, password))

	fmt.Println("Starting server on :", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Error starting server: ", err)
	}

}

func startHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var opts common.Opts
	err := json.NewDecoder(r.Body).Decode(&opts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(opts)

	Startup(moduleManager, &opts)
	w.WriteHeader(http.StatusOK)

}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	moduleName := r.URL.Query().Get("module")
	if moduleName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(moduleName)

	moduleManager.StopModule(moduleName)

	w.WriteHeader(http.StatusOK)
}
