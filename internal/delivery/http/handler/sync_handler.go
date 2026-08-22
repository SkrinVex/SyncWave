package handler

import (
	"net/http"
	"strconv"

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
	userID, _ := r.Context().Value("user_id").(string)
	if userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	if err := h.syncUsecase.TriggerSyncAll(userID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Синхронизация запущена для всех плейлистов"})
}

func (h *SyncHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	h.syncUsecase.CancelSync()
	writeJSON(w, http.StatusOK, map[string]string{"message": "Синхронизация отменена"})
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
	daysStr := r.URL.Query().Get("days")
	days := 0
	if daysStr != "" {
		days, _ = strconv.Atoi(daysStr)
	}

	if err := h.syncUsecase.ClearLogs(days); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Журнал логов очищен"})
}

func (h *SyncHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	h.eventHub.ServeHTTP(w, r)
}
