package usecase

import (
	"os"

	"github.com/syncwave/syncwave/internal/domain"
)

type TrackUsecase struct {
	trackRepo domain.TrackRepository
}

func NewTrackUsecase(trackRepo domain.TrackRepository) *TrackUsecase {
	return &TrackUsecase{
		trackRepo: trackRepo,
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
