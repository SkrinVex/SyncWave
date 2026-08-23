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

func (u *SyncUsecase) CancelSync(userID string) {
	u.workerQueue.CancelCurrent(userID)
}

func (u *SyncUsecase) GetProgress(userID string) domain.SyncProgress {
	return u.workerQueue.GetCurrentProgress(userID)
}

func (u *SyncUsecase) GetLogs(limit int, userID string) ([]domain.SyncLog, error) {
	return u.logRepo.ListRecent(limit, userID)
}

func (u *SyncUsecase) ClearLogs(days int, userID string) error {
	if days > 0 {
		return u.logRepo.ClearOlderThan(days, userID)
	}
	return u.logRepo.ClearAll(userID)
}
