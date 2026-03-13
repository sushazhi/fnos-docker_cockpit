package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

var (
	logLevel = INFO
	logFile  *os.File
)

func Init(dataDir string, level Level) error {
	logLevel = level
	
	if dataDir != "" {
		logPath := filepath.Join(dataDir, "app.log")
		var err error
		logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		log.SetOutput(logFile)
	}
	
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	return nil
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func logMessage(level Level, format string, v ...interface{}) {
	if level < logLevel {
		return
	}
	
	_, file, line, _ := runtime.Caller(2)
	file = filepath.Base(file)
	
	levelStr := map[Level]string{
		DEBUG: "DEBUG",
		INFO:  "INFO",
		WARN:  "WARN",
		ERROR: "ERROR",
		FATAL: "FATAL",
	}[level]
	
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, v...)
	logLine := fmt.Sprintf("[%s] [%s] %s:%d - %s", timestamp, levelStr, file, line, message)
	
	log.Println(logLine)
	
	if level == FATAL {
		os.Exit(1)
	}
}

func Debug(format string, v ...interface{}) {
	logMessage(DEBUG, format, v...)
}

func Info(format string, v ...interface{}) {
	logMessage(INFO, format, v...)
}

func Warn(format string, v ...interface{}) {
	logMessage(WARN, format, v...)
}

func Error(format string, v ...interface{}) {
	logMessage(ERROR, format, v...)
}

func Fatal(format string, v ...interface{}) {
	logMessage(FATAL, format, v...)
}
