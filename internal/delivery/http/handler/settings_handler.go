package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/syncwave/syncwave/internal/usecase"
)

type SettingsHandler struct {
	settingsUsecase *usecase.SettingsUsecase
}

func NewSettingsHandler(settingsUsecase *usecase.SettingsUsecase) *SettingsHandler {
	return &SettingsHandler{settingsUsecase: settingsUsecase}
}

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	settings, err := h.settingsUsecase.GetSystemSettings(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, settings)
}

func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req usecase.UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Некорректный запрос"})
		return
	}

	if err := h.settingsUsecase.UpdateSettings(req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	userID, _ := r.Context().Value("user_id").(string)
	settings, _ := h.settingsUsecase.GetSystemSettings(userID)
	writeJSON(w, http.StatusOK, settings)
}

func (h *SettingsHandler) UploadCookies(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	_ = r.ParseMultipartForm(10 << 20) // 10MB

	var content []byte
	file, _, err := r.FormFile("cookies")
	if err == nil {
		defer file.Close()
		content, err = io.ReadAll(file)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Не удалось прочитать файл cookies"})
			return
		}
	} else {
		content, err = io.ReadAll(r.Body)
		if err != nil || len(content) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Файл cookies.txt пуст"})
			return
		}
	}

	if err := h.settingsUsecase.SaveCookies(userID, content); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *SettingsHandler) DeleteCookies(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("user_id").(string)
	if err := h.settingsUsecase.DeleteCookies(userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SettingsHandler) TestProxy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProxyURL string `json:"proxy_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProxyURL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Необходимо указать proxy_url"})
		return
	}

	if err := h.settingsUsecase.TestProxy(req.ProxyURL); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}
