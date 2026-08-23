package usecase

import (
	"context"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/infrastructure/ytdlp"
)

type SettingsUsecase struct {
	settingsRepo domain.SettingsRepository
	trackRepo    domain.TrackRepository
	playlistRepo domain.PlaylistRepository
	userRepo     domain.UserRepository
	ytdlpClient  *ytdlp.Client
	dataDir      string
	dbPath       string
}

func NewSettingsUsecase(
	settingsRepo domain.SettingsRepository,
	trackRepo domain.TrackRepository,
	playlistRepo domain.PlaylistRepository,
	userRepo domain.UserRepository,
	ytdlpClient *ytdlp.Client,
	dataDir string,
	dbPath string,
) *SettingsUsecase {
	return &SettingsUsecase{
		settingsRepo: settingsRepo,
		trackRepo:    trackRepo,
		playlistRepo: playlistRepo,
		userRepo:     userRepo,
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

	allowReg := settingsMap["allow_registration"] == "1" || settingsMap["allow_registration"] == "true"
	globalLimit, _ := strconv.ParseInt(settingsMap["global_storage_limit_bytes"], 10, 64)
	defaultQuota, _ := strconv.ParseInt(settingsMap["default_user_quota_bytes"], 10, 64)
	if defaultQuota <= 0 {
		defaultQuota = 10737418240
	}

	ytdlpVer, _ := u.ytdlpClient.GetYTDLPVersion()
	ffmpegVer, _ := u.ytdlpClient.GetFFmpegVersion()

	// User-specific cookies validation
	cookieVal := u.ytdlpClient.ValidateUserCookies(userID)
	cookiesMod, _ := u.ytdlpClient.GetUserCookiesModTime(userID)
	var expiresAtStr string
	if cookieVal.ExpiresAt != nil {
		expiresAtStr = cookieVal.ExpiresAt.Format(time.RFC3339)
	}

	stats, _ := u.trackRepo.GetStats(userID)
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

	// Host Physical Disk usage
	var hostTotal, hostFree, hostUsed uint64
	var stat syscall.Statfs_t
	checkDir := u.dataDir
	if checkDir == "" {
		checkDir = "/"
	}
	if err := syscall.Statfs(checkDir, &stat); err == nil {
		hostTotal = stat.Blocks * uint64(stat.Bsize)
		hostFree = stat.Bavail * uint64(stat.Bsize)
		hostUsed = hostTotal - hostFree
	}

	// User-specific stats
	var userQuota int64
	var userUsage int64
	var isAdmin bool
	if user, err := u.userRepo.GetByID(userID); err == nil && user != nil {
		userQuota = user.StorageQuotaBytes
		isAdmin = user.IsAdmin
	}

	// Calculate user specific storage usage
	allUserStats, _ := u.userRepo.ListWithStats()
	for _, uStats := range allUserStats {
		if uStats.ID == userID {
			userUsage = uStats.StorageUsedBytes
			break
		}
	}

	return &domain.SystemSettings{
		HTTPProxy:               httpProxy,
		AudioFormat:             audioFormat,
		AudioQuality:            "0 (Best)",
		MaxConcurrent:           maxConcurrent,
		AllowRegistration:       allowReg,
		GlobalStorageLimitBytes: globalLimit,
		DefaultUserQuotaBytes:   defaultQuota,
		HasCookies:              cookieVal.HasCookies,
		CookiesValid:            cookieVal.IsValid,
		CookiesStatus:           string(cookieVal.Status),
		CookiesExpiresAt:        expiresAtStr,
		CookiesError:            cookieVal.ErrorReason,
		CookiesUpdatedAt:        cookiesMod,
		YTDLPVersion:            ytdlpVer,
		FFmpegVersion:           ffmpegVer,
		StorageUsageBytes:       storageUsage,
		DatabaseSizeBytes:       dbSize,
		TotalTracksCount:        totalTracks,
		TotalPlaylistsCount:     len(playlists),
		UserStorageUsageBytes:   userUsage,
		UserStorageQuotaBytes:   userQuota,
		HostDiskTotalBytes:      hostTotal,
		HostDiskUsedBytes:       hostUsed,
		HostDiskFreeBytes:       hostFree,
		IsAdmin:                 isAdmin,
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

func (u *SettingsUsecase) SetAllowRegistration(allowed bool) error {
	val := "0"
	if allowed {
		val = "1"
	}
	return u.settingsRepo.Set("allow_registration", val)
}

func (u *SettingsUsecase) SetGlobalStorageLimit(bytes int64) error {
	return u.settingsRepo.Set("global_storage_limit_bytes", strconv.FormatInt(bytes, 10))
}

func (u *SettingsUsecase) SaveCookies(userID string, content []byte) error {
	return u.ytdlpClient.SaveUserCookies(userID, content)
}

func (u *SettingsUsecase) DeleteCookies(userID string) error {
	return u.ytdlpClient.DeleteUserCookies(userID)
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
