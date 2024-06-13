package common

import (
	"bytes"
	"fmt"
	"os/exec"
)

type Opts struct {
	Hopper   bool
	Dork     bool
	Broker   bool
	Method   bool
	Cname    bool
	Monitor  bool
	Redirect bool
	Target   string
	File     string
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
