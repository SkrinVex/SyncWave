package domain

import "time"

type TrackStatus string

const (
	TrackStatusQueued      TrackStatus = "queued"
	TrackStatusDownloading TrackStatus = "downloading"
	TrackStatusReady       TrackStatus = "ready"
	TrackStatusFailed      TrackStatus = "failed"
)

type Track struct {
	ID           string      `json:"id"`
	YouTubeID    string      `json:"youtube_id"`
	PlaylistID   *string     `json:"playlist_id,omitempty"`
	Title        string      `json:"title"`
	Artist       string      `json:"artist"`
	Album        string      `json:"album"`
	Duration     int         `json:"duration"` // in seconds
	FilePath     string      `json:"file_path,omitempty"`
	CoverPath    string      `json:"cover_path,omitempty"`
	FileSize     int64       `json:"file_size"` // in bytes
	Format       string      `json:"format"`    // opus, m4a, mp3
	Bitrate      int         `json:"bitrate"`   // in kbps
	Status       TrackStatus `json:"status"`
	ErrorMessage string      `json:"error_message,omitempty"`
	DownloadedAt *time.Time  `json:"downloaded_at,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type TrackFilter struct {
	Query      string // search title, artist, album
	PlaylistID string
	Status     TrackStatus
	SortBy     string // "created_at", "title", "artist", "duration"
	Order      string // "asc", "desc"
	Page       int
	PageSize   int
}

type TrackListResult struct {
	Tracks     []Track `json:"tracks"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}

type TrackStats struct {
	TotalTracks      int   `json:"total_tracks"`
	ReadyTracks      int   `json:"ready_tracks"`
	FailedTracks     int   `json:"failed_tracks"`
	TotalStorageSize int64 `json:"total_storage_size"` // bytes
	TotalDuration    int64 `json:"total_duration"`     // seconds
}

type TrackRepository interface {
	Create(track *Track) error
	GetByID(id string) (*Track, error)
	GetByYouTubeID(youtubeID string) (*Track, error)
	GetExistingYouTubeIDs(youtubeIDs []string) (map[string]bool, error)
	List(filter TrackFilter) (*TrackListResult, error)
	Update(track *Track) error
	Delete(id string) error
	GetStats() (*TrackStats, error)
	GetAllReady() ([]Track, error)
}
