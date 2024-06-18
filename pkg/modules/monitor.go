package modules

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/auth"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/tools"
	"github.com/lormars/octohunter/tools/controller"
	"github.com/rs/cors"
)

var moduleManager *controller.ModuleManager

func Monitor(opts *common.Opts) {
	var username = os.Getenv("CONTROLLER_USERNAME")
	var password = os.Getenv("CONTROLLER_PASSWORD")
	//var username = "user"
	//var password = "password"
	var port = os.Getenv("PORT")

	moduleManager = controller.NewModuleManager()

	r := mux.NewRouter()

	r.HandleFunc("/start", auth.BasicAuth(startHandler, username, password))
	r.HandleFunc("/stop", auth.BasicAuth(stopHandler, username, password))
	r.HandleFunc("/upload", auth.BasicAuth(tools.UploadHandler, username, password))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	fmt.Println("Starting server on :", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
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
	Startup(moduleManager, &opts)
	logger.Infof("Module %s started\n", opts.Module)
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
