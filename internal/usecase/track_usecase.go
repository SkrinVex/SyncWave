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

func (u *TrackUsecase) GetByID(id string) (*domain.Track, error) {
	return u.trackRepo.GetByID(id)
}

func (u *TrackUsecase) GetAllReady() ([]domain.Track, error) {
	return u.trackRepo.GetAllReady()
}

func (u *TrackUsecase) Delete(id string) error {
	track, err := u.trackRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Add to blacklist before deletion
	_ = u.blacklistRepo.Add(&domain.BlacklistItem{
		YouTubeID: track.YouTubeID,
		Title:     track.Title,
		Artist:    track.Artist,
	})

	// Remove physical files
	if track.FilePath != "" {
		_ = os.Remove(track.FilePath)
	}
	if track.CoverPath != "" {
		_ = os.Remove(track.CoverPath)
	}

	return u.trackRepo.Delete(id)
}

func (u *TrackUsecase) BatchDelete(ids []string) error {
	for _, id := range ids {
		if track, err := u.trackRepo.GetByID(id); err == nil && track != nil {
			// Add to blacklist before deletion
			_ = u.blacklistRepo.Add(&domain.BlacklistItem{
				YouTubeID: track.YouTubeID,
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
	return u.trackRepo.BatchDelete(ids)
}

func (u *TrackUsecase) CleanBrokenTracks() error {
	return u.trackRepo.CleanBrokenTracks()
}

func (u *TrackUsecase) GetStats() (*domain.TrackStats, error) {
	return u.trackRepo.GetStats()
}
