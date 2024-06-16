package common

import (
	"bytes"
	"database/sql"
	"fmt"
	"os/exec"
)

type Opts struct {
	Hopper       bool   `json:"hopper"`
	Dork         bool   `json:"dork"`
	Broker       bool   `json:"broker"`
	Method       bool   `json:"method"`
	Cname        bool   `json:"cname"`
	Monitor      bool   `json:"monitor"`
	Redirect     bool   `json:"redirect"`
	Target       string `json:"target"`
	DorkFile     string `json:"dorkFile"`
	HopFile      string `json:"hopFile"`
	MethodFile   string `json:"methodFile"`
	RedirectFile string `json:"redirectFile"`
	DnsFile      string `json:"dnsFile"`
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
		fmt.Printf("Error running command: %s\n", out.String())
		return err
	}
	return nil
}

var DB *sql.DB
