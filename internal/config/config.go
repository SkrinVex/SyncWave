package config

import (
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Host                string
	Port                string
	DataDir             string
	DBPath              string
	CookiesPath         string
	MusicDir            string
	CoversDir           string
	JWTSecret           string
	SyncIntervalMinutes int
	MaxConcurrent       int
	HTTPProxy           string
	AudioFormat         string
	YtDlpPath           string
	FFmpegPath          string
}

func Load() *Config {
	dataDir := getEnv("DATA_DIR", "./data")
	if !filepath.IsAbs(dataDir) {
		cwd, err := os.Getwd()
		if err == nil {
			dataDir = filepath.Join(cwd, dataDir)
		}
	}

	_ = os.MkdirAll(dataDir, 0755)
	musicDir := filepath.Join(dataDir, "music")
	coversDir := filepath.Join(dataDir, "covers")
	_ = os.MkdirAll(musicDir, 0755)
	_ = os.MkdirAll(coversDir, 0755)

	syncInterval, _ := strconv.Atoi(getEnv("SYNC_INTERVAL_MINUTES", "60"))
	if syncInterval <= 0 {
		syncInterval = 60
	}

	maxConcurrent, _ := strconv.Atoi(getEnv("MAX_CONCURRENT_DOWNLOADS", "2"))
	if maxConcurrent <= 0 {
		maxConcurrent = 2
	}

	return &Config{
		Host:                getEnv("HOST", ""),
		Port:                getEnv("PORT", "8080"),
		DataDir:             dataDir,
		DBPath:              filepath.Join(dataDir, "syncwave.db"),
		CookiesPath:         filepath.Join(dataDir, "cookies.txt"),
		MusicDir:            musicDir,
		CoversDir:           coversDir,
		JWTSecret:           getEnv("JWT_SECRET", "syncwave-super-secret-production-key-32-chars-min!"),
		SyncIntervalMinutes: syncInterval,
		MaxConcurrent:       maxConcurrent,
		HTTPProxy:           getEnv("HTTP_PROXY", ""),
		AudioFormat:         getEnv("AUDIO_FORMAT", "opus"),
		YtDlpPath:           getEnv("YTDLP_PATH", "yt-dlp"),
		FFmpegPath:          getEnv("FFMPEG_PATH", "ffmpeg"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
