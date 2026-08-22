package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/infrastructure/ytdlp"
)

type SyncTask struct {
	PlaylistID string
	Manual     bool
	ResultChan chan error
}

type WorkerQueue struct {
	ytdlpClient  *ytdlp.Client
	trackRepo    domain.TrackRepository
	playlistRepo domain.PlaylistRepository
	logRepo      domain.SyncLogRepository
	eventHub     *EventHub

	taskQueue chan SyncTask
	mu        sync.RWMutex
	isSyncing bool
	current   domain.SyncProgress
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewWorkerQueue(
	ytdlpClient *ytdlp.Client,
	trackRepo domain.TrackRepository,
	playlistRepo domain.PlaylistRepository,
	logRepo domain.SyncLogRepository,
	eventHub *EventHub,
	maxQueueSize int,
) *WorkerQueue {
	if maxQueueSize <= 0 {
		maxQueueSize = 50
	}
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerQueue{
		ytdlpClient:  ytdlpClient,
		trackRepo:    trackRepo,
		playlistRepo: playlistRepo,
		logRepo:      logRepo,
		eventHub:     eventHub,
		taskQueue:    make(chan SyncTask, maxQueueSize),
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (q *WorkerQueue) Start() {
	go q.workerLoop()
}

func (q *WorkerQueue) Stop() {
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

	flatOutput, err := q.ytdlpClient.ExtractPlaylistDelta(q.ctx, playlist.YouTubeID)
	if err != nil {
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
		if e.ID != "" {
			extractedIDs = append(extractedIDs, e.ID)
			entryMap[e.ID] = e
		}
	}

	q.log(&playlist.ID, nil, domain.LogLevelInfo, fmt.Sprintf("Found %d tracks in remote playlist", len(extractedIDs)))

	// Step 2: Batch delta check with database
	existingMap, err := q.trackRepo.GetExistingYouTubeIDs(extractedIDs)
	if err != nil {
		q.log(&playlist.ID, nil, domain.LogLevelError, fmt.Sprintf("Failed to check existing tracks: %v", err))
		existingMap = make(map[string]bool)
	}

	// Identify missing tracks to download
	missingEntries := make([]ytdlp.PlaylistEntry, 0)
	for _, id := range extractedIDs {
		if !existingMap[id] {
			missingEntries = append(missingEntries, entryMap[id])
		}
	}

	q.log(&playlist.ID, nil, domain.LogLevelInfo, fmt.Sprintf("Delta identified: %d new tracks to download (skipped %d existing)", len(missingEntries), len(extractedIDs)-len(missingEntries)))

	// Step 3: Download missing tracks sequentially with rate-limit protection
	totalToDownload := len(missingEntries)
	successCount := 0
	failedCount := 0

	for idx, entry := range missingEntries {
		select {
		case <-q.ctx.Done():
			q.log(&playlist.ID, nil, domain.LogLevelWarn, "Sync aborted: server shutting down")
			return
		default:
		}

		currentIndex := idx + 1
		trackTitle := entry.Title
		if trackTitle == "" {
			trackTitle = entry.ID
		}

		q.mu.Lock()
		q.current.CurrentTrackIndex = currentIndex
		q.current.TotalTracks = totalToDownload
		q.current.CurrentTrackTitle = trackTitle
		q.current.CurrentTrackID = entry.ID
		q.current.Percentage = float64(idx) / float64(totalToDownload) * 100.0
		q.current.StatusText = fmt.Sprintf("[%d/%d] Downloading: %s", currentIndex, totalToDownload, trackTitle)
		q.mu.Unlock()
		q.eventHub.BroadcastProgress(q.GetCurrentProgress())

		// Create/ensure placeholder track record in DB
		trackID := uuid.New().String()
		now := time.Now().UTC()
		initialTrack := &domain.Track{
			ID:         trackID,
			YouTubeID:  entry.ID,
			PlaylistID: &playlist.ID,
			Title:      trackTitle,
			Artist:     entry.Uploader,
			Duration:   entry.Duration,
			Status:     domain.TrackStatusDownloading,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		_ = q.trackRepo.Create(initialTrack)

		q.log(&playlist.ID, &initialTrack.ID, domain.LogLevelInfo, fmt.Sprintf("[%d/%d] Fetching audio & tags: %s", currentIndex, totalToDownload, trackTitle))

		// Download track via yt-dlp
		res, dlErr := q.ytdlpClient.DownloadTrack(q.ctx, entry.ID, func(percent float64, speed, eta, status string) {
			q.mu.Lock()
			q.current.Speed = speed
			q.current.ETA = eta
			q.mu.Unlock()
			q.eventHub.BroadcastProgress(q.GetCurrentProgress())
		})

		if dlErr != nil {
			failedCount++
			initialTrack.Status = domain.TrackStatusFailed
			initialTrack.ErrorMessage = dlErr.Error()
			_ = q.trackRepo.Update(initialTrack)
			q.log(&playlist.ID, &initialTrack.ID, domain.LogLevelError, fmt.Sprintf("Failed to download %s: %v", trackTitle, dlErr))
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
		time.Sleep(1500 * time.Millisecond)
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
