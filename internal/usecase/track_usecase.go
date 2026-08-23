package usecase

import (
	"os"

	"github.com/syncwave/syncwave/internal/domain"
)

type TrackUsecase struct {
	trackRepo     domain.TrackRepository
	blacklistRepo domain.BlacklistRepository
}

func NewTrackUsecase(trackRepo domain.TrackRepository, blacklistRepo domain.BlacklistRepository) *TrackUsecase {
	return &TrackUsecase{
		trackRepo:     trackRepo,
		blacklistRepo: blacklistRepo,
	}
}

func (u *TrackUsecase) List(filter domain.TrackFilter) (*domain.TrackListResult, error) {
	return u.trackRepo.List(filter)
}

func (u *TrackUsecase) GetByID(id string, userID string) (*domain.Track, error) {
	return u.trackRepo.GetByID(id, userID)
}

func (u *TrackUsecase) GetAllReady(userID string) ([]domain.Track, error) {
	return u.trackRepo.GetAllReady(userID)
}

func (u *TrackUsecase) Delete(id string, userID string) error {
	track, err := u.trackRepo.GetByID(id, userID)
	if err != nil {
		return err
	}

	// Add to blacklist before deletion
	_ = u.blacklistRepo.Add(&domain.BlacklistItem{
		YouTubeID: track.YouTubeID,
		UserID:    userID,
		Title:     track.Title,
		Artist:    track.Artist,
	})

	// Remove physical files if no other user is referencing them
	if track.FilePath != "" {
		_ = os.Remove(track.FilePath)
	}
	if track.CoverPath != "" {
		_ = os.Remove(track.CoverPath)
	}

	return u.trackRepo.Delete(id, userID)
}

func (u *TrackUsecase) BatchDelete(ids []string, userID string) error {
	for _, id := range ids {
		if track, err := u.trackRepo.GetByID(id, userID); err == nil && track != nil {
			// Add to blacklist before deletion
			_ = u.blacklistRepo.Add(&domain.BlacklistItem{
				YouTubeID: track.YouTubeID,
				UserID:    userID,
				Title:     track.Title,
				Artist:    track.Artist,
			})

			if track.FilePath != "" {
				_ = os.Remove(track.FilePath)
			}
			if track.CoverPath != "" {
				_ = os.Remove(track.CoverPath)
			}
		}
	}
	return u.trackRepo.BatchDelete(ids, userID)
}

func (u *TrackUsecase) CleanBrokenTracks() error {
	return u.trackRepo.CleanBrokenTracks()
}

func (u *TrackUsecase) GetStats(userID string) (*domain.TrackStats, error) {
	return u.trackRepo.GetStats(userID)
}
