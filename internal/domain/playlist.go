package domain

import "time"

type PlaylistStatus string

const (
	PlaylistStatusIdle    PlaylistStatus = "idle"
	PlaylistStatusSyncing PlaylistStatus = "syncing"
	PlaylistStatusError   PlaylistStatus = "error"
)

type Playlist struct {
	ID                  string         `json:"id"`
	UserID              string         `json:"user_id"`
	Title               string         `json:"title"`
	YouTubeID           string         `json:"youtube_id"` // e.g. "LM" or playlist ID/URL
	AutoSync            bool           `json:"auto_sync"`
	SyncIntervalMinutes int            `json:"sync_interval_minutes"`
	LastSyncedAt        *time.Time     `json:"last_synced_at,omitempty"`
	Status              PlaylistStatus `json:"status"`
	TrackCount          int            `json:"track_count"`
	ErrorMessage        string         `json:"error_message,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

type PlaylistRepository interface {
	Create(playlist *Playlist) error
	GetByID(id string) (*Playlist, error)
	GetByYouTubeID(youtubeID string) (*Playlist, error)
	ListByUserID(userID string) ([]Playlist, error)
	ListAutoSync() ([]Playlist, error)
	Update(playlist *Playlist) error
	Delete(id string) error
}
