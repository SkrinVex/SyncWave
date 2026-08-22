package sqlite

import (
	"database/sql"
	"errors"
	"time"

	"github.com/syncwave/syncwave/internal/domain"
)

type SettingsRepository struct {
	db *DB
}

func NewSettingsRepository(db *DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) Get(key string) (string, error) {
	query := `SELECT value FROM settings WHERE key = ?;`
	var val string
	err := r.db.QueryRow(query, key).Scan(&val)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", err
	}
	return val, nil
}

func (r *SettingsRepository) Set(key, value string) error {
	query := `
	INSERT INTO settings (key, value, updated_at)
	VALUES (?, ?, ?)
	ON CONFLICT(key) DO UPDATE SET
		value = excluded.value,
		updated_at = excluded.updated_at;
	`
	_, err := r.db.Exec(query, key, value, time.Now().UTC())
	return err
}

func (r *SettingsRepository) GetAll() (map[string]string, error) {
	query := `SELECT key, value FROM settings;`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		res[k] = v
	}
	return res, nil
}
