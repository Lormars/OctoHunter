package dispatcher

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/logger"
)

func Input(opts *common.Opts) {
	Init(opts)
	time.Sleep(5 * time.Second)
	file, err := os.Open(opts.DispatcherFile)
	if err != nil {
		logger.Errorln("Error opening file: ", err)
		return
	}
	defer file.Close()
	lineCh := make(chan string, opts.Concurrency)
	go func() {
		for line := range lineCh {
			common.DividerP.PublishMessage(line)
		}
	}()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCh <- line
	}
}
