package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/usecase"
)

type TrackHandler struct {
	trackUsecase *usecase.TrackUsecase
}

func NewTrackHandler(trackUsecase *usecase.TrackUsecase) *TrackHandler {
	return &TrackHandler{trackUsecase: trackUsecase}
}

func (h *TrackHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 {
		pageSize = 50
	}

	filter := domain.TrackFilter{
		Query:      q.Get("q"),
		PlaylistID: q.Get("playlist_id"),
		Status:     domain.TrackStatus(q.Get("status")),
		SortBy:     q.Get("sort_by"),
		Order:      q.Get("order"),
		Page:       page,
		PageSize:   pageSize,
	}

	result, err := h.trackUsecase.List(filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *TrackHandler) GetAllReady(w http.ResponseWriter, r *http.Request) {
	tracks, err := h.trackUsecase.GetAllReady()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, tracks)
}

func (h *TrackHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	track, err := h.trackUsecase.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Трек не найден"})
		return
	}
	writeJSON(w, http.StatusOK, track)
}

func (h *TrackHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.trackUsecase.GetStats()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *TrackHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.trackUsecase.Delete(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Трек успешно удален"})
}

type batchDeleteRequest struct {
	IDs []string `json:"ids"`
}

func (h *TrackHandler) BatchDelete(w http.ResponseWriter, r *http.Request) {
	var req batchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Необходимо указать список идентификаторов треков"})
		return
	}

	if err := h.trackUsecase.BatchDelete(req.IDs); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"deleted": len(req.IDs),
	})
}
