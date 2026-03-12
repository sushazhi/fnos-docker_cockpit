package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Config struct {
	Port           int
	DataDir        string
	SessionExpiry  time.Duration
	RateLimit      RateLimitConfig
	Login          LoginConfig
	Audit          AuditConfig
	CSRF           CSRFConfig
	Docker         DockerConfig
	PasswordFile   string
	AuditLogFile   string
}

type RateLimitConfig struct {
	WindowMs        time.Duration
	MaxRequests     int
	AuthMaxRequests int // 敏感端点（登录等）的速率限制
}

type LoginConfig struct {
	MaxAttempts int
	LockoutTime time.Duration
}

type AuditConfig struct {
	MaxLogs int
}

type CSRFConfig struct {
	Expiry time.Duration
}

type DockerConfig struct {
	SocketPath string
	Timeout    time.Duration
}

var cfg *Config

func Init() {
	dataDir := getEnvString("DOCKPIT_DATA_DIR", "")
	if dataDir == "" {
		if _, err := os.Stat("/vol1/@appdata"); err == nil {
			dataDir = "/vol1/@appdata/dockpit"
		} else {
			dataDir = filepath.Join(os.TempDir(), "dockpit")
		}
	}
	
	cfg = &Config{
		Port:          getEnvInt("PORT", 8807),
		DataDir:       dataDir,
		SessionExpiry: 24 * time.Hour,
		RateLimit: RateLimitConfig{
			WindowMs:        time.Minute,
			MaxRequests:     300, // 通用端点每分钟300次
			AuthMaxRequests: 20,  // 敏感端点每分钟20次
		},
		Login: LoginConfig{
			MaxAttempts: 5,
			LockoutTime: 30 * time.Minute,
		},
		Audit: AuditConfig{
			MaxLogs: 1000,
		},
		CSRF: CSRFConfig{
			Expiry: 2 * time.Hour,
		},
		Docker: DockerConfig{
			SocketPath: "/var/run/docker.sock",
			Timeout:    time.Minute,
		},
	}

	cfg.PasswordFile = filepath.Join(cfg.DataDir, ".password")
	cfg.AuditLogFile = filepath.Join(cfg.DataDir, "audit.log")

	os.MkdirAll(cfg.DataDir, 0755)
}

func Get() *Config {
	return cfg
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
