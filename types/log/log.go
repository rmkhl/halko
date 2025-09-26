package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type LogLevel int

const (
	ERROR LogLevel = iota
	WARN
	INFO
	DEBUG
	TRACE
)

var (
	mu       sync.RWMutex
	level    LogLevel  = INFO
	output   io.Writer = os.Stdout
	traceLog *log.Logger
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
	nullLog  *log.Logger
)

func init() {
	nullLog = log.New(io.Discard, "", 0)
	SetLevel(level)
}

func SetLevel(l LogLevel) {
	mu.Lock()
	defer mu.Unlock()

	level = l

	traceLog = nullLog
	debugLog = nullLog
	infoLog = nullLog
	warnLog = nullLog
	errorLog = log.New(output, "[ERROR] ", log.LstdFlags)

	if l >= WARN {
		warnLog = log.New(output, "[WARN]  ", log.LstdFlags)
	}
	if l >= INFO {
		infoLog = log.New(output, "[INFO]  ", log.LstdFlags)
	}
	if l >= DEBUG {
		debugLog = log.New(output, "[DEBUG] ", log.LstdFlags)
	}
	if l >= TRACE {
		traceLog = log.New(output, "[TRACE] ", log.LstdFlags)
	}
}

func SetOutput(w io.Writer) {
	mu.Lock()
	output = w
	currentLevel := level
	mu.Unlock()
	SetLevel(currentLevel)
}

func Debug(format string, v ...interface{}) {
	debugLog.Println(fmt.Sprintf(format, v...))
}

func Info(format string, v ...interface{}) {
	infoLog.Println(fmt.Sprintf(format, v...))
}

func Warning(format string, v ...interface{}) {
	warnLog.Println(fmt.Sprintf(format, v...))
}

func Error(format string, v ...interface{}) {
	errorLog.Println(fmt.Sprintf(format, v...))
}

func Trace(format string, v ...interface{}) {
	traceLog.Println(fmt.Sprintf(format, v...))
}

func Fatal(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}
