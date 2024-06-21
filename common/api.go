package common

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	"github.com/lormars/octohunter/internal/logger"
)

type ModuleList []string

func (m *ModuleList) String() string {
	return strings.Join(*m, ", ")
}
func (m *ModuleList) Set(value string) error {
	*m = strings.Split(value, ",")
	return nil
}

func (m *ModuleList) Contains(module string) bool {
	for _, m := range *m {
		if m == module {
			return true
		}
	}
	return false
}

func (m *ModuleList) UnmarshalJSON(data []byte) error {
	var modules string
	if err := json.Unmarshal(data, &modules); err != nil {
		return err
	}
	*m = strings.Split(modules, ",")
	return nil
}

type Opts struct {
	Module         ModuleList `json:"modules"`
	Concurrency    int        `json:"concurrency"`
	Target         string     `json:"target"`
	DorkFile       string     `json:"dorkFile"`
	HopperFile     string     `json:"hopperFile"`
	MethodFile     string     `json:"methodFile"`
	RedirectFile   string     `json:"redirectFile"`
	CnameFile      string     `json:"cnameFile"`
	DispatcherFile string     `json:"dispatcherFile"`
}

type TakeoverRecord struct {
	CicdPass      bool     `json:"cicd_pass"`
	Cname         []string `json:"cname"`
	Discussion    string   `json:"discussion"`
	Documentation string   `json:"documentation"`
	Fingerprint   string   `json:"fingerprint"`
	HttpStatus    *int     `json:"http_status"` // Use *int to handle null values
	Nxdomain      bool     `json:"nxdomain"`
	Service       string   `json:"service"`
	Status        string   `json:"status"`
	Vulnerable    bool     `json:"vulnerable"`
}

const (
	OK = iota
	REDIRECT
	CLIENTERR
	SERVERERR
	XERROR
)

type Atomic func(options *Opts)

func RunCommand(name string, args []string) error {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		logger.Errorf("Error running command: %s\n", out.String())
		return err
	}
	return nil
}

var DB *sql.DB

type ServerResult struct {
	Url        string      `json:"url"`
	Online     bool        `json:"online"`
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers"`
	Body       string      `json:"body"`
}
