package sqlite_test

import (
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
	existing, err := trackRepo.GetExistingYouTubeIDs([]string{"yt-101", "yt-102", "yt-999"})
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

	// Stats test
	stats, err := trackRepo.GetStats()
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

	logEntry := &domain.SyncLog{
		Level:     domain.LogLevelInfo,
		Message:   "Sync test initiated",
		CreatedAt: time.Now().UTC(),
	}
	if err := logRepo.Create(logEntry); err != nil {
		t.Fatalf("failed to insert log: %v", err)
	}

	logs, err := logRepo.ListRecent(10)
	if err != nil || len(logs) == 0 {
		t.Fatalf("expected logs, got %v, err: %v", logs, err)
	}
}
