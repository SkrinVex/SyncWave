package worker

import (
	"context"
	"time"

	"github.com/syncwave/syncwave/internal/domain"
)

type Scheduler struct {
	playlistRepo domain.PlaylistRepository
	queue        *WorkerQueue
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewScheduler(playlistRepo domain.PlaylistRepository, queue *WorkerQueue) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		playlistRepo: playlistRepo,
		queue:        queue,
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (s *Scheduler) Start() {
	go s.cronLoop()
}

func (s *Scheduler) Stop() {
	s.cancel()
}

func (s *Scheduler) cronLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkAndTriggerSyncs()
		}
	}
}

func (s *Scheduler) checkAndTriggerSyncs() {
	if s.queue.IsBusy() {
		return
	}

	playlists, err := s.playlistRepo.ListAutoSync()
	if err != nil {
		return
	}

	now := time.Now().UTC()
	for _, p := range playlists {
		if !p.AutoSync {
			continue
		}

		intervalMinutes := p.SyncIntervalMinutes
		if intervalMinutes <= 0 {
			intervalMinutes = 60
		}

		shouldSync := false
		if p.LastSyncedAt == nil {
			shouldSync = true
		} else {
			diff := now.Sub(*p.LastSyncedAt)
			if diff >= time.Duration(intervalMinutes)*time.Minute {
				shouldSync = true
			}
		}

		if shouldSync {
			_ = s.queue.Enqueue(p.ID, false)
			// Enqueue one at a time to prevent backlog spike
			break
		}
	}
}
