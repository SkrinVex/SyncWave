package handler

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/usecase"
)

type AdminHandler struct {
	userRepo        domain.UserRepository
	trackRepo       domain.TrackRepository
	settingsUsecase *usecase.SettingsUsecase
}

func NewAdminHandler(userRepo domain.UserRepository, trackRepo domain.TrackRepository, settingsUsecase *usecase.SettingsUsecase) *AdminHandler {
	return &AdminHandler{
		userRepo:        userRepo,
		trackRepo:       trackRepo,
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

	// 1. Fetch user tracks to clean up files from disk
	var userTracks []domain.Track
	if h.trackRepo != nil {
		userTracks, _ = h.trackRepo.GetAllReady(targetID, "")
	}

	// 2. Delete user from database (cascades playlists, tracks, sync_logs, blacklist)
	if err := h.userRepo.Delete(targetID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// 3. Clean up physical audio files and covers if not used by any other user
	if h.trackRepo != nil && len(userTracks) > 0 {
		for _, track := range userTracks {
			if track.FilePath != "" {
				if count, _ := h.trackRepo.CountTracksByFilePath(track.FilePath); count == 0 {
					_ = os.Remove(track.FilePath)
				}
			}
			if track.CoverPath != "" {
				_ = os.Remove(track.CoverPath)
			}
		}
	}

	// 4. Delete user's isolated cookies.txt
	if h.settingsUsecase != nil {
		_ = h.settingsUsecase.DeleteCookies(targetID)
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
