package apperr

import (
	"errors"
	"fmt"
	"net/http"
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
	CodeRequestTooLarge  Code = "payload_too_large"
	CodeInternalError    Code = "internal_error"
	CodeServiceDown      Code = "service_unavailable"
)

type Error struct {
	Code    Code
	Message string
	Details any
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func InvalidRequest(message string) error {
	return &Error{Code: CodeInvalidRequest, Message: message}
}

func ValidationFailed(details any) error {
	return &Error{
		Code:    CodeValidationFailed,
		Message: "validation failed",
		Details: details,
	}
}

func Unauthorized(message string) error {
	return &Error{Code: CodeUnauthorized, Message: message}
}

func Forbidden(message string) error {
	return &Error{Code: CodeForbidden, Message: message}
}

func NotFound(message string) error {
	return &Error{Code: CodeNotFound, Message: message}
}

func Conflict(message string) error {
	return &Error{Code: CodeConflict, Message: message}
}

func RequestTooLarge(limit int64) error {
	return &Error{
		Code:    CodeRequestTooLarge,
		Message: fmt.Sprintf("request body exceeds %d bytes", limit),
	}
}

func Internal(message string, err error) error {
	return &Error{
		Code:    CodeInternalError,
		Message: message,
		Err:     err,
	}
}

func ServiceUnavailable(message string, err error) error {
	return &Error{
		Code:    CodeServiceDown,
		Message: message,
		Err:     err,
	}
}

func Status(err error) int {
	var appErr *Error
	if !errors.As(err, &appErr) {
		return http.StatusInternalServerError
	}

	switch appErr.Code {
	case CodeInvalidRequest:
		return http.StatusBadRequest
	case CodeValidationFailed:
		return http.StatusUnprocessableEntity
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case CodeConflict:
		return http.StatusConflict
	case CodeRequestTooLarge:
		return http.StatusRequestEntityTooLarge
	case CodeServiceDown:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func Details(err error) any {
	var appErr *Error
	if !errors.As(err, &appErr) {
		return nil
	}
	return appErr.Details
}

func Message(err error) string {
	var appErr *Error
	if !errors.As(err, &appErr) {
		return "internal server error"
	}
	if appErr.Message != "" {
		return appErr.Message
	}
	return string(appErr.Code)
}

func ErrorCode(err error) Code {
	var appErr *Error
	if !errors.As(err, &appErr) {
		return CodeInternalError
	}
	return appErr.Code
}
