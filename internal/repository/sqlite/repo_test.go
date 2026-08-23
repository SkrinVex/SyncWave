package sqlite_test

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/repository/sqlite"
)

func setupTestDB(t *testing.T) *sqlite.DB {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	return db
}

func TestUserRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	repo := sqlite.NewUserRepository(db)

	count, err := repo.Count()
	if err != nil || count != 0 {
		t.Fatalf("expected count 0, got %d, err: %v", count, err)
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Username:     "admin",
		PasswordHash: "hashed_secret",
		IsAdmin:      true,
	}

	if err := repo.Create(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	fetched, err := repo.GetByUsername("admin")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if fetched.ID != user.ID || fetched.Username != "admin" || !fetched.IsAdmin {
		t.Fatalf("fetched user mismatch: %+v", fetched)
	}
}

func TestTrackAndPlaylistRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := sqlite.NewUserRepository(db)
	playlistRepo := sqlite.NewPlaylistRepository(db)
	trackRepo := sqlite.NewTrackRepository(db)

	// Create user
	userID := uuid.New().String()
	_ = userRepo.Create(&domain.User{
		ID:           userID,
		Username:     "tester",
		PasswordHash: "hash",
	})

	// Create playlist
	playlistID := uuid.New().String()
	pl := &domain.Playlist{
		ID:                  playlistID,
		UserID:              userID,
		Title:               "Liked Songs",
		YouTubeID:           "LM",
		AutoSync:            true,
		SyncIntervalMinutes: 30,
		Status:              domain.PlaylistStatusIdle,
	}
	if err := playlistRepo.Create(pl); err != nil {
		t.Fatalf("failed to create playlist: %v", err)
	}

	// Create tracks
	track1 := &domain.Track{
		ID:         uuid.New().String(),
		UserID:     userID,
		YouTubeID:  "yt-101",
		PlaylistID: &playlistID,
		Title:      "Solaris",
		Artist:     "Photay",
		Album:      "Waking Season",
		Duration:   215,
		FilePath:   "/tmp/yt-101.opus",
		Format:     "opus",
		Status:     domain.TrackStatusReady,
	}
	if err := trackRepo.Create(track1); err != nil {
		t.Fatalf("failed to create track1: %v", err)
	}

	track2 := &domain.Track{
		ID:         uuid.New().String(),
		UserID:     userID,
		YouTubeID:  "yt-102",
		PlaylistID: &playlistID,
		Title:      "Aura",
		Artist:     "Bicep",
		Album:      "Bicep",
		Duration:   310,
		FilePath:   "/tmp/yt-102.opus",
		Format:     "opus",
		Status:     domain.TrackStatusReady,
	}
	if err := trackRepo.Create(track2); err != nil {
		t.Fatalf("failed to create track2: %v", err)
	}

	// Delta lookup test
	existing, err := trackRepo.GetExistingYouTubeIDs([]string{"yt-101", "yt-102", "yt-999"}, "")
	if err != nil {
		t.Fatalf("delta query failed: %v", err)
	}
	if !existing["yt-101"] || !existing["yt-102"] || existing["yt-999"] {
		t.Fatalf("delta map incorrect: %+v", existing)
	}

	// List filter test
	listRes, err := trackRepo.List(domain.TrackFilter{
		Query: "Photay",
		Page:  1,
	})
	if err != nil || listRes.Total != 1 {
		t.Fatalf("search filter failed, expected 1 track, got %d, err: %v", listRes.Total, err)
	}

	// GetAllReady tests
	allReady, err := trackRepo.GetAllReady(userID, "")
	if err != nil || len(allReady) != 2 {
		t.Fatalf("expected 2 ready tracks, got %d, err: %v", len(allReady), err)
	}

	plReady, err := trackRepo.GetAllReady(userID, playlistID)
	if err != nil || len(plReady) != 2 {
		t.Fatalf("expected 2 ready tracks for playlist, got %d, err: %v", len(plReady), err)
	}

	// File count test
	count, err := trackRepo.CountTracksByFilePath("/tmp/yt-101.opus")
	if err != nil || count != 1 {
		t.Fatalf("expected count 1 for file, got %d, err: %v", count, err)
	}

	// Stats test
	stats, err := trackRepo.GetStats("")
	if err != nil {
		t.Fatalf("stats failed: %v", err)
	}
	if stats.TotalTracks != 2 || stats.ReadyTracks != 2 {
		t.Fatalf("stats mismatch: %+v", stats)
	}
}

func TestSettingsAndLogs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	settingsRepo := sqlite.NewSettingsRepository(db)
	logRepo := sqlite.NewSyncLogRepository(db)

	_ = settingsRepo.Set("audio_format", "flac")
	val, err := settingsRepo.Get("audio_format")
	if err != nil || val != "flac" {
		t.Fatalf("expected flac, got %s, err: %v", val, err)
	}

	userID := uuid.New().String()
	logEntry := &domain.SyncLog{
		UserID:    userID,
		Level:     domain.LogLevelInfo,
		Message:   "Sync test initiated",
		CreatedAt: time.Now().UTC(),
	}
	if err := logRepo.Create(logEntry); err != nil {
		t.Fatalf("failed to insert log: %v", err)
	}

	logs, err := logRepo.ListRecent(10, userID)
	if err != nil || len(logs) == 0 {
		t.Fatalf("expected logs, got %v, err: %v", logs, err)
	}

	// Test user isolation for logs
	otherLogs, err := logRepo.ListRecent(10, "other-user")
	if err != nil || len(otherLogs) != 0 {
		t.Fatalf("expected 0 logs for other user, got %d", len(otherLogs))
	}
}

func TestLegacyMigration(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "legacy.db")

	// 1. Manually create old legacy tables WITHOUT user_id and storage_quota_bytes
	legacySchema := `
	CREATE TABLE users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		is_admin INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE playlists (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		youtube_id TEXT NOT NULL,
		auto_sync INTEGER NOT NULL DEFAULT 1,
		sync_interval_minutes INTEGER NOT NULL DEFAULT 60,
		last_synced_at DATETIME,
		status TEXT NOT NULL DEFAULT 'idle',
		error_message TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE tracks (
		id TEXT PRIMARY KEY,
		youtube_id TEXT NOT NULL,
		playlist_id TEXT,
		title TEXT NOT NULL,
		artist TEXT NOT NULL,
		album TEXT NOT NULL DEFAULT '',
		duration INTEGER NOT NULL DEFAULT 0,
		file_path TEXT NOT NULL DEFAULT '',
		cover_path TEXT NOT NULL DEFAULT '',
		file_size INTEGER NOT NULL DEFAULT 0,
		format TEXT NOT NULL DEFAULT 'opus',
		bitrate INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'ready',
		error_message TEXT NOT NULL DEFAULT '',
		downloaded_at DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE sync_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		playlist_id TEXT,
		track_id TEXT,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE blacklist (
		youtube_id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		artist TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open raw sqlite: %v", err)
	}
	if _, err := db.Exec(legacySchema); err != nil {
		t.Fatalf("failed to create legacy schema: %v", err)
	}
	db.Close()

	// 2. Now open with NewDB (which runs Migrate) and verify no errors
	upgradedDB, err := sqlite.NewDB(dbPath)
	if err != nil {
		t.Fatalf("failed to upgrade legacy db: %v", err)
	}
	defer upgradedDB.Close()

	// 3. Verify user_id column now exists and works
	trackRepo := sqlite.NewTrackRepository(upgradedDB)
	tracks, err := trackRepo.GetAllReady("some-user", "")
	if err != nil {
		t.Fatalf("failed to query upgraded tracks: %v", err)
	}
	if len(tracks) != 0 {
		t.Fatalf("expected 0 tracks, got %d", len(tracks))
	}
}
