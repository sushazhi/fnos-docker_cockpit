package middleware

import (
	"dockpit/internal/config"
	"dockpit/internal/service"
	"dockpit/pkg/response"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	requests    map[string]*ClientInfo
	authLimit   map[string]*ClientInfo // 敏感端点专用
	mu          sync.RWMutex
	authMu      sync.RWMutex
}

type ClientInfo struct {
	Count     int
	ResetTime time.Time
}

var rateLimiter *RateLimiter

func InitRateLimiter() {
	rateLimiter = &RateLimiter{
		requests:  make(map[string]*ClientInfo),
		authLimit: make(map[string]*ClientInfo),
	}
	// 启动清理goroutine
	go rateLimiter.cleanup()
}

// cleanup 定期清理过期的客户端记录
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, info := range rl.requests {
			if now.After(info.ResetTime) {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()

		rl.authMu.Lock()
		for ip, info := range rl.authLimit {
			if now.After(info.ResetTime) {
				delete(rl.authLimit, ip)
			}
		}
		rl.authMu.Unlock()
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cfg := config.Get().RateLimit

	info, exists := rl.requests[ip]
	if !exists || now.After(info.ResetTime) {
		rl.requests[ip] = &ClientInfo{
			Count:     1,
			ResetTime: now.Add(cfg.WindowMs),
		}
		return true
	}

	if info.Count >= cfg.MaxRequests {
		return false
	}

	info.Count++
	return true
}

// AllowAuth 检查敏感端点是否允许请求（更严格的限制）
func (rl *RateLimiter) AllowAuth(ip string) bool {
	rl.authMu.Lock()
	defer rl.authMu.Unlock()

	now := time.Now()
	cfg := config.Get().RateLimit

	info, exists := rl.authLimit[ip]
	if !exists || now.After(info.ResetTime) {
		rl.authLimit[ip] = &ClientInfo{
			Count:     1,
			ResetTime: now.Add(cfg.WindowMs),
		}
		return true
	}

	if info.Count >= cfg.AuthMaxRequests {
		return false
	}

	info.Count++
	return true
}

// RateLimit 速率限制中间件
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if rateLimiter == nil {
			c.Next()
			return
		}

		// 终端API路径白名单，跳过速率限制
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/exec/") ||
		   strings.HasPrefix(path, "/api/container/") && strings.HasSuffix(path, "/terminal") {
			c.Next()
			return
		}

		ip := getClientIP(c)
		if !rateLimiter.Allow(ip) {
			response.Error(c, 429, "RATE_LIMIT_EXCEEDED", "Too many requests. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRateLimit 敏感端点速率限制中间件（登录、密码修改等）
func AuthRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if rateLimiter == nil {
			c.Next()
			return
		}

		ip := getClientIP(c)
		if !rateLimiter.AllowAuth(ip) {
			response.Error(c, 429, "RATE_LIMIT_EXCEEDED", "Too many requests. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

func getSessionToken(c *gin.Context) string {
	if token, err := c.Cookie("session_token"); err == nil && token != "" {
		return token
	}

	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 移除URL参数传递token的方式，避免安全风险
	return ""
}

func getClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	return c.ClientIP()
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// WebSocket 路径跳过认证（已在创建 exec 时验证过）
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/exec/") && strings.HasSuffix(path, "/ws") {
			c.Next()
			return
		}

		token := getSessionToken(c)
		clientIP := getClientIP(c)

		c.Set("clientIP", clientIP)

		if token == "" {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		if !service.GetSessionService().ValidateSession(token) {
			response.Unauthorized(c, "Session expired")
			c.Abort()
			return
		}

		c.Set("sessionToken", token)
		c.Next()
	}
}

func CSRFRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" {
			c.Next()
			return
		}

		sessionToken, exists := c.Get("sessionToken")
		if !exists {
			response.Forbidden(c, "CSRF validation failed")
			c.Abort()
			return
		}

		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			csrfToken = c.PostForm("_csrf")
		}

		if !service.GetSessionService().ValidateCSRF(sessionToken.(string), csrfToken) {
			response.Forbidden(c, "CSRF validation failed")
			c.Abort()
			return
		}

		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 基础安全头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")

		// HSTS - 强制 HTTPS（仅在启用时添加）
		// 通过环境变量 HSTS_ENABLED=true 启用
		// 注意：如果使用反向代理，应在反向代理层配置 HSTS
		if os.Getenv("HSTS_ENABLED") == "true" {
			maxAge := "31536000" // 1年
			c.Header("Strict-Transport-Security", "max-age="+maxAge+"; includeSubDomains; preload")
		}

		// Referrer 策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=()")
		
		// 跨域资源策略
		// 注意：COOP/COEP 在非 HTTPS 环境下不被浏览器支持
		// 仅在 HTTPS 或 localhost 环境下启用
		isSecureContext := c.Request.TLS != nil ||
			strings.HasPrefix(c.Request.Host, "localhost:") ||
			strings.HasPrefix(c.Request.Host, "127.0.0.1:")

		if isSecureContext {
			c.Header("Cross-Origin-Opener-Policy", "same-origin")
			c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		}
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// X-Frame-Options: 允许同源和特定域名嵌入
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = c.GetHeader("Referer")
		}

		// 从环境变量读取允许嵌入的域名
		allowedFrameAncestors := os.Getenv("ALLOWED_FRAME_ANCESTORS")
		if allowedFrameAncestors == "" {
			// 内网部署场景：默认允许所有来源嵌入
			// 这是因为内网环境中，父页面可能来自不同端口或路径
			// 如果需要更严格的安全控制，请配置 ALLOWED_FRAME_ANCESTORS 环境变量
			allowedFrameAncestors = "*"
		}
		
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// CSP 策略
		// 注意：Vue 的模板编译器需要 'unsafe-eval'，生产环境建议预编译模板
		csp := fmt.Sprintf(
			"default-src 'self'; "+
			"script-src 'self' 'unsafe-eval'; "+ // Vue 模板编译需要 unsafe-eval
			"style-src 'self' 'unsafe-inline'; "+ // Vue 样式需要 unsafe-inline
			"img-src 'self' data: blob: https:; "+
			"font-src 'self' data:; "+
			"connect-src 'self' ws: wss: https://api.github.com; "+
			"frame-ancestors %s; "+
			"base-uri 'self'; "+
			"form-action 'self'; "+
			"object-src 'none';",
			allowedFrameAncestors,
		)
		c.Header("Content-Security-Policy", csp)

		c.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 从环境变量读取允许的域名
		envOrigins := os.Getenv("ALLOWED_ORIGINS")

		// 生产环境：必须通过环境变量配置允许的域名
		// 开发环境：如果未配置 ALLOWED_ORIGINS，则允许 localhost
		var allowedOrigins []string
		if envOrigins != "" {
			// 生产模式：只使用环境变量配置的域名
			allowedOrigins = strings.Split(envOrigins, ",")
		} else {
			// 开发模式：允许 localhost（仅在未配置 ALLOWED_ORIGINS 时）
			log.Println("[警告] 未配置 ALLOWED_ORIGINS 环境变量，使用开发模式 CORS 设置。生产环境请配置 ALLOWED_ORIGINS")
			allowedOrigins = []string{
				"http://localhost:3000",
				"http://localhost:8807",
				"http://127.0.0.1:3000",
				"http://127.0.0.1:8807",
			}
		}

		// 检查origin是否在允许列表中
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		// 如果是同源请求（origin为空或与host相同），也允许
		if origin == "" || origin == "http://"+c.Request.Host || origin == "https://"+c.Request.Host {
			allowed = true
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-CSRF-Token")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// HTTPSRedirect 强制 HTTPS 重定向中间件
func HTTPSRedirect() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用 HTTPS 强制重定向
		if os.Getenv("FORCE_HTTPS") != "true" {
			c.Next()
			return
		}

		// 检查是否已经是 HTTPS
		if c.Request.TLS != nil {
			c.Next()
			return
		}

		// 检查 X-Forwarded-Proto 头（用于反向代理）
		if c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Next()
			return
		}

		// 重定向到 HTTPS
		host := c.Request.Host
		if host == "" {
			host = c.GetHeader("Host")
		}
		
		// 移除端口号（如果有）
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		
		httpsURL := "https://" + host + c.Request.RequestURI
		c.Redirect(301, httpsURL)
		c.Abort()
	}
}
