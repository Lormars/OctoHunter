package smuggle

import (
	"testing"
	"time"
)

func TestCheckCl0(t *testing.T) {
	for {
		go CheckCl0("http://localhost:7777")
		time.Sleep(5 * time.Second)
	}

	select {}
}
