package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/syncwave/syncwave/internal/usecase"
)

type PlaylistHandler struct {
	playlistUsecase *usecase.PlaylistUsecase
	syncUsecase     *usecase.SyncUsecase
}

func NewPlaylistHandler(
	playlistUsecase *usecase.PlaylistUsecase,
	syncUsecase *usecase.SyncUsecase,
) *PlaylistHandler {
	return &PlaylistHandler{
		playlistUsecase: playlistUsecase,
		syncUsecase:     syncUsecase,
	}
}

type createPlaylistRequest struct {
	Title               string `json:"title"`
	URLOrID             string `json:"url_or_id"`
	AutoSync            bool   `json:"auto_sync"`
	SyncIntervalMinutes int    `json:"sync_interval_minutes"`
}

func (h *PlaylistHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	playlists, err := h.playlistUsecase.List(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, playlists)
}

func (h *PlaylistHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	var req createPlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный запрос"})
		return
	}

	pl, err := h.playlistUsecase.Create(
		userID,
		usecase.CreatePlaylistRequest{
			Title:               req.Title,
			YouTubeURLOrID:      req.URLOrID,
			AutoSync:            req.AutoSync,
			SyncIntervalMinutes: req.SyncIntervalMinutes,
		},
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, pl)
}

func (h *PlaylistHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	pl, err := h.playlistUsecase.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Плейлист не найден"})
		return
	}

	writeJSON(w, http.StatusOK, pl)
}

func (h *PlaylistHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req createPlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный запрос"})
		return
	}

	pl, err := h.playlistUsecase.Update(
		id,
		usecase.CreatePlaylistRequest{
			Title:               req.Title,
			YouTubeURLOrID:      req.URLOrID,
			AutoSync:            req.AutoSync,
			SyncIntervalMinutes: req.SyncIntervalMinutes,
		},
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, pl)
}

func (h *PlaylistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.playlistUsecase.Delete(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PlaylistHandler) Sync(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.syncUsecase.TriggerSyncPlaylist(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "queued"})
}
