package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

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
		storage_quota_bytes INTEGER NOT NULL DEFAULT 10737418240,
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
		youtube_id TEXT NOT NULL,
		playlist_id TEXT,
		user_id TEXT,
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
		FOREIGN KEY (playlist_id) REFERENCES playlists(id) ON DELETE SET NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
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
		user_id TEXT NOT NULL DEFAULT '',
		playlist_id TEXT,
		track_id TEXT,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_sync_logs_created_at ON sync_logs(created_at DESC);

	CREATE TABLE IF NOT EXISTS blacklist (
		youtube_id TEXT NOT NULL,
		user_id TEXT NOT NULL DEFAULT '',
		title TEXT NOT NULL,
		artist TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (youtube_id, user_id)
	);
	`

	if _, err := db.Exec(baseSchema); err != nil {
		return err
	}

	// 1. Safe migrations for existing databases (ensure all columns exist before creating indexes on them)
	addColumnIfNotExists(db.DB, "users", "storage_quota_bytes INTEGER NOT NULL DEFAULT 10737418240", "storage_quota_bytes")
	addColumnIfNotExists(db.DB, "tracks", "user_id TEXT", "user_id")
	addColumnIfNotExists(db.DB, "blacklist", "user_id TEXT NOT NULL DEFAULT ''", "user_id")
	addColumnIfNotExists(db.DB, "sync_logs", "user_id TEXT NOT NULL DEFAULT ''", "user_id")

	// 2. Safe index creations (after columns are guaranteed to exist in old and new DBs)
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_tracks_user_id ON tracks(user_id);")
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_tracks_user_yt ON tracks(user_id, youtube_id);")
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_blacklist_user_id ON blacklist(user_id);")
	_, _ = db.Exec("CREATE INDEX IF NOT EXISTS idx_sync_logs_user_id ON sync_logs(user_id);")

	// 3. Populate user_id from parent playlists if empty
	_, _ = db.Exec(`
		UPDATE tracks 
		SET user_id = (SELECT user_id FROM playlists WHERE playlists.id = tracks.playlist_id)
		WHERE (user_id IS NULL OR user_id = '') AND playlist_id IS NOT NULL;
	`)
	_, _ = db.Exec(`
		UPDATE sync_logs 
		SET user_id = (SELECT user_id FROM playlists WHERE playlists.id = sync_logs.playlist_id)
		WHERE (user_id IS NULL OR user_id = '') AND playlist_id IS NOT NULL;
	`)

	// 4. Default settings
	_, _ = db.Exec("INSERT OR IGNORE INTO settings (key, value) VALUES ('allow_registration', '0');")
	_, _ = db.Exec("INSERT OR IGNORE INTO settings (key, value) VALUES ('global_storage_limit_bytes', '0');")
	_, _ = db.Exec("INSERT OR IGNORE INTO settings (key, value) VALUES ('default_user_quota_bytes', '10737418240');")

	// 5. Clean up tracks with artist 'NA' and 'Artist - Title' format
	_, _ = db.Exec(`
		UPDATE tracks 
		SET artist = TRIM(SUBSTR(title, 1, INSTR(title, ' - ') - 1)),
		    title = TRIM(SUBSTR(title, INSTR(title, ' - ') + 3))
		WHERE (artist = 'NA' OR artist = '' OR artist = 'Unknown Artist') 
		  AND INSTR(title, ' - ') > 0;
	`)

	return nil
}

func addColumnIfNotExists(db *sql.DB, tableName, columnDef string, columnName string) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s);", tableName))
	if err != nil {
		return
	}
	defer rows.Close()

	exists := false
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil {
			if strings.EqualFold(name, columnName) {
				exists = true
				break
			}
		}
	}
	if !exists {
		_, _ = db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", tableName, columnDef))
	}
}
