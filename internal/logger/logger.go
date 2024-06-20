package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var Logger *log.Logger
var logLevel LogLevel

func init() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file")
	}

	multiWrite := io.MultiWriter(os.Stdout, logFile)

	Logger = log.New(multiWrite, "INFO: ", log.Ldate|log.Ltime)
	logLevel = INFO
}

func SetLogLevel(level LogLevel) {
	logLevel = level
}

func logMessage(level LogLevel, prefix string, v ...interface{}) {
	if level >= logLevel {
		// Get the caller's file and line number
		_, file, line, ok := runtime.Caller(2)
		if ok {
			file = filepath.Base(file)
			Logger.SetPrefix(prefix)
			Logger.Output(2, fmt.Sprintf("%s:%d: %s", file, line, fmt.Sprint(v...)))
		} else {
			Logger.SetPrefix(prefix)
			Logger.Output(2, fmt.Sprint(v...))
		}
	}
}

func logFormatted(level LogLevel, prefix string, format string, v ...interface{}) {
	if level >= logLevel {
		// Get the caller's file and line number
		_, file, line, ok := runtime.Caller(2)
		if ok {
			file = filepath.Base(file)
			Logger.SetPrefix(prefix)
			Logger.Output(2, fmt.Sprintf("%s:%d: %s", file, line, fmt.Sprintf(format, v...)))
		} else {
			Logger.SetPrefix(prefix)
			Logger.Output(2, fmt.Sprintf(format, v...))
		}
	}
}

func Debugln(v ...interface{}) {
	logMessage(DEBUG, "DEBUG: ", v...)
}

func Debugf(format string, v ...interface{}) {
	logFormatted(DEBUG, "DEBUG: ", format, v...)
}

func Infoln(v ...interface{}) {
	logMessage(INFO, "INFO: ", v...)
}

func Infof(format string, v ...interface{}) {
	logFormatted(INFO, "INFO: ", format, v...)
}

func Warnln(v ...interface{}) {
	logMessage(WARN, "WARN: ", v...)
}

func Warnf(format string, v ...interface{}) {
	logFormatted(WARN, "WARN: ", format, v...)
}

func Errorln(v ...interface{}) {
	logMessage(ERROR, "ERROR: ", v...)
}

func Errorf(format string, v ...interface{}) {
	logFormatted(ERROR, "ERROR: ", format, v...)
}

func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		fmt.Printf("Unknown log level: %s, defaulting to INFO\n", level)
		return INFO
	}
}
