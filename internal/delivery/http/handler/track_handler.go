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
	userID, _ := r.Context().Value("user_id").(string)
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
		UserID:     userID,
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
	userID, _ := r.Context().Value("user_id").(string)
	playlistID := r.URL.Query().Get("playlist_id")
	tracks, err := h.trackUsecase.GetAllReady(userID, playlistID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, tracks)
}

func (h *TrackHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)

	// Max 500 MB upload
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ошибка обработки формы загрузки: файл слишком большой"})
		return
	}

	playlistID := r.FormValue("playlist_id")
	if r.MultipartForm == nil || len(r.MultipartForm.File) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Файлы для загрузки не найдены"})
		return
	}

	var inputs []usecase.UploadTrackInput
	for _, fileHeaders := range r.MultipartForm.File {
		for _, fh := range fileHeaders {
			f, err := fh.Open()
			if err != nil {
				continue
			}
			defer f.Close()

			inputs = append(inputs, usecase.UploadTrackInput{
				Filename: fh.Filename,
				Reader:   f,
				Size:     fh.Size,
			})
		}
	}

	if len(inputs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Не удалось открыть прикрепленные файлы"})
		return
	}

	result, err := h.trackUsecase.UploadTracks(r.Context(), userID, playlistID, inputs)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *TrackHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	track, err := h.trackUsecase.GetByID(id, userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Трек не найден"})
		return
	}
	writeJSON(w, http.StatusOK, track)
}

func (h *TrackHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	stats, err := h.trackUsecase.GetStats(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *TrackHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	id := chi.URLParam(r, "id")
	if err := h.trackUsecase.Delete(id, userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Трек успешно удален"})
}

type batchDeleteRequest struct {
	IDs []string `json:"ids"`
}

func (h *TrackHandler) BatchDelete(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	var req batchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Необходимо указать список идентификаторов треков"})
		return
	}

	if err := h.trackUsecase.BatchDelete(req.IDs, userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"deleted": len(req.IDs),
	})
}
