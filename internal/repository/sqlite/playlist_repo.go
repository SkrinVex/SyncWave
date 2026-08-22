package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/syncwave/syncwave/internal/domain"
)

type PlaylistRepository struct {
	db *DB
}

func NewPlaylistRepository(db *DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

func (r *PlaylistRepository) Create(p *domain.Playlist) error {
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	query := `
	INSERT INTO playlists (
		id, user_id, title, youtube_id, auto_sync, sync_interval_minutes,
		last_synced_at, status, error_message, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	autoSyncInt := 0
	if p.AutoSync {
		autoSyncInt = 1
	}

	_, err := r.db.Exec(
		query,
		p.ID, p.UserID, p.Title, p.YouTubeID, autoSyncInt, p.SyncIntervalMinutes,
		p.LastSyncedAt, string(p.Status), p.ErrorMessage, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PlaylistRepository) GetByID(id string) (*domain.Playlist, error) {
	query := `
	SELECT p.id, p.user_id, p.title, p.youtube_id, p.auto_sync, p.sync_interval_minutes,
	       p.last_synced_at, p.status, p.error_message, p.created_at, p.updated_at,
	       (SELECT COUNT(*) FROM tracks t WHERE t.playlist_id = p.id) AS track_count
	FROM playlists p
	WHERE p.id = ?;
	`
	return r.scanPlaylist(r.db.QueryRow(query, id))
}

func (r *PlaylistRepository) GetByYouTubeID(youtubeID string) (*domain.Playlist, error) {
	query := `
	SELECT p.id, p.user_id, p.title, p.youtube_id, p.auto_sync, p.sync_interval_minutes,
	       p.last_synced_at, p.status, p.error_message, p.created_at, p.updated_at,
	       (SELECT COUNT(*) FROM tracks t WHERE t.playlist_id = p.id) AS track_count
	FROM playlists p
	WHERE p.youtube_id = ?;
	`
	return r.scanPlaylist(r.db.QueryRow(query, youtubeID))
}

func (r *PlaylistRepository) ListByUserID(userID string) ([]domain.Playlist, error) {
	query := `
	SELECT p.id, p.user_id, p.title, p.youtube_id, p.auto_sync, p.sync_interval_minutes,
	       p.last_synced_at, p.status, p.error_message, p.created_at, p.updated_at,
	       (SELECT COUNT(*) FROM tracks t WHERE t.playlist_id = p.id) AS track_count
	FROM playlists p
	WHERE p.user_id = ?
	ORDER BY p.created_at DESC;
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	playlists := make([]domain.Playlist, 0)
	for rows.Next() {
		p, err := r.scanPlaylistRow(rows)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, *p)
	}
	return playlists, nil
}

func (r *PlaylistRepository) ListAutoSync() ([]domain.Playlist, error) {
	query := `
	SELECT p.id, p.user_id, p.title, p.youtube_id, p.auto_sync, p.sync_interval_minutes,
	       p.last_synced_at, p.status, p.error_message, p.created_at, p.updated_at,
	       (SELECT COUNT(*) FROM tracks t WHERE t.playlist_id = p.id) AS track_count
	FROM playlists p
	WHERE p.auto_sync = 1;
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	playlists := make([]domain.Playlist, 0)
	for rows.Next() {
		p, err := r.scanPlaylistRow(rows)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, *p)
	}
	return playlists, nil
}

func (r *PlaylistRepository) Update(p *domain.Playlist) error {
	p.UpdatedAt = time.Now().UTC()
	query := `
	UPDATE playlists SET
		title = ?, youtube_id = ?, auto_sync = ?, sync_interval_minutes = ?,
		last_synced_at = ?, status = ?, error_message = ?, updated_at = ?
	WHERE id = ?;
	`
	autoSyncInt := 0
	if p.AutoSync {
		autoSyncInt = 1
	}

	res, err := r.db.Exec(
		query,
		p.Title, p.YouTubeID, autoSyncInt, p.SyncIntervalMinutes,
		p.LastSyncedAt, string(p.Status), p.ErrorMessage, p.UpdatedAt,
		p.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PlaylistRepository) Delete(id string) error {
	query := `DELETE FROM playlists WHERE id = ?;`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PlaylistRepository) scanPlaylist(row *sql.Row) (*domain.Playlist, error) {
	var p domain.Playlist
	var autoSyncInt int
	var statusStr string
	var lastSyncedAt sql.NullTime

	err := row.Scan(
		&p.ID, &p.UserID, &p.Title, &p.YouTubeID, &autoSyncInt, &p.SyncIntervalMinutes,
		&lastSyncedAt, &statusStr, &p.ErrorMessage, &p.CreatedAt, &p.UpdatedAt,
		&p.TrackCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	p.AutoSync = (autoSyncInt == 1)
	p.Status = domain.PlaylistStatus(statusStr)
	if lastSyncedAt.Valid {
		p.LastSyncedAt = &lastSyncedAt.Time
	}
	return &p, nil
}

func (r *PlaylistRepository) scanPlaylistRow(rows *sql.Rows) (*domain.Playlist, error) {
	var p domain.Playlist
	var autoSyncInt int
	var statusStr string
	var lastSyncedAt sql.NullTime

	err := rows.Scan(
		&p.ID, &p.UserID, &p.Title, &p.YouTubeID, &autoSyncInt, &p.SyncIntervalMinutes,
		&lastSyncedAt, &statusStr, &p.ErrorMessage, &p.CreatedAt, &p.UpdatedAt,
		&p.TrackCount,
	)
	if err != nil {
		return nil, err
	}

	p.AutoSync = (autoSyncInt == 1)
	p.Status = domain.PlaylistStatus(statusStr)
	if lastSyncedAt.Valid {
		p.LastSyncedAt = &lastSyncedAt.Time
	}
	return &p, nil
}
