package handler

import (
	"dockpit/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

func addAuditLog(c *gin.Context, action string, details map[string]interface{}) {
	clientIP, _ := c.Get("clientIP")
	userAgent := c.GetHeader("User-Agent")
	
	ipStr := ""
	if clientIP != nil {
		ipStr = clientIP.(string)
	}
	
	service.GetAuditService().AddLog(action, details, ipStr, userAgent)
}

func parseInt(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
