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
	INSERT INTO sync_logs (playlist_id, track_id, level, message, created_at)
	VALUES (?, ?, ?, ?, ?);
	`
	res, err := r.db.Exec(query, log.PlaylistID, log.TrackID, string(log.Level), log.Message, log.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		log.ID = id
	}
	return nil
}

func (r *SyncLogRepository) ListRecent(limit int) ([]domain.SyncLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	query := `
	SELECT id, playlist_id, track_id, level, message, created_at
	FROM sync_logs
	ORDER BY id DESC
	LIMIT ?;
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]domain.SyncLog, 0)
	for rows.Next() {
		var l domain.SyncLog
		var playlistID, trackID sql.NullString
		var levelStr string

		err := rows.Scan(&l.ID, &playlistID, &trackID, &levelStr, &l.Message, &l.CreatedAt)
		if err != nil {
			return nil, err
		}

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

func (r *SyncLogRepository) ClearOlderThan(days int) error {
	if days <= 0 {
		return r.ClearAll()
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	query := `DELETE FROM sync_logs WHERE created_at < ?;`
	_, err := r.db.Exec(query, cutoff)
	return err
}

func (r *SyncLogRepository) ClearAll() error {
	query := `DELETE FROM sync_logs;`
	_, err := r.db.Exec(query)
	return err
}
