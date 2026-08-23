package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/syncwave/syncwave/internal/domain"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u *domain.User) error {
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now
	if u.StorageQuotaBytes <= 0 {
		u.StorageQuotaBytes = 10737418240 // 10 GB default
	}

	query := `
	INSERT INTO users (id, username, password_hash, is_admin, storage_quota_bytes, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	isAdminInt := 0
	if u.IsAdmin {
		isAdminInt = 1
	}

	_, err := r.db.Exec(query, u.ID, u.Username, u.PasswordHash, isAdminInt, u.StorageQuotaBytes, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetByID(id string) (*domain.User, error) {
	query := `SELECT id, username, password_hash, is_admin, storage_quota_bytes, created_at, updated_at FROM users WHERE id = ?;`
	row := r.db.QueryRow(query, id)

	var u domain.User
	var isAdminInt int
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &isAdminInt, &u.StorageQuotaBytes, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	u.IsAdmin = (isAdminInt == 1)
	return &u, nil
}

func (r *UserRepository) GetByUsername(username string) (*domain.User, error) {
	query := `SELECT id, username, password_hash, is_admin, storage_quota_bytes, created_at, updated_at FROM users WHERE username = ?;`
	row := r.db.QueryRow(query, username)

	var u domain.User
	var isAdminInt int
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &isAdminInt, &u.StorageQuotaBytes, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	u.IsAdmin = (isAdminInt == 1)
	return &u, nil
}

func (r *UserRepository) Count() (int, error) {
	query := `SELECT COUNT(*) FROM users;`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

func (r *UserRepository) ListWithStats() ([]domain.UserWithStats, error) {
	query := `
	SELECT 
		u.id, 
		u.username, 
		u.is_admin, 
		u.storage_quota_bytes, 
		u.created_at, 
		u.updated_at,
		COALESCE((SELECT COUNT(*) FROM playlists WHERE user_id = u.id), 0) as playlists_count,
		COALESCE((
			SELECT COUNT(t.id) 
			FROM tracks t 
			JOIN playlists p ON t.playlist_id = p.id 
			WHERE p.user_id = u.id
		), 0) as tracks_count,
		COALESCE((
			SELECT SUM(t.file_size) 
			FROM tracks t 
			JOIN playlists p ON t.playlist_id = p.id 
			WHERE p.user_id = u.id
		), 0) as storage_used_bytes
	FROM users u
	ORDER BY u.created_at ASC;
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.UserWithStats
	for rows.Next() {
		var item domain.UserWithStats
		var isAdminInt int
		if err := rows.Scan(
			&item.ID,
			&item.Username,
			&isAdminInt,
			&item.StorageQuotaBytes,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.PlaylistsCount,
			&item.TracksCount,
			&item.StorageUsedBytes,
		); err != nil {
			return nil, err
		}
		item.IsAdmin = (isAdminInt == 1)
		result = append(result, item)
	}

	if result == nil {
		result = []domain.UserWithStats{}
	}
	return result, nil
}

func (r *UserRepository) UpdateQuota(userID string, quota int64) error {
	query := `UPDATE users SET storage_quota_bytes = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;`
	res, err := r.db.Exec(query, quota, userID)
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

func (r *UserRepository) Update(u *domain.User) error {
	u.UpdatedAt = time.Now().UTC()
	query := `
	UPDATE users SET username = ?, password_hash = ?, is_admin = ?, updated_at = ?
	WHERE id = ?;
	`
	isAdminInt := 0
	if u.IsAdmin {
		isAdminInt = 1
	}

	res, err := r.db.Exec(query, u.Username, u.PasswordHash, isAdminInt, u.UpdatedAt, u.ID)
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

func (r *UserRepository) Delete(id string) error {
	query := `DELETE FROM users WHERE id = ?;`
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
