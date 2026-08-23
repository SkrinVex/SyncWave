package domain

import "time"

type LogLevel string

const (
	LogLevelInfo    LogLevel = "info"
	LogLevelSuccess LogLevel = "success"
	LogLevelWarn    LogLevel = "warn"
	LogLevelError   LogLevel = "error"
)

type SyncLog struct {
	ID         int64     `json:"id"`
	UserID     string    `json:"user_id,omitempty"`
	PlaylistID *string   `json:"playlist_id,omitempty"`
	TrackID    *string   `json:"track_id,omitempty"`
	Level      LogLevel  `json:"level"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

type SyncProgress struct {
	Active            bool    `json:"active"`
	UserID            string  `json:"user_id,omitempty"`
	PlaylistID        string  `json:"playlist_id"`
	PlaylistTitle     string  `json:"playlist_title"`
	CurrentTrackIndex int     `json:"current_track_index"`
	TotalTracks       int     `json:"total_tracks"`
	CurrentTrackTitle string  `json:"current_track_title"`
	CurrentTrackID    string  `json:"current_track_id"`
	TrackPercentage   float64 `json:"track_percentage"`
	Percentage        float64 `json:"percentage"`
	Speed             string  `json:"speed"`
	ETA               string  `json:"eta"`
	StatusText        string  `json:"status_text"`
}

type SyncLogRepository interface {
	Create(log *SyncLog) error
	ListRecent(limit int, userID string) ([]SyncLog, error)
	ClearOlderThan(days int, userID string) error
	ClearAll(userID string) error
}
