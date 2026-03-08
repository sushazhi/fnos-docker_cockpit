package service

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"sync"
	"time"

	"dockpit/internal/config"
	"dockpit/internal/model"

	"golang.org/x/crypto/argon2"
)

type SessionService struct {
	sessions   map[string]*model.Session
	csrfTokens map[string]string
	execOwners map[string]string // execID -> sessionToken
	mu         sync.RWMutex
}

var sessionService *SessionService

func InitSessionService() {
	sessionService = &SessionService{
		sessions:   make(map[string]*model.Session),
		csrfTokens: make(map[string]string),
		execOwners: make(map[string]string),
	}
	go sessionService.cleanupExpired()
}

func GetSessionService() *SessionService {
	return sessionService
}

func (s *SessionService) generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *SessionService) CreateSession() *model.Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	token := s.generateToken()
	csrfToken := s.generateToken()

	session := &model.Session{
		Token:      token,
		CSRFToken:  csrfToken,
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
	}

	s.sessions[token] = session
	s.csrfTokens[csrfToken] = token

	return session
}

func (s *SessionService) ValidateSession(token string) bool {
	if token == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[token]
	if !exists {
		return false
	}

	if time.Since(session.CreatedAt) > config.Get().SessionExpiry {
		delete(s.sessions, token)
		delete(s.csrfTokens, session.CSRFToken)
		return false
	}

	session.LastAccess = time.Now()
	return true
}

func (s *SessionService) ValidateCSRF(sessionToken, csrfToken string) bool {
	if sessionToken == "" || csrfToken == "" {
		return false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionToken]
	if !exists {
		return false
	}

	return session.CSRFToken == csrfToken
}

func (s *SessionService) DestroySession(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, exists := s.sessions[token]; exists {
		delete(s.csrfTokens, session.CSRFToken)
	}
	delete(s.sessions, token)
}

func (s *SessionService) GetCSRFToken(sessionToken string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if session, exists := s.sessions[sessionToken]; exists {
		return session.CSRFToken
	}
	return ""
}

func (s *SessionService) cleanupExpired() {
	ticker := time.NewTicker(time.Hour)
	for range ticker.C {
		s.mu.Lock()
		for token, session := range s.sessions {
			if time.Since(session.CreatedAt) > config.Get().SessionExpiry {
				delete(s.sessions, token)
				delete(s.csrfTokens, session.CSRFToken)
				// 清理该会话关联的所有 execID
				for execID, sessionToken := range s.execOwners {
					if sessionToken == token {
						delete(s.execOwners, execID)
					}
				}
			}
		}
		s.mu.Unlock()
	}
}

// RegisterExec 注册 execID 与会话的关联
func (s *SessionService) RegisterExec(sessionToken, execID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.execOwners[execID] = sessionToken
}

// ValidateExecOwnership 验证 execID 是否属于指定会话
func (s *SessionService) ValidateExecOwnership(sessionToken, execID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ownerToken, exists := s.execOwners[execID]
	if !exists {
		return false
	}

	return ownerToken == sessionToken
}

// RemoveExec 移除 execID 记录
func (s *SessionService) RemoveExec(execID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.execOwners, execID)
}

type AuthService struct {
	failedAttempts map[string]*LoginAttempt
	mu             sync.RWMutex
}

type LoginAttempt struct {
	Count     int
	LockedUntil time.Time
	LastAttempt time.Time
}

var authService *AuthService

func InitAuthService() {
	authService = &AuthService{
		failedAttempts: make(map[string]*LoginAttempt),
	}
	go authService.cleanupAttempts()
}

func NewAuthService() *AuthService {
	return authService
}

// cleanupAttempts 定期清理过期的登录尝试记录
func (s *AuthService) cleanupAttempts() {
	ticker := time.NewTicker(time.Hour)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for ip, attempt := range s.failedAttempts {
			// 清理锁定已过期的记录
			if now.After(attempt.LockedUntil) && now.Sub(attempt.LastAttempt) > time.Hour {
				delete(s.failedAttempts, ip)
			}
		}
		s.mu.Unlock()
	}
}

// IsLocked 检查IP是否被锁定
func (s *AuthService) IsLocked(ip string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attempt, exists := s.failedAttempts[ip]
	if !exists {
		return false
	}

	return time.Now().Before(attempt.LockedUntil)
}

// RecordFailedAttempt 记录失败的登录尝试
func (s *AuthService) RecordFailedAttempt(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	attempt, exists := s.failedAttempts[ip]
	if !exists {
		attempt = &LoginAttempt{Count: 0}
		s.failedAttempts[ip] = attempt
	}

	attempt.Count++
	attempt.LastAttempt = time.Now()

	// 达到最大尝试次数，锁定账户
	cfg := config.Get().Login
	if attempt.Count >= cfg.MaxAttempts {
		attempt.LockedUntil = time.Now().Add(cfg.LockoutTime)
	}
}

// ClearFailedAttempts 清除失败尝试记录（登录成功后调用）
func (s *AuthService) ClearFailedAttempts(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.failedAttempts, ip)
}

// GetRemainingAttempts 获取剩余尝试次数
func (s *AuthService) GetRemainingAttempts(ip string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attempt, exists := s.failedAttempts[ip]
	if !exists {
		return config.Get().Login.MaxAttempts
	}

	remaining := config.Get().Login.MaxAttempts - attempt.Count
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (s *AuthService) IsInitialized() bool {
	_, err := os.Stat(config.Get().PasswordFile)
	return err == nil
}

func (s *AuthService) SetPassword(password string) error {
	// 生成随机盐值（16字节）
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	// 使用Argon2id生成密码哈希
	// 参数：time=3, memory=64MB, threads=4, keyLen=32
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

	// 将盐值和哈希一起存储（盐值:哈希）
	combined := append(salt, hash...)

	os.MkdirAll(config.Get().DataDir, 0755)
	return os.WriteFile(config.Get().PasswordFile, combined, 0600)
}

func (s *AuthService) VerifyPassword(password string) bool {
	data, err := os.ReadFile(config.Get().PasswordFile)
	if err != nil {
		return false
	}

	// 新版本格式：48字节（16字节盐值 + 32字节哈希）
	if len(data) == 48 {
		// 提取盐值和存储的哈希
		salt := data[:16]
		storedHash := data[16:]

		// 使用相同的盐值计算输入密码的哈希
		inputHash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)

		// 使用恒定时间比较防止时序攻击
		return compareHashes(storedHash, inputHash)
	}

	return false
}

// compareHashes 使用恒定时间比较两个哈希值，防止时序攻击
func compareHashes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}
