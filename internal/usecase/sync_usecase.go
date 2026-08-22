package usecase

import (
	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/infrastructure/worker"
)

type SyncUsecase struct {
	playlistRepo domain.PlaylistRepository
	logRepo      domain.SyncLogRepository
	workerQueue  *worker.WorkerQueue
}

func NewSyncUsecase(
	playlistRepo domain.PlaylistRepository,
	logRepo domain.SyncLogRepository,
	workerQueue *worker.WorkerQueue,
) *SyncUsecase {
	return &SyncUsecase{
		playlistRepo: playlistRepo,
		logRepo:      logRepo,
		workerQueue:  workerQueue,
	}
}

func (u *SyncUsecase) TriggerSyncAll(userID string) error {
	playlists, err := u.playlistRepo.ListByUserID(userID)
	if err != nil {
		return err
	}

	for _, p := range playlists {
		_ = u.workerQueue.Enqueue(p.ID, true)
	}
	return nil
}

func (u *SyncUsecase) TriggerSyncPlaylist(playlistID string) error {
	return u.workerQueue.Enqueue(playlistID, true)
}

func (u *SyncUsecase) CancelSync() {
	u.workerQueue.CancelCurrent()
}

func (u *SyncUsecase) GetProgress() domain.SyncProgress {
	return u.workerQueue.GetCurrentProgress()
}

func (u *SyncUsecase) GetLogs(limit int) ([]domain.SyncLog, error) {
	return u.logRepo.ListRecent(limit)
}

func (u *SyncUsecase) ClearLogs(days int) error {
	return u.logRepo.ClearAll()
}
