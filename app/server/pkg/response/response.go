package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    string      `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   message,
		Code:    code,
	})
}

func ErrorWithDetails(c *gin.Context, statusCode int, code, message string, details interface{}) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   message,
		Code:    code,
		Details: details,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "AUTH_REQUIRED", message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func DockerError(c *gin.Context, message string, details interface{}) {
	ErrorWithDetails(c, http.StatusInternalServerError, "DOCKER_ERROR", message, details)
}
