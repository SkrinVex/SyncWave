package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/usecase"
)

type AdminHandler struct {
	userRepo        domain.UserRepository
	settingsUsecase *usecase.SettingsUsecase
}

func NewAdminHandler(userRepo domain.UserRepository, settingsUsecase *usecase.SettingsUsecase) *AdminHandler {
	return &AdminHandler{
		userRepo:        userRepo,
		settingsUsecase: settingsUsecase,
	}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.ListWithStats()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, users)
}

type updateQuotaReq struct {
	QuotaBytes int64 `json:"quota_bytes"`
}

func (h *AdminHandler) UpdateUserQuota(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user id required"})
		return
	}

	var req updateQuotaReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
		return
	}

	if err := h.userRepo.UpdateQuota(userID, req.QuotaBytes); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value("user_id").(string)
	targetID := chi.URLParam(r, "id")
	if targetID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user id required"})
		return
	}

	if currentUserID == targetID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot delete your own administrator account"})
		return
	}

	if err := h.userRepo.Delete(targetID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type setRegistrationReq struct {
	Allow bool `json:"allow"`
}

func (h *AdminHandler) SetRegistration(w http.ResponseWriter, r *http.Request) {
	var req setRegistrationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
		return
	}

	if err := h.settingsUsecase.SetAllowRegistration(req.Allow); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type setGlobalLimitReq struct {
	LimitBytes int64 `json:"limit_bytes"`
}

func (h *AdminHandler) SetGlobalLimit(w http.ResponseWriter, r *http.Request) {
	var req setGlobalLimitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
		return
	}

	if err := h.settingsUsecase.SetGlobalStorageLimit(req.LimitBytes); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
