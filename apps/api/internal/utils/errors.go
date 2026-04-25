package utils

import (
	"context"
	"net/http"

	"github.com/lania-smp/backend/internal/logger"
)

// CustomError represents an error with an associated HTTP status code.
type CustomError struct {
	HttpStatus    int
	Message       string
	OriginalError error
}

func (e *CustomError) Error() string {
	return e.Message
}

func MapCustomErrorToHttpStatus(err error) int {
	if customErr, ok := err.(*CustomError); ok {
		return customErr.HttpStatus
	}
	// Default to 500 if the error is unknown
	return http.StatusInternalServerError
}

func ExtractErrorMessage(err error) string {
	if customErr, ok := err.(*CustomError); ok {
		if customErr.Message == "" && customErr.OriginalError != nil {
			return customErr.OriginalError.Error()
		}
		return customErr.Message
	}
	return err.Error()
}

func LogCustomError(ctx context.Context, err error) {
	if customErr, ok := err.(*CustomError); ok {
		logger.Errorf(ctx, "%v: %v", customErr.Message, customErr.OriginalError)
	} else {
		logger.Errorf(ctx, "Error: %v", err)
	}
}

func LogError(ctx context.Context, err error) {
	if customErr, ok := err.(*CustomError); ok {
		logger.Errorf(ctx, "%v: %v\n", customErr.Message, customErr.OriginalError)
	} else {
		logger.Errorf(ctx, "Error: %v\n", err)
	}
}

// Helper functions to create errors
func NewBadRequestError(msg string, err error) error {
	return &CustomError{HttpStatus: http.StatusBadRequest, Message: msg, OriginalError: err}
}

func NewNotFoundError(msg string, err error) error {
	return &CustomError{HttpStatus: http.StatusNotFound, Message: msg, OriginalError: err}
}

func NewUnauthorizedError(msg string, err error) error {
	return &CustomError{HttpStatus: http.StatusUnauthorized, Message: msg, OriginalError: err}
}

func NewForbiddenError(msg string, err error) error {
	return &CustomError{HttpStatus: http.StatusForbidden, Message: msg, OriginalError: err}
}

func NewInternalServerError(msg string, err error) error {
	return &CustomError{HttpStatus: http.StatusInternalServerError, Message: msg, OriginalError: err}
}

func NewConflictError(msg string, err error) error {
	return &CustomError{HttpStatus: http.StatusConflict, Message: msg, OriginalError: err}
}
