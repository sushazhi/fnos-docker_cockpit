package service

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"dockpit/internal/config"
	"dockpit/internal/model"
)

const (
	maxAuditLogSize  = 10 * 1024 * 1024 // 10MB 后轮转
	maxAuditLogFiles = 3                // 保留3个历史文件
)

type AuditService struct {
	logs        []model.AuditLog
	mu          sync.RWMutex
	file        *os.File
	currentSize int64
	filePath    string
}

var auditService *AuditService

func InitAuditService() {
	cfg := config.Get()
	auditService = &AuditService{
		logs:     make([]model.AuditLog, 0, cfg.Audit.MaxLogs),
		filePath: cfg.AuditLogFile,
	}
	auditService.loadFromFile()
}

func GetAuditService() *AuditService {
	return auditService
}

func (s *AuditService) AddLog(action string, details map[string]interface{}, clientIP, userAgent string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log := model.AuditLog{
		ID:        generateLogID(),
		Action:    action,
		Timestamp: time.Now(),
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Details:   details,
	}

	s.logs = append(s.logs, log)

	if len(s.logs) > config.Get().Audit.MaxLogs {
		s.logs = s.logs[len(s.logs)-config.Get().Audit.MaxLogs:]
	}

	s.writeToFile(log)
}

func (s *AuditService) GetLogs(limit int) []model.AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.logs) {
		limit = len(s.logs)
	}

	// 获取最新的日志（倒序，最新的在前）
	result := make([]model.AuditLog, limit)
	for i := 0; i < limit; i++ {
		result[i] = s.logs[len(s.logs)-1-i]
	}
	return result
}

func (s *AuditService) writeToFile(log model.AuditLog) {
	// 检查是否需要轮转
	if s.currentSize >= maxAuditLogSize {
		s.rotateFile()
	}

	if s.file == nil {
		f, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		s.file = f
	}

	data, _ := json.Marshal(log)
	line := string(data) + "\n"
	s.file.WriteString(line)
	s.currentSize += int64(len(line))
}

func (s *AuditService) rotateFile() {
	if s.file != nil {
		s.file.Close()
		s.file = nil
	}

	// 删除最老的文件
	oldest := fmt.Sprintf("%s.%d", s.filePath, maxAuditLogFiles)
	os.Remove(oldest)

	// 重命名现有备份文件
	for i := maxAuditLogFiles - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", s.filePath, i)
		newPath := fmt.Sprintf("%s.%d", s.filePath, i+1)
		os.Rename(oldPath, newPath)
	}

	// 重命名当前日志文件
	os.Rename(s.filePath, fmt.Sprintf("%s.1", s.filePath))

	s.currentSize = 0
}

func (s *AuditService) loadFromFile() {
	// 获取文件大小
	if stat, err := os.Stat(s.filePath); err == nil {
		s.currentSize = stat.Size()
	}

	f, err := os.Open(s.filePath)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var log model.AuditLog
		if err := json.Unmarshal(scanner.Bytes(), &log); err == nil {
			s.logs = append(s.logs, log)
		}
	}

	if len(s.logs) > config.Get().Audit.MaxLogs {
		s.logs = s.logs[len(s.logs)-config.Get().Audit.MaxLogs:]
	}
}

func generateLogID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	rand.Read(b)
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}
