package utils

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/lania-smp/backend/internal/logger"
)

const HeaderContentType = "Content-Type"
const ApplicationJsonType = "application/json"

func CalculateTotalPages(totalElements int64, size int) int64 {
	totalPages := totalElements / int64(size)
	if totalElements%int64(size) > 0 {
		totalPages++
	}
	return totalPages
}

func LogAndWriteError(ctx context.Context, w http.ResponseWriter, err error) {
	LogCustomError(ctx, err)
	status := MapCustomErrorToHttpStatus(err)
	msg := ExtractErrorMessage(err)
	if msg == "" {
		msg = http.StatusText(status)
	}
	http.Error(w, msg, status)
}

func HttpError(ctx context.Context, w http.ResponseWriter, err error) {
	LogError(ctx, err)
	status := MapCustomErrorToHttpStatus(err)
	msg := ExtractErrorMessage(err)
	if msg == "" {
		msg = http.StatusText(status)
	}
	http.Error(w, msg, status)
}

func HttpBusinessError(ctx context.Context, w http.ResponseWriter, msg string, status int) {
	err := &CustomError{HttpStatus: status, Message: msg, OriginalError: nil}
	HttpError(ctx, w, err)
}

func WriteHttpJsonResponse[T any](ctx context.Context, w http.ResponseWriter, res T) {
	w.Header().Set(HeaderContentType, ApplicationJsonType)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		logger.Errorf(ctx, "Error data marshalling: %v", err)
	}
}
