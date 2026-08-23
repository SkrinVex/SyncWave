package sqlite

import (
	"fmt"

	"github.com/skrinvex/SyncWave/internal/domain"
)

type BlacklistRepo struct {
	db *DB
}

func NewBlacklistRepo(db *DB) domain.BlacklistRepository {
	return &BlacklistRepo{db: db}
}

func (r *BlacklistRepo) Add(item *domain.BlacklistItem) error {
	query := `
		INSERT INTO blacklist (youtube_id, title, artist, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(youtube_id) DO UPDATE SET
			title = excluded.title,
			artist = excluded.artist
	`
	_, err := r.db.Exec(query, item.YouTubeID, item.Title, item.Artist)
	if err != nil {
		return fmt.Errorf("failed to insert blacklist item: %w", err)
	}
	return nil
}

func (r *BlacklistRepo) Remove(youtubeID string) error {
	query := `DELETE FROM blacklist WHERE youtube_id = ?`
	_, err := r.db.Exec(query, youtubeID)
	if err != nil {
		return fmt.Errorf("failed to remove blacklist item: %w", err)
	}
	return nil
}

func (r *BlacklistRepo) Exists(youtubeID string) (bool, error) {
	var count int
	query := `SELECT COUNT(1) FROM blacklist WHERE youtube_id = ?`
	err := r.db.QueryRow(query, youtubeID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist existence: %w", err)
	}
	return count > 0, nil
}

func (r *BlacklistRepo) List(searchQuery string) ([]domain.BlacklistItem, error) {
	query := `SELECT youtube_id, title, artist, created_at FROM blacklist`
	var args []interface{}
	
	if searchQuery != "" {
		query += ` WHERE title LIKE ? OR artist LIKE ?`
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
		if err := rows.Scan(&item.YouTubeID, &item.Title, &item.Artist, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan blacklist item: %w", err)
		}
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
