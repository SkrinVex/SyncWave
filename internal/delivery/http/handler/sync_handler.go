package handler

import (
	"net/http"
	"strconv"

	"github.com/syncwave/syncwave/internal/delivery/http/middleware"
	"github.com/syncwave/syncwave/internal/infrastructure/worker"
	"github.com/syncwave/syncwave/internal/usecase"
)

type SyncHandler struct {
	syncUsecase *usecase.SyncUsecase
	eventHub    *worker.EventHub
}

func NewSyncHandler(syncUsecase *usecase.SyncUsecase, eventHub *worker.EventHub) *SyncHandler {
	return &SyncHandler{
		syncUsecase: syncUsecase,
		eventHub:    eventHub,
	}
}

func (h *SyncHandler) TriggerAll(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := h.syncUsecase.TriggerSyncAll(claims.UserID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "sync started for all playlists"})
}

func (h *SyncHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	progress := h.syncUsecase.GetProgress()
	writeJSON(w, http.StatusOK, progress)
}

func (h *SyncHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 100
	}

	logs, err := h.syncUsecase.GetLogs(limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, logs)
}

func (h *SyncHandler) ClearLogs(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 7
	}

	if err := h.syncUsecase.ClearLogs(days); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logs cleaned"})
}

func (h *SyncHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	h.eventHub.ServeHTTP(w, r)
}
