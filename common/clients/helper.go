package clients

import (
	"bufio"
	"os"

	"github.com/lormars/octohunter/internal/logger"
)

func ParseProxies() []string {
	fileName := "list/proxy"
	file, err := os.Open(fileName)
	if err != nil {
		logger.Errorf("Error opening file: %v\n", err)
		return nil
	}

	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		logger.Errorf("Error scanning file: %v\n", err)
		return nil
	}
	return lines
}
