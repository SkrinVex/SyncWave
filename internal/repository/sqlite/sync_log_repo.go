package sqlite

import (
	"database/sql"
	"time"

	"github.com/syncwave/syncwave/internal/domain"
)

type SyncLogRepository struct {
	db *DB
}

func NewSyncLogRepository(db *DB) *SyncLogRepository {
	return &SyncLogRepository{db: db}
}

func (r *SyncLogRepository) Create(log *domain.SyncLog) error {
	log.CreatedAt = time.Now().UTC()
	query := `
	INSERT INTO sync_logs (user_id, playlist_id, track_id, level, message, created_at)
	VALUES (?, ?, ?, ?, ?, ?);
	`
	res, err := r.db.Exec(query, log.UserID, log.PlaylistID, log.TrackID, string(log.Level), log.Message, log.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		log.ID = id
	}
	return nil
}

func (r *SyncLogRepository) ListRecent(limit int, userID string) ([]domain.SyncLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	var query string
	var args []interface{}

	if userID != "" {
		query = `
		SELECT id, user_id, playlist_id, track_id, level, message, created_at
		FROM sync_logs
		WHERE user_id = ? OR user_id = ''
		ORDER BY id DESC
		LIMIT ?;
		`
		args = []interface{}{userID, limit}
	} else {
		query = `
		SELECT id, user_id, playlist_id, track_id, level, message, created_at
		FROM sync_logs
		ORDER BY id DESC
		LIMIT ?;
		`
		args = []interface{}{limit}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]domain.SyncLog, 0)
	for rows.Next() {
		var l domain.SyncLog
		var uid string
		var playlistID, trackID sql.NullString
		var levelStr string

		err := rows.Scan(&l.ID, &uid, &playlistID, &trackID, &levelStr, &l.Message, &l.CreatedAt)
		if err != nil {
			return nil, err
		}

		l.UserID = uid
		l.Level = domain.LogLevel(levelStr)
		if playlistID.Valid {
			l.PlaylistID = &playlistID.String
		}
		if trackID.Valid {
			l.TrackID = &trackID.String
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *SyncLogRepository) ClearOlderThan(days int, userID string) error {
	if days <= 0 {
		return r.ClearAll(userID)
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	if userID != "" {
		query := `DELETE FROM sync_logs WHERE created_at < ? AND (user_id = ? OR user_id = '');`
		_, err := r.db.Exec(query, cutoff, userID)
		return err
	}
	query := `DELETE FROM sync_logs WHERE created_at < ?;`
	_, err := r.db.Exec(query, cutoff)
	return err
}

func (r *SyncLogRepository) ClearAll(userID string) error {
	if userID != "" {
		query := `DELETE FROM sync_logs WHERE user_id = ? OR user_id = '';`
		_, err := r.db.Exec(query, userID)
		return err
	}
	query := `DELETE FROM sync_logs;`
	_, err := r.db.Exec(query)
	return err
}
