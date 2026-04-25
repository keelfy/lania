package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

type StatusHandler interface {
	Health(w http.ResponseWriter, r *http.Request)
}

type statusHandler struct {
	sqlDb storage.MainStorage
	cache storage.CacheStorage
}

func NewStatusHandler(sqlDb storage.MainStorage, cache storage.CacheStorage) StatusHandler {
	return &statusHandler{
		sqlDb: sqlDb,
		cache: cache,
	}
}

// @Summary Get system health status
// @Description Get the health status of all system components including API, Database, Search, and Cache
// @Tags status
// @Accept json
// @Produce json
// @Success 200 {object} responses.StatusRes
// @Failure 500 {object} responses.StatusRes
// @Router /v1/health [get]
func (handler *statusHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	response := &responses.StatusRes{
		API:      "OK",
		Database: "OK",
		Cache:    "OK",
	}
	statusCode := http.StatusOK

	err := handler.sqlDb.Ping(ctx)
	if err != nil {
		response.Database = "ERROR"
		statusCode = http.StatusInternalServerError
	}

	err = handler.cache.Ping(ctx)
	if err != nil {
		response.Cache = "ERROR"
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set(utils.HeaderContentType, utils.ApplicationJsonType)
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Errorf(ctx, "Error writing data: %v", err)
	}
}
