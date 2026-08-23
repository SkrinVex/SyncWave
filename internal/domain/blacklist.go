package domain

import "time"

type BlacklistItem struct {
	YouTubeID string    `json:"youtube_id"`
	UserID    string    `json:"user_id,omitempty"`
	Title     string    `json:"title"`
	Artist    string    `json:"artist"`
	CreatedAt time.Time `json:"created_at"`
}

type BlacklistRepository interface {
	Add(item *BlacklistItem) error
	Remove(youtubeID string, userID string) error
	Exists(youtubeID string, userID string) (bool, error)
	List(userID string, query string) ([]BlacklistItem, error)
}
