package utils

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"time"
)

type LogLevel string

const (
	Info  LogLevel = "INFO"
	Warn  LogLevel = "WARN"
	Error LogLevel = "ERROR"
)

func Log(level LogLevel, format string, vars ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "???"
		line = 0
	}
	fileName := path.Base(file)
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	message := fmt.Sprintf(format, vars...)
	logLine := fmt.Sprintf("[%s] [%s] [%s:%d] %s", timestamp, level, fileName, line, message)
	if level == Error {
		log.Fatalln(logLine)
	}
	fmt.Fprintln(os.Stdout, logLine)
}
