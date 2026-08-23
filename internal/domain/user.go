package domain

import "time"

type User struct {
	ID                string    `json:"id"`
	Username          string    `json:"username"`
	PasswordHash      string    `json:"-"`
	IsAdmin           bool      `json:"is_admin"`
	StorageQuotaBytes int64     `json:"storage_quota_bytes"` // 0 = unlimited, default 10GB
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type UserWithStats struct {
	User
	StorageUsedBytes int64 `json:"storage_used_bytes"`
	TracksCount      int   `json:"tracks_count"`
	PlaylistsCount   int   `json:"playlists_count"`
}

type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByUsername(username string) (*User, error)
	Count() (int, error)
	ListWithStats() ([]UserWithStats, error)
	Update(user *User) error
	UpdateQuota(userID string, quota int64) error
	Delete(id string) error
}
