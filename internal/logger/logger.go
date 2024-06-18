package logger

import (
	"io"
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file")
	}

	multiWrite := io.MultiWriter(os.Stdout, logFile)

	Logger = log.New(multiWrite, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Infoln(v ...interface{}) {
	Logger.Println(v...)
}

func Infof(format string, v ...interface{}) {
	Logger.Printf(format, v...)
}
