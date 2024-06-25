package smuggle

import (
	"testing"
	"time"
)

func TestCheckCl0(t *testing.T) {
	for {
		go CheckCl0("https://example.com")
		time.Sleep(5 * time.Second)
	}

	select {}
}
