package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/syncwave/syncwave/internal/domain"
)

type TrackRepository struct {
	db *DB
}

func NewTrackRepository(db *DB) *TrackRepository {
	return &TrackRepository{db: db}
}

func (r *TrackRepository) Create(t *domain.Track) error {
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now

	query := `
	INSERT INTO tracks (
		id, youtube_id, playlist_id, user_id, title, artist, album, duration,
		file_path, cover_path, file_size, format, bitrate, status,
		error_message, downloaded_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`

	_, err := r.db.Exec(
		query,
		t.ID, t.YouTubeID, t.PlaylistID, t.UserID, t.Title, t.Artist, t.Album, t.Duration,
		t.FilePath, t.CoverPath, t.FileSize, t.Format, t.Bitrate, string(t.Status),
		t.ErrorMessage, t.DownloadedAt, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (r *TrackRepository) GetByID(id string, userID string) (*domain.Track, error) {
	query := `
	SELECT id, youtube_id, playlist_id, user_id, title, artist, album, duration,
	       file_path, cover_path, file_size, format, bitrate, status,
	       error_message, downloaded_at, created_at, updated_at
	FROM tracks WHERE id = ?`
	var args []interface{}
	args = append(args, id)

	if userID != "" {
		query += " AND user_id = ?"
		args = append(args, userID)
	}
	query += ";"

	return r.scanTrack(r.db.QueryRow(query, args...))
}

func (r *TrackRepository) GetByYouTubeID(youtubeID string, userID string) (*domain.Track, error) {
	query := `
	SELECT id, youtube_id, playlist_id, user_id, title, artist, album, duration,
	       file_path, cover_path, file_size, format, bitrate, status,
	       error_message, downloaded_at, created_at, updated_at
	FROM tracks WHERE youtube_id = ?`
	var args []interface{}
	args = append(args, youtubeID)

	if userID != "" {
		query += " AND user_id = ?"
		args = append(args, userID)
	}
	query += ";"

	return r.scanTrack(r.db.QueryRow(query, args...))
}

func (r *TrackRepository) GetExistingYouTubeIDs(youtubeIDs []string, userID string) (map[string]bool, error) {
	result := make(map[string]bool)
	if len(youtubeIDs) == 0 {
		return result, nil
	}

	chunkSize := 400
	for i := 0; i < len(youtubeIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(youtubeIDs) {
			end = len(youtubeIDs)
		}
		chunk := youtubeIDs[i:end]

		placeholders := make([]string, len(chunk))
		args := make([]interface{}, 0, len(chunk)+1)
		for idx, id := range chunk {
			placeholders[idx] = "?"
			args = append(args, id)
		}

		userFilter := ""
		if userID != "" {
			userFilter = " AND user_id = ?"
			args = append(args, userID)
		}

		query := fmt.Sprintf("SELECT youtube_id FROM tracks WHERE youtube_id IN (%s)%s", strings.Join(placeholders, ","), userFilter)
		rows, err := r.db.Query(query, args...)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var ytID string
			if err := rows.Scan(&ytID); err == nil {
				result[ytID] = true
			}
		}
		_ = rows.Close()
	}

	return result, nil
}

func (r *TrackRepository) List(filter domain.TrackFilter) (*domain.TrackListResult, error) {
	where := []string{"1=1"}
	args := []interface{}{}

	if filter.UserID != "" {
		where = append(where, "user_id = ?")
		args = append(args, filter.UserID)
	}

	if filter.Query != "" {
		where = append(where, "(title LIKE ? OR artist LIKE ? OR album LIKE ?)")
		searchTerm := "%" + filter.Query + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	if filter.PlaylistID != "" {
		where = append(where, "playlist_id = ?")
		args = append(args, filter.PlaylistID)
	}

	if filter.Status != "" {
		where = append(where, "status = ?")
		args = append(args, string(filter.Status))
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tracks WHERE %s", whereClause)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Sorting
	sortBy := "created_at"
	switch filter.SortBy {
	case "title":
		sortBy = "title"
	case "artist":
		sortBy = "artist"
	case "duration":
		sortBy = "duration"
	case "created_at":
		sortBy = "created_at"
	}

	order := "DESC"
	if strings.ToLower(filter.Order) == "asc" {
		order = "ASC"
	}

	// Pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 500 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
	SELECT id, youtube_id, playlist_id, user_id, title, artist, album, duration,
	       file_path, cover_path, file_size, format, bitrate, status,
	       error_message, downloaded_at, created_at, updated_at
	FROM tracks
	WHERE %s
	ORDER BY %s %s
	LIMIT ? OFFSET ?;
	`, whereClause, sortBy, order)

	queryArgs := append(args, pageSize, offset)
	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tracks := make([]domain.Track, 0)
	for rows.Next() {
		var t domain.Track
		var statusStr string
		var downloadedAt sql.NullTime
		var playlistID sql.NullString
		var userID sql.NullString

		err := rows.Scan(
			&t.ID, &t.YouTubeID, &playlistID, &userID, &t.Title, &t.Artist, &t.Album, &t.Duration,
			&t.FilePath, &t.CoverPath, &t.FileSize, &t.Format, &t.Bitrate, &statusStr,
			&t.ErrorMessage, &downloadedAt, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		t.Status = domain.TrackStatus(statusStr)
		if playlistID.Valid {
			t.PlaylistID = &playlistID.String
		}
		if userID.Valid {
			t.UserID = userID.String
		}
		if downloadedAt.Valid {
			t.DownloadedAt = &downloadedAt.Time
		}
		tracks = append(tracks, t)
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return &domain.TrackListResult{
		Tracks:     tracks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *TrackRepository) GetAllReady(userID string) ([]domain.Track, error) {
	query := `
	SELECT id, youtube_id, playlist_id, user_id, title, artist, album, duration,
	       file_path, cover_path, file_size, format, bitrate, status,
	       error_message, downloaded_at, created_at, updated_at
	FROM tracks
	WHERE status = 'ready'`
	var args []interface{}

	if userID != "" {
		query += " AND user_id = ?"
		args = append(args, userID)
	}
	query += " ORDER BY created_at DESC;"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tracks := make([]domain.Track, 0)
	for rows.Next() {
		var t domain.Track
		var statusStr string
		var downloadedAt sql.NullTime
		var playlistID sql.NullString
		var uID sql.NullString

		err := rows.Scan(
			&t.ID, &t.YouTubeID, &playlistID, &uID, &t.Title, &t.Artist, &t.Album, &t.Duration,
			&t.FilePath, &t.CoverPath, &t.FileSize, &t.Format, &t.Bitrate, &statusStr,
			&t.ErrorMessage, &downloadedAt, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		t.Status = domain.TrackStatus(statusStr)
		if playlistID.Valid {
			t.PlaylistID = &playlistID.String
		}
		if uID.Valid {
			t.UserID = uID.String
		}
		if downloadedAt.Valid {
			t.DownloadedAt = &downloadedAt.Time
		}
		tracks = append(tracks, t)
	}
	return tracks, nil
}

func (r *TrackRepository) Update(t *domain.Track) error {
	t.UpdatedAt = time.Now().UTC()
	query := `
	UPDATE tracks SET
		playlist_id = ?, user_id = ?, title = ?, artist = ?, album = ?, duration = ?,
		file_path = ?, cover_path = ?, file_size = ?, format = ?, bitrate = ?,
		status = ?, error_message = ?, downloaded_at = ?, updated_at = ?
	WHERE id = ?;
	`
	_, err := r.db.Exec(
		query,
		t.PlaylistID, t.UserID, t.Title, t.Artist, t.Album, t.Duration,
		t.FilePath, t.CoverPath, t.FileSize, t.Format, t.Bitrate,
		string(t.Status), t.ErrorMessage, t.DownloadedAt, t.UpdatedAt,
		t.ID,
	)
	return err
}

func (r *TrackRepository) Delete(id string, userID string) error {
	query := `DELETE FROM tracks WHERE id = ?`
	var args []interface{}
	args = append(args, id)

	if userID != "" {
		query += " AND user_id = ?"
		args = append(args, userID)
	}
	query += ";"

	res, err := r.db.Exec(query, args...)
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

func (r *TrackRepository) BatchDelete(ids []string, userID string) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, 0, len(ids)+1)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	userFilter := ""
	if userID != "" {
		userFilter = " AND user_id = ?"
		args = append(args, userID)
	}

	query := fmt.Sprintf("DELETE FROM tracks WHERE id IN (%s)%s;", strings.Join(placeholders, ","), userFilter)
	_, err := r.db.Exec(query, args...)
	return err
}

func (r *TrackRepository) CleanBrokenTracks() error {
	query := `DELETE FROM tracks WHERE status IN ('downloading', 'failed', 'queued');`
	_, err := r.db.Exec(query)
	return err
}

func (r *TrackRepository) GetStats(userID string) (*domain.TrackStats, error) {
	stats := &domain.TrackStats{}

	query := `
	SELECT 
		COUNT(*),
		COALESCE(SUM(CASE WHEN status = 'ready' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(file_size), 0),
		COALESCE(SUM(duration), 0)
	FROM tracks`

	var args []interface{}
	if userID != "" {
		query += " WHERE user_id = ?"
		args = append(args, userID)
	}
	query += ";"

	err := r.db.QueryRow(query, args...).Scan(
		&stats.TotalTracks,
		&stats.ReadyTracks,
		&stats.FailedTracks,
		&stats.TotalStorageSize,
		&stats.TotalDuration,
	)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *TrackRepository) scanTrack(row *sql.Row) (*domain.Track, error) {
	var t domain.Track
	var statusStr string
	var downloadedAt sql.NullTime
	var playlistID sql.NullString
	var userID sql.NullString

	err := row.Scan(
		&t.ID, &t.YouTubeID, &playlistID, &userID, &t.Title, &t.Artist, &t.Album, &t.Duration,
		&t.FilePath, &t.CoverPath, &t.FileSize, &t.Format, &t.Bitrate, &statusStr,
		&t.ErrorMessage, &downloadedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	t.Status = domain.TrackStatus(statusStr)
	if playlistID.Valid {
		t.PlaylistID = &playlistID.String
	}
	if userID.Valid {
		t.UserID = userID.String
	}
	if downloadedAt.Valid {
		t.DownloadedAt = &downloadedAt.Time
	}
	return &t, nil
}
