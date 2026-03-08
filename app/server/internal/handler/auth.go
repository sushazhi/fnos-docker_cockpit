package handler

import (
	"dockpit/internal/service"
	"dockpit/pkg/response"
	"fmt"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

func (h *AuthHandler) CheckInit(c *gin.Context) {
	initialized := h.authService.IsInitialized()
	response.Success(c, gin.H{"initialized": initialized})
}

func (h *AuthHandler) Setup(c *gin.Context) {
	if h.authService.IsInitialized() {
		response.BadRequest(c, "System already initialized")
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Password required")
		return
	}

	if len(req.Password) < 8 {
		response.BadRequest(c, "Password must be at least 8 characters")
		return
	}

	if err := h.authService.SetPassword(req.Password); err != nil {
		response.InternalError(c, "Failed to set password")
		return
	}

	session := service.GetSessionService().CreateSession()
	setSessionCookie(c, session.Token)

	response.Success(c, gin.H{
		"success":   true,
		"csrfToken": session.CSRFToken,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	if !h.authService.IsInitialized() {
		response.BadRequest(c, "System not initialized")
		return
	}

	// 获取客户端IP
	clientIP := c.GetString("clientIP")
	if clientIP == "" {
		clientIP = c.ClientIP()
	}
	// 将 clientIP 设置到 Context 中，供 addAuditLog 使用
	c.Set("clientIP", clientIP)

	// 检查是否被锁定
	if h.authService.IsLocked(clientIP) {
		addAuditLog(c, "login_locked", map[string]interface{}{"ip": clientIP})
		response.Error(c, 429, "ACCOUNT_LOCKED", "Too many failed attempts. Please try again later.")
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Password required")
		return
	}

	if !h.authService.VerifyPassword(req.Password) {
		// 记录失败尝试
		h.authService.RecordFailedAttempt(clientIP)
		remaining := h.authService.GetRemainingAttempts(clientIP)

		addAuditLog(c, "login_failed", map[string]interface{}{
			"ip":        clientIP,
			"remaining": remaining,
		})

		response.Unauthorized(c, fmt.Sprintf("Invalid password. %d attempts remaining.", remaining))
		return
	}

	// 登录成功，清除失败记录
	h.authService.ClearFailedAttempts(clientIP)

	session := service.GetSessionService().CreateSession()
	setSessionCookie(c, session.Token)

	addAuditLog(c, "login_success", map[string]interface{}{"ip": clientIP})
	response.Success(c, gin.H{
		"success":   true,
		"csrfToken": session.CSRFToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token, exists := c.Get("sessionToken")
	if exists {
		service.GetSessionService().DestroySession(token.(string))
	}

	clearSessionCookie(c)
	response.Success(c, gin.H{"success": true})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid parameters")
		return
	}

	if !h.authService.VerifyPassword(req.CurrentPassword) {
		response.Unauthorized(c, "Current password is incorrect")
		return
	}

	if len(req.NewPassword) < 8 {
		response.BadRequest(c, "New password must be at least 8 characters")
		return
	}

	if err := h.authService.SetPassword(req.NewPassword); err != nil {
		response.InternalError(c, "Failed to change password")
		return
	}

	addAuditLog(c, "password_change", nil)
	response.Success(c, gin.H{"success": true})
}

func (h *AuthHandler) GetCSRFToken(c *gin.Context) {
	token, exists := c.Get("sessionToken")
	if !exists {
		response.Unauthorized(c, "Not authenticated")
		return
	}

	csrfToken := service.GetSessionService().GetCSRFToken(token.(string))
	response.Success(c, gin.H{"csrfToken": csrfToken})
}

func (h *AuthHandler) Me(c *gin.Context) {
	response.Success(c, gin.H{
		"authenticated": true,
	})
}

func setSessionCookie(c *gin.Context, token string) {
	c.SetCookie("session_token", token, 86400, "/", "", false, true)
}

func clearSessionCookie(c *gin.Context) {
	c.SetCookie("session_token", "", -1, "/", "", false, true)
}

type LoginRequest struct {
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}

type AuditHandler struct{}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{}
}

func (h *AuditHandler) List(c *gin.Context) {
	limit := parseInt(c.Query("limit"), 100)
	logs := service.GetAuditService().GetLogs(limit)
	response.Success(c, gin.H{"logs": logs})
}
