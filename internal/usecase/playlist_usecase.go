package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/infrastructure/worker"
	"github.com/syncwave/syncwave/internal/infrastructure/ytdlp"
)

type PlaylistUsecase struct {
	playlistRepo domain.PlaylistRepository
	ytdlpClient  *ytdlp.Client
	workerQueue  *worker.WorkerQueue
}

func NewPlaylistUsecase(
	playlistRepo domain.PlaylistRepository,
	ytdlpClient *ytdlp.Client,
	workerQueue *worker.WorkerQueue,
) *PlaylistUsecase {
	return &PlaylistUsecase{
		playlistRepo: playlistRepo,
		ytdlpClient:  ytdlpClient,
		workerQueue:  workerQueue,
	}
}

type CreatePlaylistRequest struct {
	Title               string `json:"title"`
	YouTubeURLOrID      string `json:"url_or_id"`
	AutoSync            bool   `json:"auto_sync"`
	SyncIntervalMinutes int    `json:"sync_interval_minutes"`
}

func (u *PlaylistUsecase) Create(userID string, req CreatePlaylistRequest) (*domain.Playlist, error) {
	cleanInput := strings.TrimSpace(req.YouTubeURLOrID)
	if cleanInput == "" {
		return nil, errors.New("Необходимо указать ссылку на плейлист или его ID")
	}

	normalizedInput := ytdlp.NormalizePlaylistURL(cleanInput)

	// Check if already registered
	existing, _ := u.playlistRepo.GetByYouTubeID(cleanInput)
	if existing != nil {
		return nil, errors.New("Этот плейлист уже добавлен в список ваших подписок")
	}
	existingNorm, _ := u.playlistRepo.GetByYouTubeID(normalizedInput)
	if existingNorm != nil {
		return nil, errors.New("Этот плейлист уже добавлен в список ваших подписок")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		if cleanInput == "LM" || cleanInput == "liked" {
			title = "Понравившиеся"
		} else {
			// Try to fetch title quickly
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			out, err := u.ytdlpClient.ExtractPlaylistDelta(ctx, cleanInput)
			if err == nil && out.Title != "" {
				title = out.Title
			} else {
				title = "YouTube Playlist"
			}
		}
	}

	syncInterval := req.SyncIntervalMinutes
	if syncInterval <= 0 {
		syncInterval = 60
	}

	p := &domain.Playlist{
		ID:                  uuid.New().String(),
		UserID:              userID,
		Title:               title,
		YouTubeID:           cleanInput,
		AutoSync:            req.AutoSync,
		SyncIntervalMinutes: syncInterval,
		Status:              domain.PlaylistStatusIdle,
	}

	if err := u.playlistRepo.Create(p); err != nil {
		return nil, err
	}

	// Automatically trigger initial sync in background
	_ = u.workerQueue.Enqueue(p.ID, true)

	return p, nil
}

func (u *PlaylistUsecase) List(userID string) ([]domain.Playlist, error) {
	return u.playlistRepo.ListByUserID(userID)
}

func (u *PlaylistUsecase) GetByID(id string) (*domain.Playlist, error) {
	return u.playlistRepo.GetByID(id)
}

func (u *PlaylistUsecase) Update(id string, req CreatePlaylistRequest) (*domain.Playlist, error) {
	p, err := u.playlistRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		p.Title = req.Title
	}
	if req.YouTubeURLOrID != "" {
		p.YouTubeID = strings.TrimSpace(req.YouTubeURLOrID)
	}
	p.AutoSync = req.AutoSync
	if req.SyncIntervalMinutes > 0 {
		p.SyncIntervalMinutes = req.SyncIntervalMinutes
	}

	if err := u.playlistRepo.Update(p); err != nil {
		return nil, err
	}

	return p, nil
}

func (u *PlaylistUsecase) Delete(id string) error {
	return u.playlistRepo.Delete(id)
}

func (u *PlaylistUsecase) TriggerSync(id string) error {
	p, err := u.playlistRepo.GetByID(id)
	if err != nil {
		return err
	}
	return u.workerQueue.Enqueue(p.ID, true)
}
