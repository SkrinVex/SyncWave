package domain

import "time"

type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SystemSettings struct {
	HTTPProxy           string `json:"http_proxy"`
	AudioFormat         string `json:"audio_format"` // opus, m4a, mp3
	AudioQuality        string `json:"audio_quality"`
	MaxConcurrent       int    `json:"max_concurrent"`
	HasCookies          bool   `json:"has_cookies"`
	CookiesValid        bool   `json:"cookies_valid"`
	CookiesUpdatedAt    string `json:"cookies_updated_at,omitempty"`
	YTDLPVersion        string `json:"ytdlp_version"`
	FFmpegVersion       string `json:"ffmpeg_version"`
	StorageUsageBytes   int64  `json:"storage_usage_bytes"`
	StorageFreeBytes    int64  `json:"storage_free_bytes"`
	DatabaseSizeBytes   int64  `json:"database_size_bytes"`
	TotalTracksCount    int    `json:"total_tracks_count"`
	TotalPlaylistsCount int    `json:"total_playlists_count"`
}

type SettingsRepository interface {
	Get(key string) (string, error)
	Set(key, value string) error
	GetAll() (map[string]string, error)
}
