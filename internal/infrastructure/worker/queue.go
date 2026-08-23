package worker

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/infrastructure/ytdlp"
)

type SyncTask struct {
	PlaylistID string
	Manual     bool
}

type WorkerQueue struct {
	ytdlpClient   *ytdlp.Client
	trackRepo     domain.TrackRepository
	playlistRepo  domain.PlaylistRepository
	logRepo       domain.SyncLogRepository
	blacklistRepo domain.BlacklistRepository
	userRepo      domain.UserRepository
	settingsRepo  domain.SettingsRepository
	eventHub      *EventHub

	taskQueue  chan SyncTask
	isSyncing  bool
	current    domain.SyncProgress
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	taskCtx    context.Context
	taskCancel context.CancelFunc
}

func NewWorkerQueue(
	ytdlpClient *ytdlp.Client,
	trackRepo domain.TrackRepository,
	playlistRepo domain.PlaylistRepository,
	logRepo domain.SyncLogRepository,
	blacklistRepo domain.BlacklistRepository,
	userRepo domain.UserRepository,
	settingsRepo domain.SettingsRepository,
	eventHub *EventHub,
	queueSize int,
) *WorkerQueue {
	ctx, cancel := context.WithCancel(context.Background())
	// On startup, clean up any unfinished/broken tracks from previous crashes
	_ = trackRepo.CleanBrokenTracks()

	return &WorkerQueue{
		ytdlpClient:   ytdlpClient,
		trackRepo:     trackRepo,
		playlistRepo:  playlistRepo,
		logRepo:       logRepo,
		blacklistRepo: blacklistRepo,
		userRepo:      userRepo,
		settingsRepo:  settingsRepo,
		eventHub:      eventHub,
		taskQueue:     make(chan SyncTask, queueSize),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (q *WorkerQueue) Start() {
	go q.workerLoop()
}

func (q *WorkerQueue) Stop() {
	q.CancelCurrent()
	q.cancel()
}

func (q *WorkerQueue) Enqueue(playlistID string, manual bool) error {
	task := SyncTask{
		PlaylistID: playlistID,
		Manual:     manual,
	}

	select {
	case q.taskQueue <- task:
		return nil
	default:
		return fmt.Errorf("sync queue is currently full")
	}
}

func (q *WorkerQueue) CancelCurrent() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.isSyncing && q.taskCancel != nil {
		q.taskCancel()
		q.isSyncing = false
		q.current = domain.SyncProgress{Active: false}
		q.eventHub.BroadcastProgress(domain.SyncProgress{Active: false})
		q.log(nil, nil, domain.LogLevelWarn, "Синхронизация была отменена пользователем")
	}
}

func (q *WorkerQueue) GetCurrentProgress() domain.SyncProgress {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.current
}

func (q *WorkerQueue) IsBusy() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.isSyncing
}

func (q *WorkerQueue) workerLoop() {
	for {
		select {
		case <-q.ctx.Done():
			return
		case task := <-q.taskQueue:
			q.processTask(task)
		}
	}
}

func (q *WorkerQueue) log(playlistID *string, trackID *string, level domain.LogLevel, message string) {
	logEntry := &domain.SyncLog{
		PlaylistID: playlistID,
		TrackID:    trackID,
		Level:      level,
		Message:    message,
		CreatedAt:  time.Now().UTC(),
	}
	_ = q.logRepo.Create(logEntry)
	q.eventHub.BroadcastLog(*logEntry)
}

func (q *WorkerQueue) processTask(task SyncTask) {
	q.mu.Lock()
	q.taskCtx, q.taskCancel = context.WithCancel(q.ctx)
	taskCtx := q.taskCtx
	q.isSyncing = true
	q.current = domain.SyncProgress{
		Active:     true,
		PlaylistID: task.PlaylistID,
		StatusText: "Initializing sync...",
	}
	q.mu.Unlock()

	defer func() {
		q.mu.Lock()
		q.isSyncing = false
		q.current = domain.SyncProgress{Active: false}
		q.mu.Unlock()
		q.eventHub.BroadcastProgress(domain.SyncProgress{Active: false})
	}()

	playlist, err := q.playlistRepo.GetByID(task.PlaylistID)
	if err != nil {
		q.log(nil, nil, domain.LogLevelError, fmt.Sprintf("Failed to find playlist %s: %v", task.PlaylistID, err))
		return
	}

	q.mu.Lock()
	q.current.PlaylistTitle = playlist.Title
	q.mu.Unlock()

	playlist.Status = domain.PlaylistStatusSyncing
	playlist.ErrorMessage = ""
	_ = q.playlistRepo.Update(playlist)
	q.eventHub.Broadcast(EventMessage{Type: EventTypePlaylist, Data: playlist})

	q.log(&playlist.ID, nil, domain.LogLevelInfo, fmt.Sprintf("Starting sync for playlist: %s (%s)", playlist.Title, playlist.YouTubeID))

	// Step 1: Extract playlist entries via flat-playlist
	q.mu.Lock()
	q.current.StatusText = "Extracting playlist tracks..."
	q.mu.Unlock()
	q.eventHub.BroadcastProgress(q.GetCurrentProgress())

	flatOutput, err := q.ytdlpClient.ExtractPlaylistDeltaForUser(taskCtx, playlist.YouTubeID, playlist.UserID)
	if err != nil {
		if taskCtx.Err() != nil {
			playlist.Status = domain.PlaylistStatusIdle
			_ = q.playlistRepo.Update(playlist)
			return
		}
		playlist.Status = domain.PlaylistStatusError
		playlist.ErrorMessage = err.Error()
		_ = q.playlistRepo.Update(playlist)
		q.log(&playlist.ID, nil, domain.LogLevelError, fmt.Sprintf("Playlist extraction failed: %v", err))
		q.eventHub.Broadcast(EventMessage{Type: EventTypePlaylist, Data: playlist})
		return
	}

	// Update playlist title if empty or default
	if (playlist.Title == "New Playlist" || playlist.Title == "") && flatOutput.Title != "" {
		playlist.Title = flatOutput.Title
	}

	extractedIDs := make([]string, 0, len(flatOutput.Entries))
	entryMap := make(map[string]ytdlp.PlaylistEntry)
	for _, e := range flatOutput.Entries {
		id := e.GetID()
		if id != "" {
			extractedIDs = append(extractedIDs, id)
			entryMap[id] = e
		}
	}

	q.log(&playlist.ID, nil, domain.LogLevelInfo, fmt.Sprintf("Found %d tracks in remote playlist", len(extractedIDs)))

	// Step 2: Batch delta check with database (User-scoped)
	existingMap, err := q.trackRepo.GetExistingYouTubeIDs(extractedIDs, playlist.UserID)
	if err != nil {
		q.log(&playlist.ID, nil, domain.LogLevelError, fmt.Sprintf("Failed to check existing tracks: %v", err))
		existingMap = make(map[string]bool)
	}

	// Identify missing tracks to download
	missingEntries := make([]ytdlp.PlaylistEntry, 0)
	blacklistedCount := 0

	for _, id := range extractedIDs {
		if !existingMap[id] {
			// Check if blacklisted for this user
			blacklisted, err := q.blacklistRepo.Exists(id, playlist.UserID)
			if err != nil {
				q.log(&playlist.ID, nil, domain.LogLevelError, fmt.Sprintf("Blacklist check error for %s: %v", id, err))
			}
			if blacklisted {
				blacklistedCount++
				continue
			}
			missingEntries = append(missingEntries, entryMap[id])
		}
	}

	q.log(&playlist.ID, nil, domain.LogLevelInfo, fmt.Sprintf("Delta identified: %d new tracks to download (skipped %d existing, %d blacklisted)", len(missingEntries), len(extractedIDs)-len(missingEntries)-blacklistedCount, blacklistedCount))

	// Step 3: Download missing tracks sequentially with rate-limit protection
	totalToDownload := len(missingEntries)
	successCount := 0
	failedCount := 0

	for idx, entry := range missingEntries {
		select {
		case <-taskCtx.Done():
			playlist.Status = domain.PlaylistStatusIdle
			_ = q.playlistRepo.Update(playlist)
			q.log(&playlist.ID, nil, domain.LogLevelWarn, "Синхронизация была прервана пользователем")
			return
		default:
		}

		// Check 1: User Storage Quota
		if user, err := q.userRepo.GetByID(playlist.UserID); err == nil && user != nil && user.StorageQuotaBytes > 0 {
			allStats, _ := q.userRepo.ListWithStats()
			for _, st := range allStats {
				if st.ID == user.ID && st.StorageUsedBytes >= user.StorageQuotaBytes {
					msg := fmt.Sprintf("Достигнут лимит дискового пространства пользователя (%d MB / %d MB). Синхронизация приостановлена.", st.StorageUsedBytes/(1024*1024), user.StorageQuotaBytes/(1024*1024))
					q.log(&playlist.ID, nil, domain.LogLevelWarn, msg)
					playlist.Status = domain.PlaylistStatusIdle
					_ = q.playlistRepo.Update(playlist)
					return
				}
			}
		}

		// Check 2: Global Server Music Limit
		if globalLimitStr, err := q.settingsRepo.Get("global_storage_limit_bytes"); err == nil && globalLimitStr != "" {
			if globalLimit, err := strconv.ParseInt(globalLimitStr, 10, 64); err == nil && globalLimit > 0 {
				if stats, _ := q.trackRepo.GetStats(""); stats != nil && stats.TotalStorageSize >= globalLimit {
					msg := fmt.Sprintf("Достигнут общий лимит хранилища музыки на сервере (%d MB / %d MB). Синхронизация остановлена.", stats.TotalStorageSize/(1024*1024), globalLimit/(1024*1024))
					q.log(&playlist.ID, nil, domain.LogLevelWarn, msg)
					playlist.Status = domain.PlaylistStatusIdle
					_ = q.playlistRepo.Update(playlist)
					return
				}
			}
		}

		currentIndex := idx + 1
		trackTitle := entry.Title
		if trackTitle == "" {
			trackTitle = entry.GetID()
		}

		q.mu.Lock()
		q.current.CurrentTrackIndex = currentIndex
		q.current.TotalTracks = totalToDownload
		q.current.CurrentTrackTitle = trackTitle
		q.current.CurrentTrackID = entry.GetID()
		q.current.TrackPercentage = 0
		q.current.Percentage = float64(idx) / float64(totalToDownload) * 100.0
		q.current.StatusText = fmt.Sprintf("[%d/%d] %s", currentIndex, totalToDownload, trackTitle)
		q.mu.Unlock()
		q.eventHub.BroadcastProgress(q.GetCurrentProgress())

		// Create temporary placeholder track record in DB
		trackID := uuid.New().String()
		now := time.Now().UTC()
		initialTrack := &domain.Track{
			ID:         trackID,
			YouTubeID:  entry.GetID(),
			PlaylistID: &playlist.ID,
			UserID:     playlist.UserID,
			Title:      trackTitle,
			Artist:     entry.Uploader,
			Duration:   entry.GetDuration(),
			Status:     domain.TrackStatusDownloading,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		_ = q.trackRepo.Create(initialTrack)

		q.log(&playlist.ID, &initialTrack.ID, domain.LogLevelInfo, fmt.Sprintf("[%d/%d] Fetching audio & tags: %s", currentIndex, totalToDownload, trackTitle))

		// Download track via yt-dlp using user's session
		res, dlErr := q.ytdlpClient.DownloadTrackForUser(taskCtx, entry.GetID(), playlist.UserID, func(percent float64, speed, eta, status string) {
			q.mu.Lock()
			q.current.TrackPercentage = percent
			q.current.Speed = speed
			q.current.ETA = eta
			if totalToDownload > 0 {
				q.current.Percentage = (float64(idx) + percent/100.0) / float64(totalToDownload) * 100.0
			}
			q.mu.Unlock()
			q.eventHub.BroadcastProgress(q.GetCurrentProgress())
		})

		if dlErr != nil {
			if taskCtx.Err() != nil {
				_ = q.trackRepo.Delete(initialTrack.ID, playlist.UserID)
				playlist.Status = domain.PlaylistStatusIdle
				_ = q.playlistRepo.Update(playlist)
				return
			}
			failedCount++
			// Remove the broken track from the database immediately so it never pollutes the library!
			_ = q.trackRepo.Delete(initialTrack.ID, playlist.UserID)
			q.log(&playlist.ID, nil, domain.LogLevelError, fmt.Sprintf("Failed to download %s: %v", trackTitle, dlErr))
		} else {
			successCount++
			initialTrack.Title = res.Title
			initialTrack.Artist = res.Artist
			initialTrack.Album = res.Album
			initialTrack.Duration = res.Duration
			initialTrack.FilePath = res.FilePath
			initialTrack.CoverPath = res.CoverPath
			initialTrack.FileSize = res.FileSize
			initialTrack.Format = res.Format
			initialTrack.Bitrate = res.Bitrate
			initialTrack.Status = domain.TrackStatusReady
			initialTrack.ErrorMessage = ""
			downloadedTime := time.Now().UTC()
			initialTrack.DownloadedAt = &downloadedTime
			_ = q.trackRepo.Update(initialTrack)

			q.log(&playlist.ID, &initialTrack.ID, domain.LogLevelSuccess, fmt.Sprintf("Successfully archived: %s - %s", res.Artist, res.Title))
			q.eventHub.Broadcast(EventMessage{Type: EventTypeTrack, Data: initialTrack})
		}

		// Jitter/sleep between downloads to avoid aggressive throttling
		time.Sleep(1200 * time.Millisecond)
	}

	// Finalize playlist status
	nowSync := time.Now().UTC()
	playlist.LastSyncedAt = &nowSync
	playlist.Status = domain.PlaylistStatusIdle
	playlist.ErrorMessage = ""
	_ = q.playlistRepo.Update(playlist)

	q.log(&playlist.ID, nil, domain.LogLevelSuccess, fmt.Sprintf("Sync completed for %s. Downloaded: %d, Failed: %d, Skipped: %d", playlist.Title, successCount, failedCount, len(extractedIDs)-totalToDownload))
	q.eventHub.Broadcast(EventMessage{Type: EventTypePlaylist, Data: playlist})
}
