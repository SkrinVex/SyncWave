package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/skrinvex/SyncWave/internal/domain"
)

type BlacklistHandler struct {
	repo domain.BlacklistRepository
}

func NewBlacklistHandler(repo domain.BlacklistRepository) *BlacklistHandler {
	return &BlacklistHandler{repo: repo}
}

func (h *BlacklistHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	items, err := h.repo.List(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *BlacklistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	youtubeID := chi.URLParam(r, "id")
	if youtubeID == "" {
		http.Error(w, "missing youtube_id", http.StatusBadRequest)
		return
	}

	if err := h.repo.Remove(youtubeID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
