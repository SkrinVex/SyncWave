package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/infrastructure/worker"
)

type TrackUsecase struct {
	trackRepo     domain.TrackRepository
	blacklistRepo domain.BlacklistRepository
	userRepo      domain.UserRepository
	settingsRepo  domain.SettingsRepository
	eventHub      *worker.EventHub
	musicDir      string
	coversDir     string
	ffmpegPath    string
}

func NewTrackUsecase(
	trackRepo domain.TrackRepository,
	blacklistRepo domain.BlacklistRepository,
	userRepo domain.UserRepository,
	settingsRepo domain.SettingsRepository,
	eventHub *worker.EventHub,
	musicDir string,
	coversDir string,
	ffmpegPath string,
) *TrackUsecase {
	_ = os.MkdirAll(musicDir, 0755)
	_ = os.MkdirAll(coversDir, 0755)
	return &TrackUsecase{
		trackRepo:     trackRepo,
		blacklistRepo: blacklistRepo,
		userRepo:      userRepo,
		settingsRepo:  settingsRepo,
		eventHub:      eventHub,
		musicDir:      musicDir,
		coversDir:     coversDir,
		ffmpegPath:    ffmpegPath,
	}
}

type UploadTrackInput struct {
	Filename string
	Reader   io.Reader
	Size     int64
}

type UploadResult struct {
	Uploaded []*domain.Track `json:"uploaded"`
	Errors   []string        `json:"errors"`
}

func (u *TrackUsecase) List(filter domain.TrackFilter) (*domain.TrackListResult, error) {
	return u.trackRepo.List(filter)
}

func (u *TrackUsecase) GetByID(id string, userID string) (*domain.Track, error) {
	return u.trackRepo.GetByID(id, userID)
}

func (u *TrackUsecase) GetAllReady(userID string, playlistID string) ([]domain.Track, error) {
	return u.trackRepo.GetAllReady(userID, playlistID)
}

func (u *TrackUsecase) Delete(id string, userID string) error {
	track, err := u.trackRepo.GetByID(id, userID)
	if err != nil {
		return err
	}

	// Add to blacklist before deletion
	_ = u.blacklistRepo.Add(&domain.BlacklistItem{
		YouTubeID: track.YouTubeID,
		UserID:    userID,
		Title:     track.Title,
		Artist:    track.Artist,
	})

	err = u.trackRepo.Delete(id, userID)
	if err != nil {
		return err
	}

	// Check if any other user's track references the physical audio or cover file before deleting from disk
	if track.FilePath != "" {
		if count, _ := u.trackRepo.CountTracksByFilePath(track.FilePath); count == 0 {
			_ = os.Remove(track.FilePath)
		}
	}
	if track.CoverPath != "" {
		_ = os.Remove(track.CoverPath)
	}

	return nil
}

func (u *TrackUsecase) BatchDelete(ids []string, userID string) error {
	for _, id := range ids {
		if track, err := u.trackRepo.GetByID(id, userID); err == nil && track != nil {
			_ = u.blacklistRepo.Add(&domain.BlacklistItem{
				YouTubeID: track.YouTubeID,
				UserID:    userID,
				Title:     track.Title,
				Artist:    track.Artist,
			})

			_ = u.trackRepo.Delete(id, userID)

			if track.FilePath != "" {
				if count, _ := u.trackRepo.CountTracksByFilePath(track.FilePath); count == 0 {
					_ = os.Remove(track.FilePath)
				}
			}
			if track.CoverPath != "" {
				_ = os.Remove(track.CoverPath)
			}
		}
	}
	return nil
}

func (u *TrackUsecase) CleanBrokenTracks() error {
	return u.trackRepo.CleanBrokenTracks()
}

func (u *TrackUsecase) GetStats(userID string) (*domain.TrackStats, error) {
	return u.trackRepo.GetStats(userID)
}

// UploadTracks processes manual audio uploads from PC / phone
func (u *TrackUsecase) UploadTracks(ctx context.Context, userID string, playlistID string, files []UploadTrackInput) (*UploadResult, error) {
	result := &UploadResult{
		Uploaded: make([]*domain.Track, 0),
		Errors:   make([]string, 0),
	}

	validExts := map[string]bool{
		".mp3":  true,
		".flac": true,
		".m4a":  true,
		".opus": true,
		".ogg":  true,
		".wav":  true,
		".aac":  true,
		".webm": true,
		".wma":  true,
		".mp4":  true,
	}

	for _, fileInput := range files {
		ext := strings.ToLower(filepath.Ext(fileInput.Filename))
		if !validExts[ext] {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: неподдерживаемый аудиоформат (%s)", fileInput.Filename, ext))
			continue
		}

		// 1. Quota Check for User
		if u.userRepo != nil && userID != "" {
			if user, err := u.userRepo.GetByID(userID); err == nil && user != nil && user.StorageQuotaBytes > 0 {
				allStats, _ := u.userRepo.ListWithStats()
				for _, st := range allStats {
					if st.ID == user.ID && (st.StorageUsedBytes+fileInput.Size) > user.StorageQuotaBytes {
						result.Errors = append(result.Errors, fmt.Sprintf("%s: превышена квота дискового пространства пользователя", fileInput.Filename))
						continue
					}
				}
			}
		}

		// 2. Global Storage Limit Check
		if u.settingsRepo != nil {
			if limitStr, err := u.settingsRepo.Get("global_storage_limit_bytes"); err == nil && limitStr != "" {
				if globalLimit, _ := strconv.ParseInt(limitStr, 10, 64); globalLimit > 0 {
					if stats, _ := u.trackRepo.GetStats(""); stats != nil && (stats.TotalStorageSize+fileInput.Size) > globalLimit {
						result.Errors = append(result.Errors, fmt.Sprintf("%s: превышен общий лимит дискового пространства сервера", fileInput.Filename))
						continue
					}
				}
			}
		}

		trackID := uuid.New().String()
		destFileName := fmt.Sprintf("%s%s", trackID, ext)
		destFilePath := filepath.Join(u.musicDir, destFileName)

		outFile, err := os.Create(destFilePath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: не удалось сохранить файл на сервере: %v", fileInput.Filename, err))
			continue
		}

		written, err := io.Copy(outFile, fileInput.Reader)
		_ = outFile.Close()
		if err != nil {
			_ = os.Remove(destFilePath)
			result.Errors = append(result.Errors, fmt.Sprintf("%s: ошибка при записи файла: %v", fileInput.Filename, err))
			continue
		}

		if written < 512 {
			_ = os.Remove(destFilePath)
			result.Errors = append(result.Errors, fmt.Sprintf("%s: загруженный файл пуст или поврежден", fileInput.Filename))
			continue
		}

		// 3. Extract Metadata and Cover Art
		meta := u.extractMetadataAndCover(ctx, destFilePath, trackID, fileInput.Filename)

		var plIDPtr *string
		if playlistID != "" {
			plIDPtr = &playlistID
		}

		now := time.Now().UTC()
		track := &domain.Track{
			ID:           trackID,
			YouTubeID:    "local_" + trackID,
			PlaylistID:   plIDPtr,
			UserID:       userID,
			Title:        meta.Title,
			Artist:       meta.Artist,
			Album:        meta.Album,
			Duration:     meta.Duration,
			FilePath:     destFilePath,
			CoverPath:    meta.CoverPath,
			FileSize:     written,
			Format:       strings.TrimPrefix(ext, "."),
			Bitrate:      meta.Bitrate,
			Status:       domain.TrackStatusReady,
			DownloadedAt: &now,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := u.trackRepo.Create(track); err != nil {
			_ = os.Remove(destFilePath)
			if meta.CoverPath != "" {
				_ = os.Remove(meta.CoverPath)
			}
			result.Errors = append(result.Errors, fmt.Sprintf("%s: ошибка сохранения в базу данных: %v", fileInput.Filename, err))
			continue
		}

		if u.eventHub != nil {
			u.eventHub.BroadcastUser(userID, worker.EventMessage{
				Type: worker.EventTypeTrack,
				Data: track,
			})
		}

		result.Uploaded = append(result.Uploaded, track)
	}

	return result, nil
}

type extractedMeta struct {
	Title     string
	Artist    string
	Album     string
	Duration  int
	Bitrate   int
	CoverPath string
}

func (u *TrackUsecase) extractMetadataAndCover(ctx context.Context, filePath, trackID, originalFilename string) extractedMeta {
	meta := extractedMeta{
		Title:    "",
		Artist:   "",
		Album:    "",
		Duration: 0,
		Bitrate:  320,
	}

	// 1. Try ffprobe for audio stream and format metadata
	ffprobeBin := "ffprobe"
	if u.ffmpegPath != "" {
		probeCandidate := filepath.Join(filepath.Dir(u.ffmpegPath), "ffprobe")
		if _, err := os.Stat(probeCandidate); err == nil {
			ffprobeBin = probeCandidate
		}
	}

	probeCmd := exec.CommandContext(ctx, ffprobeBin,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)
	var stdout bytes.Buffer
	probeCmd.Stdout = &stdout

	if err := probeCmd.Run(); err == nil {
		var probeData struct {
			Streams []struct {
				CodecName string `json:"codec_name"`
				BitRate   string `json:"bit_rate"`
				Duration  string `json:"duration"`
			} `json:"streams"`
			Format struct {
				Duration string            `json:"duration"`
				BitRate  string            `json:"bit_rate"`
				Tags     map[string]string `json:"tags"`
			} `json:"format"`
		}

		if err := json.Unmarshal(stdout.Bytes(), &probeData); err == nil {
			// Duration
			if probeData.Format.Duration != "" {
				if d, err := strconv.ParseFloat(probeData.Format.Duration, 64); err == nil && d > 0 {
					meta.Duration = int(d)
				}
			}
			if meta.Duration == 0 {
				for _, s := range probeData.Streams {
					if s.Duration != "" {
						if d, err := strconv.ParseFloat(s.Duration, 64); err == nil && d > 0 {
							meta.Duration = int(d)
							break
						}
					}
				}
			}

			// Bitrate
			if probeData.Format.BitRate != "" {
				if br, err := strconv.ParseInt(probeData.Format.BitRate, 10, 64); err == nil && br > 0 {
					meta.Bitrate = int(br / 1000)
				}
			}

			// Tags (case-insensitive lookup)
			getTag := func(keys ...string) string {
				for _, k := range keys {
					for tagK, tagV := range probeData.Format.Tags {
						if strings.EqualFold(tagK, k) && strings.TrimSpace(tagV) != "" {
							return strings.TrimSpace(tagV)
						}
					}
				}
				return ""
			}

			meta.Title = getTag("title", "track_name", "song_name")
			meta.Artist = getTag("artist", "album_artist", "composer", "author", "performer")
			meta.Album = getTag("album", "album_name", "collection")
		}
	}

	// 2. Fallback to clean filename if Title or Artist are missing
	cleanBase := strings.TrimSuffix(originalFilename, filepath.Ext(originalFilename))
	if meta.Title == "" {
		if strings.Contains(cleanBase, " - ") {
			parts := strings.SplitN(cleanBase, " - ", 2)
			if meta.Artist == "" {
				meta.Artist = strings.TrimSpace(parts[0])
			}
			meta.Title = strings.TrimSpace(parts[1])
		} else {
			meta.Title = cleanBase
		}
	}
	if meta.Artist == "" {
		meta.Artist = "Unknown Artist"
	}
	if meta.Title == "" {
		meta.Title = "Uploaded Track"
	}

	// 3. Extract embedded cover art with ffmpeg
	ffmpegBin := u.ffmpegPath
	if ffmpegBin == "" {
		ffmpegBin = "ffmpeg"
	}

	coverDest := filepath.Join(u.coversDir, trackID+".jpg")
	coverCmd := exec.CommandContext(ctx, ffmpegBin,
		"-y",
		"-i", filePath,
		"-an",
		"-vcodec", "copy",
		coverDest,
	)
	_ = coverCmd.Run()

	if fi, err := os.Stat(coverDest); err == nil && fi.Size() >= 512 {
		meta.CoverPath = coverDest
	} else {
		_ = os.Remove(coverDest)
		// Fallback attempt: transcode thumbnail to JPEG
		coverCmd2 := exec.CommandContext(ctx, ffmpegBin,
			"-y",
			"-i", filePath,
			"-an",
			"-vcodec", "mjpeg",
			"-vf", "scale='min(500,iw)':-1",
			coverDest,
		)
		_ = coverCmd2.Run()
		if fi2, err := os.Stat(coverDest); err == nil && fi2.Size() >= 512 {
			meta.CoverPath = coverDest
		} else {
			_ = os.Remove(coverDest)
		}
	}

	return meta
}
