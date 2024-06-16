package modules

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/auth"
	"github.com/lormars/octohunter/internal/controller"
	"github.com/lormars/octohunter/pkg/modules/takeover"
)

func Monitor(opts *common.Opts) {
	if opts.Cname {
		go func() {
			for {
				//takeover.MonitorPreprocess()
				takeover.CNAMETakeover(opts)
				time.Sleep(15 * time.Minute)
			}
		}()
	}

	if opts.Dork {
		go func() {
			for {
				GoogleDork(opts)
				time.Sleep(15 * time.Minute)
			}
		}()
	}

	if opts.Hopper {
		go func() {
			for {
				CheckHop(opts)
				time.Sleep(15 * time.Minute)
			}
		}()
	}

	if opts.Redirect {
		go func() {
			for {
				CheckRedirect(opts)
				time.Sleep(15 * time.Minute)
			}
		}()
	}

	if opts.Method {
		go func() {
			for {
				CheckMethod(opts)
				time.Sleep(15 * time.Minute)
			}
		}()
	}

	var username = os.Getenv("CONTROLLER_USERNAME")
	var password = os.Getenv("CONTROLLER_PASSWORD")
	var port = os.Getenv("PORT")

	http.HandleFunc("/command", auth.BasicAuth(controller.ControllerHandler, username, password))

	fmt.Println("Starting server on :", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Error starting server: ", err)
	}

}
