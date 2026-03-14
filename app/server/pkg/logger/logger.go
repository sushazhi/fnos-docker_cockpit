package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
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

const (
	maxLogSize  = 10 * 1024 * 1024 // 10MB 后轮转
	maxLogFiles = 3                // 保留3个历史文件
)

var (
	logLevel    = INFO
	logFile     *os.File
	currentSize int64
	dataDirPath string
	mu          sync.Mutex
)

func Init(dataDir string, level Level) error {
	logLevel = level
	dataDirPath = dataDir

	if dataDir != "" {
		logPath := filepath.Join(dataDir, "app.log")
		var err error
		logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}

		// 获取当前文件大小
		if stat, err := logFile.Stat(); err == nil {
			currentSize = stat.Size()
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

func rotateLog() {
	if logFile != nil {
		logFile.Close()
	}

	logPath := filepath.Join(dataDirPath, "app.log")

	// 删除最老的文件
	oldest := filepath.Join(dataDirPath, fmt.Sprintf("app.log.%d", maxLogFiles))
	os.Remove(oldest)

	// 重命名现有备份文件
	for i := maxLogFiles - 1; i >= 1; i-- {
		oldPath := filepath.Join(dataDirPath, fmt.Sprintf("app.log.%d", i))
		newPath := filepath.Join(dataDirPath, fmt.Sprintf("app.log.%d", i+1))
		os.Rename(oldPath, newPath)
	}

	// 重命名当前日志文件
	os.Rename(logPath, filepath.Join(dataDirPath, "app.log.1"))

	// 创建新的日志文件
	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
		currentSize = 0
	}
}

func logMessage(level Level, format string, v ...interface{}) {
	if level < logLevel {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// 检查是否需要轮转
	if currentSize >= maxLogSize {
		rotateLog()
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
	currentSize += int64(len(logLine) + 1) // +1 for newline

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
