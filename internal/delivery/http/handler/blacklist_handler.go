package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/syncwave/syncwave/internal/domain"
)

type BlacklistHandler struct {
	blacklistRepo domain.BlacklistRepository
}

func NewBlacklistHandler(blacklistRepo domain.BlacklistRepository) *BlacklistHandler {
	return &BlacklistHandler{
		blacklistRepo: blacklistRepo,
	}
}

func (h *BlacklistHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	query := r.URL.Query().Get("q")
	items, err := h.blacklistRepo.List(userID, query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *BlacklistHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	var item domain.BlacklistItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		return
	}

	if item.YouTubeID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "youtube_id is required"})
		return
	}

	item.UserID = userID
	if err := h.blacklistRepo.Add(&item); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *BlacklistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	if id == "" {
		id = chi.URLParam(r, "youtube_id")
	}
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}

	if err := h.blacklistRepo.Remove(id, userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BlacklistHandler) Remove(w http.ResponseWriter, r *http.Request) {
	h.Delete(w, r)
}
