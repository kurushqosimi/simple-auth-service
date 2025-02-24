package http

import (
	"fullstack-simple-app/internal/errcode"
	"github.com/gin-gonic/gin"
	"net/http"
)

var codeToHTTPStatus = map[string]int{
	errcode.ErrInvalidRequest:     http.StatusBadRequest,          // 400
	errcode.ErrUnauthorized:       http.StatusUnauthorized,        // 401
	errcode.ErrForbidden:          http.StatusForbidden,           // 403
	errcode.ErrNotFound:           http.StatusNotFound,            // 404
	errcode.ErrConflict:           http.StatusConflict,            // 409
	errcode.ErrInternal:           http.StatusInternalServerError, // 500
	errcode.ErrEmailAlreadyExists: http.StatusConflict,            // 409
	errcode.ErrAccountCreated:     http.StatusCreated,             // 201
	errcode.ErrOTPNotFound:        http.StatusNotFound,            // 404
	errcode.ErrOTPInvalid:         http.StatusBadRequest,          // 400
	errcode.ErrInvalidPassword:    http.StatusUnauthorized,        // 401
	errcode.ErrLoginRedirect:      http.StatusFound,               // 302
}

func statusFromCode(code string) int {
	if s, ok := codeToHTTPStatus[code]; ok {
		return s
	}
	// default fallback
	return http.StatusBadRequest
}

func respondWithError(ctx *gin.Context, statusCode int, errorCode string, overrideMsg string, err error) {
	msg := errcode.GetErrorMessage(errorCode)
	if overrideMsg != "" {
		msg = overrideMsg
	}

	payload := gin.H{
		"error":   errorCode,
		"message": msg,
	}

	if err != nil {
		payload["details"] = err.Error()
	}

	if reqID := ctx.GetString("request_id"); reqID != "" {
		payload["request_id"] = reqID
	}
	ctx.JSON(statusCode, payload)
}
