package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	// Enable WAL mode, busy timeout, and foreign keys in DSN
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=synchronous(NORMAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// SQLite handles concurrency best with WAL and a connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	sqliteDB := &DB{DB: db}
	if err := sqliteDB.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return sqliteDB, nil
}

func (db *DB) Migrate() error {
	baseSchema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		is_admin INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS playlists (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		youtube_id TEXT NOT NULL,
		auto_sync INTEGER NOT NULL DEFAULT 1,
		sync_interval_minutes INTEGER NOT NULL DEFAULT 60,
		last_synced_at DATETIME,
		status TEXT NOT NULL DEFAULT 'idle',
		error_message TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_playlists_user_id ON playlists(user_id);
	CREATE INDEX IF NOT EXISTS idx_playlists_youtube_id ON playlists(youtube_id);

	CREATE TABLE IF NOT EXISTS tracks (
		id TEXT PRIMARY KEY,
		youtube_id TEXT UNIQUE NOT NULL,
		playlist_id TEXT,
		title TEXT NOT NULL,
		artist TEXT NOT NULL,
		album TEXT NOT NULL DEFAULT '',
		duration INTEGER NOT NULL DEFAULT 0,
		file_path TEXT NOT NULL DEFAULT '',
		cover_path TEXT NOT NULL DEFAULT '',
		file_size INTEGER NOT NULL DEFAULT 0,
		format TEXT NOT NULL DEFAULT 'opus',
		bitrate INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'ready',
		error_message TEXT NOT NULL DEFAULT '',
		downloaded_at DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_tracks_youtube_id ON tracks(youtube_id);
	CREATE INDEX IF NOT EXISTS idx_tracks_playlist_id ON tracks(playlist_id);
	CREATE INDEX IF NOT EXISTS idx_tracks_artist ON tracks(artist);
	CREATE INDEX IF NOT EXISTS idx_tracks_title ON tracks(title);
	CREATE INDEX IF NOT EXISTS idx_tracks_status ON tracks(status);
	CREATE INDEX IF NOT EXISTS idx_tracks_created_at ON tracks(created_at DESC);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sync_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		playlist_id TEXT,
		track_id TEXT,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_sync_logs_created_at ON sync_logs(created_at DESC);

	CREATE TABLE IF NOT EXISTS blacklist (
		youtube_id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		artist TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(baseSchema); err != nil {
		return err
	}

	// 1. Safe migrations for existing databases (alter table column additions)
	_, _ = db.Exec("ALTER TABLE users ADD COLUMN storage_quota_bytes INTEGER NOT NULL DEFAULT 10737418240;")
	_, _ = db.Exec("ALTER TABLE tracks ADD COLUMN user_id TEXT;")
	_, _ = db.Exec("ALTER TABLE blacklist ADD COLUMN user_id TEXT;")

	// 2. Safe index creations after columns are guaranteed to exist
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_tracks_user_id ON tracks(user_id);")

	// 3. Populate tracks.user_id from parent playlists if empty
	_, _ = db.Exec(`
		UPDATE tracks 
		SET user_id = (SELECT user_id FROM playlists WHERE playlists.id = tracks.playlist_id)
		WHERE (user_id IS NULL OR user_id = '') AND playlist_id IS NOT NULL;
	`)

	// 4. Default settings
	_, _ = db.Exec("INSERT OR IGNORE INTO settings (key, value) VALUES ('allow_registration', '0');")
	_, _ = db.Exec("INSERT OR IGNORE INTO settings (key, value) VALUES ('global_storage_limit_bytes', '0');")
	_, _ = db.Exec("INSERT OR IGNORE INTO settings (key, value) VALUES ('default_user_quota_bytes', '10737418240');")

	return nil
}
