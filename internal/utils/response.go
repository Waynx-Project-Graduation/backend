package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard API response format
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []FieldError  `json:"details,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Meta struct {
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Total   int64 `json:"total"`
}

// Success sends a 200 JSON response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// Created sends a 201 JSON response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMeta sends a paginated response
func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// BadRequest sends a 400 error
func BadRequest(c *gin.Context, message string, details ...FieldError) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    "BAD_REQUEST",
			Message: message,
			Details: details,
		},
	})
}

// Unauthorized sends a 401 error
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    "UNAUTHORIZED",
			Message: message,
		},
	})
}

// Forbidden sends a 403 error
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    "FORBIDDEN",
			Message: message,
		},
	})
}

// NotFound sends a 404 error
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    "NOT_FOUND",
			Message: message,
		},
	})
}

// InternalError sends a 500 error
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    "INTERNAL_ERROR",
			Message: message,
		},
	})
}

// ServiceUnavailable sends a 503 error
func ServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    "SERVICE_UNAVAILABLE",
			Message: message,
		},
	})
}

// ValidationError sends a 422 error with field details
func ValidationError(c *gin.Context, details []FieldError) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: "Input validation failed",
			Details: details,
		},
	})
}
