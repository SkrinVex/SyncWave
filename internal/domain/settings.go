package domain

import "time"

type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SystemSettings struct {
	HTTPProxy               string `json:"http_proxy"`
	AudioFormat             string `json:"audio_format"` // opus, m4a, mp3
	AudioQuality            string `json:"audio_quality"`
	MaxConcurrent           int    `json:"max_concurrent"`
	AllowRegistration       bool   `json:"allow_registration"`
	GlobalStorageLimitBytes int64  `json:"global_storage_limit_bytes"` // 0 = unlimited
	DefaultUserQuotaBytes   int64  `json:"default_user_quota_bytes"`
	HasCookies              bool   `json:"has_cookies"`
	CookiesValid            bool   `json:"cookies_valid"`
	CookiesStatus           string `json:"cookies_status"` // valid, expiring_soon, expired, missing, invalid
	CookiesExpiresAt        string `json:"cookies_expires_at,omitempty"`
	CookiesError            string `json:"cookies_error,omitempty"`
	CookiesUpdatedAt        string `json:"cookies_updated_at,omitempty"`
	YTDLPVersion            string `json:"ytdlp_version"`
	FFmpegVersion           string `json:"ffmpeg_version"`
	StorageUsageBytes       int64  `json:"storage_usage_bytes"` // Total server music storage
	DatabaseSizeBytes       int64  `json:"database_size_bytes"`
	TotalTracksCount        int    `json:"total_tracks_count"`
	TotalPlaylistsCount     int    `json:"total_playlists_count"`
	UserStorageUsageBytes   int64  `json:"user_storage_usage_bytes"`
	UserStorageQuotaBytes   int64  `json:"user_storage_quota_bytes"`
	HostDiskTotalBytes      uint64 `json:"host_disk_total_bytes"`
	HostDiskUsedBytes       uint64 `json:"host_disk_used_bytes"`
	HostDiskFreeBytes       uint64 `json:"host_disk_free_bytes"`
	IsAdmin                 bool   `json:"is_admin"`
}

type SettingsRepository interface {
	Get(key string) (string, error)
	Set(key, value string) error
	GetAll() (map[string]string, error)
}
