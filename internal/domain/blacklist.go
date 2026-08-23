package domain

import "time"

type BlacklistItem struct {
	YouTubeID string    `json:"youtube_id"`
	Title     string    `json:"title"`
	Artist    string    `json:"artist"`
	CreatedAt time.Time `json:"created_at"`
}

type BlacklistRepository interface {
	Add(item *BlacklistItem) error
	Remove(youtubeID string) error
	Exists(youtubeID string) (bool, error)
	List(query string) ([]BlacklistItem, error)
}
