package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gobkd/internal/appctx"
)

type Code string

const (
	CodeInvalidRequest   Code = "invalid_request"
	CodeValidationFailed Code = "validation_failed"
	CodeUnauthorized     Code = "unauthorized"
	CodeForbidden        Code = "forbidden"
	CodeNotFound         Code = "not_found"
	CodeMethodNotAllowed Code = "method_not_allowed"
	CodeConflict         Code = "conflict"
	CodeInternalError    Code = "internal_error"
	CodeServiceDown      Code = "service_unavailable"
)

type ErrorField struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
}

type ErrorBody struct {
	Error     Code        `json:"error"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

func OK(c *gin.Context, status int, payload gin.H) {
	c.JSON(status, payload)
}

func Error(c *gin.Context, status int, code Code, message string, details interface{}) {
	body := ErrorBody{
		Error:     code,
		Message:   message,
		Details:   details,
		RequestID: appctx.GetString(c, appctx.RequestIDKey),
	}
	c.AbortWithStatusJSON(status, body)
}

func InvalidRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, CodeInvalidRequest, message, nil)
}

func ValidationFailed(c *gin.Context, details interface{}) {
	Error(c, http.StatusUnprocessableEntity, CodeValidationFailed, "validation failed", details)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, CodeUnauthorized, message, nil)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, CodeForbidden, message, nil)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, CodeNotFound, message, nil)
}

func MethodNotAllowed(c *gin.Context, message string) {
	Error(c, http.StatusMethodNotAllowed, CodeMethodNotAllowed, message, nil)
}

func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, CodeConflict, message, nil)
}

func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, CodeInternalError, message, nil)
}

func ServiceUnavailable(c *gin.Context, message string) {
	Error(c, http.StatusServiceUnavailable, CodeServiceDown, message, nil)
}
