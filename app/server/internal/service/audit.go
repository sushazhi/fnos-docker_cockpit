package service

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"os"
	"sync"
	"time"

	"dockpit/internal/config"
	"dockpit/internal/model"
)

type AuditService struct {
	logs []model.AuditLog
	mu   sync.RWMutex
	file *os.File
}

var auditService *AuditService

func InitAuditService() {
	auditService = &AuditService{
		logs: make([]model.AuditLog, 0, config.Get().Audit.MaxLogs),
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
	if s.file == nil {
		f, err := os.OpenFile(config.Get().AuditLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		s.file = f
	}

	data, _ := json.Marshal(log)
	s.file.WriteString(string(data) + "\n")
}

func (s *AuditService) loadFromFile() {
	f, err := os.Open(config.Get().AuditLogFile)
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
