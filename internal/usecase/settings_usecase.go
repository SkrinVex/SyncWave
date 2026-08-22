package usecase

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/infrastructure/ytdlp"
)

type SettingsUsecase struct {
	settingsRepo domain.SettingsRepository
	trackRepo    domain.TrackRepository
	playlistRepo domain.PlaylistRepository
	ytdlpClient  *ytdlp.Client
	dataDir      string
	dbPath       string
}

func NewSettingsUsecase(
	settingsRepo domain.SettingsRepository,
	trackRepo domain.TrackRepository,
	playlistRepo domain.PlaylistRepository,
	ytdlpClient *ytdlp.Client,
	dataDir string,
	dbPath string,
) *SettingsUsecase {
	return &SettingsUsecase{
		settingsRepo: settingsRepo,
		trackRepo:    trackRepo,
		playlistRepo: playlistRepo,
		ytdlpClient:  ytdlpClient,
		dataDir:      dataDir,
		dbPath:       dbPath,
	}
}

type UpdateSettingsRequest struct {
	HTTPProxy     *string `json:"http_proxy"`
	AudioFormat   *string `json:"audio_format"`
	MaxConcurrent *int    `json:"max_concurrent"`
}

func (u *SettingsUsecase) GetSystemSettings(userID string) (*domain.SystemSettings, error) {
	settingsMap, _ := u.settingsRepo.GetAll()

	httpProxy := settingsMap["http_proxy"]
	audioFormat := settingsMap["audio_format"]
	if audioFormat == "" {
		audioFormat = "opus"
	}
	maxConcurrent, _ := strconv.Atoi(settingsMap["max_concurrent"])
	if maxConcurrent <= 0 {
		maxConcurrent = 2
	}

	ytdlpVer, _ := u.ytdlpClient.GetYTDLPVersion()
	ffmpegVer, _ := u.ytdlpClient.GetFFmpegVersion()

	hasCookies := u.ytdlpClient.HasCookies()
	cookiesMod, _ := u.ytdlpClient.GetCookiesModTime()

	stats, _ := u.trackRepo.GetStats()
	playlists, _ := u.playlistRepo.ListByUserID(userID)

	var dbSize int64
	if fi, err := os.Stat(u.dbPath); err == nil {
		dbSize = fi.Size()
	}

	totalTracks := 0
	var storageUsage int64
	if stats != nil {
		totalTracks = stats.TotalTracks
		storageUsage = stats.TotalStorageSize
	}

	return &domain.SystemSettings{
		HTTPProxy:           httpProxy,
		AudioFormat:         audioFormat,
		AudioQuality:        "0 (Best)",
		MaxConcurrent:       maxConcurrent,
		HasCookies:          hasCookies,
		CookiesValid:        hasCookies,
		CookiesUpdatedAt:    cookiesMod,
		YTDLPVersion:        ytdlpVer,
		FFmpegVersion:       ffmpegVer,
		StorageUsageBytes:   storageUsage,
		DatabaseSizeBytes:   dbSize,
		TotalTracksCount:    totalTracks,
		TotalPlaylistsCount: len(playlists),
	}, nil
}

func (u *SettingsUsecase) UpdateSettings(req UpdateSettingsRequest) error {
	if req.HTTPProxy != nil {
		_ = u.settingsRepo.Set("http_proxy", *req.HTTPProxy)
		u.ytdlpClient.SetProxy(*req.HTTPProxy)
	}
	if req.AudioFormat != nil && *req.AudioFormat != "" {
		_ = u.settingsRepo.Set("audio_format", *req.AudioFormat)
		u.ytdlpClient.SetAudioFormat(*req.AudioFormat)
	}
	if req.MaxConcurrent != nil && *req.MaxConcurrent > 0 {
		_ = u.settingsRepo.Set("max_concurrent", strconv.Itoa(*req.MaxConcurrent))
	}
	return nil
}

func (u *SettingsUsecase) SaveCookies(content []byte) error {
	return u.ytdlpClient.SaveCookies(content)
}

func (u *SettingsUsecase) DeleteCookies() error {
	cookiesPath := filepath.Join(u.dataDir, "cookies.txt")
	return os.Remove(cookiesPath)
}

func (u *SettingsUsecase) TestProxy(proxyURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return u.ytdlpClient.TestProxyConnection(ctx, proxyURL)
}

func (u *SettingsUsecase) InitFromDB() {
	if proxy, err := u.settingsRepo.Get("http_proxy"); err == nil && proxy != "" {
		u.ytdlpClient.SetProxy(proxy)
	}
	if format, err := u.settingsRepo.Get("audio_format"); err == nil && format != "" {
		u.ytdlpClient.SetAudioFormat(format)
	}
}
