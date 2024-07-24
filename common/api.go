package common

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync"

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
	DispatcherFile string     `json:"dispatcherFile"`
}

type Config struct {
	Loglevel    string
	CacheTime   int
	MemoryUsage bool
	RateLimit   int
	PurgeBroker bool
	UseProxy    bool
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
var Paths sync.Map
var Domains sync.Map
var SendOutput bool
var ConsumerSemaphore = make(chan struct{}, 565)

var NeedBrowser = make(map[string]bool)

type ServerResult struct {
	Url        string      `json:"url"`
	FinalUrl   *url.URL    `json:"final_url"`
	Online     bool        `json:"online"`
	StatusCode int         `json:"status_code"`
	Headers    http.Header `json:"headers"`
	Body       string      `json:"body"`
	Depth      int         `json:"depth"` // Used to limit the depth of the crawler
}

type XssInput struct {
	Url      string `json:"url"`
	Param    string `json:"param"`
	Location string `json:"location"`
}

func MapsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

// Merge map b into map a
func MergeMaps(a, b map[string]bool) {
	for k, v := range b {
		a[k] = v
	}
}

// Function to check if map A is a superset of map B
func IsSuperset(a, b map[string]bool) bool {
	for k, v := range b {
		if av, ok := a[k]; !ok || av != v {
			return false
		}
	}
	return true
}
