package sqlite

import (
	"fmt"

	"github.com/syncwave/syncwave/internal/domain"
)

type BlacklistRepo struct {
	db *DB
}

func NewBlacklistRepo(db *DB) domain.BlacklistRepository {
	return &BlacklistRepo{db: db}
}

func (r *BlacklistRepo) Add(item *domain.BlacklistItem) error {
	query := `
		INSERT INTO blacklist (youtube_id, user_id, title, artist, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(youtube_id) DO UPDATE SET
			user_id = excluded.user_id,
			title = excluded.title,
			artist = excluded.artist;
	`
	_, err := r.db.Exec(query, item.YouTubeID, item.UserID, item.Title, item.Artist)
	if err != nil {
		return fmt.Errorf("failed to insert blacklist item: %w", err)
	}
	return nil
}

func (r *BlacklistRepo) Remove(youtubeID string, userID string) error {
	query := `DELETE FROM blacklist WHERE youtube_id = ? AND (user_id = ? OR user_id IS NULL OR user_id = '')`
	_, err := r.db.Exec(query, youtubeID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove blacklist item: %w", err)
	}
	return nil
}

func (r *BlacklistRepo) Exists(youtubeID string, userID string) (bool, error) {
	var count int
	query := `SELECT COUNT(1) FROM blacklist WHERE youtube_id = ? AND (user_id = ? OR user_id IS NULL OR user_id = '')`
	err := r.db.QueryRow(query, youtubeID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist existence: %w", err)
	}
	return count > 0, nil
}

func (r *BlacklistRepo) List(userID string, searchQuery string) ([]domain.BlacklistItem, error) {
	query := `SELECT youtube_id, user_id, title, artist, created_at FROM blacklist WHERE (user_id = ? OR user_id IS NULL OR user_id = '')`
	args := []interface{}{userID}

	if searchQuery != "" {
		query += ` AND (title LIKE ? OR artist LIKE ?)`
		searchPattern := "%" + searchQuery + "%"
		args = append(args, searchPattern, searchPattern)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list blacklist items: %w", err)
	}
	defer rows.Close()

	var items []domain.BlacklistItem
	for rows.Next() {
		var item domain.BlacklistItem
		var uid sqlNullStringOrRegular
		if err := rows.Scan(&item.YouTubeID, &uid.val, &item.Title, &item.Artist, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan blacklist item: %w", err)
		}
		item.UserID = uid.val
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error during blacklist listing: %w", err)
	}

	if items == nil {
		items = []domain.BlacklistItem{}
	}
	return items, nil
}

type sqlNullStringOrRegular struct {
	val string
}

func (s *sqlNullStringOrRegular) Scan(value interface{}) error {
	if value == nil {
		s.val = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		s.val = v
	case []byte:
		s.val = string(v)
	}
	return nil
}
