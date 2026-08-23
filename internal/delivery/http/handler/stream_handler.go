package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/syncwave/syncwave/internal/usecase"
)

type StreamHandler struct {
	trackUsecase *usecase.TrackUsecase
}

func NewStreamHandler(trackUsecase *usecase.TrackUsecase) *StreamHandler {
	return &StreamHandler{trackUsecase: trackUsecase}
}

// StreamAudio serves audio files with full RFC 7233 Range support (HTTP 206 Partial Content)
func (h *StreamHandler) StreamAudio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, _ := r.Context().Value("user_id").(string)
	track, err := h.trackUsecase.GetByID(id, userID)
	if err != nil || track == nil || track.FilePath == "" {
		http.Error(w, "track not found or not ready", http.StatusNotFound)
		return
	}

	// Try primary path, or look for alternative audio formats on disk with same ID
	filePath := track.FilePath
	file, err := os.Open(filePath)
	if err != nil {
		dir := filepath.Dir(filePath)
		baseName := track.YouTubeID
		if baseName != "" {
			validExts := map[string]bool{
				".opus": true, ".m4a": true, ".mp3": true,
				".flac": true, ".webm": true, ".ogg": true,
				".aac": true, ".mp4": true, ".wav": true,
			}
			matches, _ := filepath.Glob(filepath.Join(dir, fmt.Sprintf("%s.*", baseName)))
			for _, m := range matches {
				ext := filepath.Ext(m)
				if validExts[ext] {
					if altFile, altErr := os.Open(m); altErr == nil {
						if fi, fiErr := altFile.Stat(); fiErr == nil && fi.Size() > 1024 {
							file = altFile
							filePath = m
							err = nil
							break
						} else {
							altFile.Close()
						}
					}
				}
			}
		}
	}

	if err != nil {
		http.Error(w, "audio file not found on disk", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "failed to stat audio file", http.StatusInternalServerError)
		return
	}

	// Set audio Content-Type based on actual file extension
	ext := filepath.Ext(filePath)
	contentType := "audio/ogg"
	switch ext {
	case ".opus":
		contentType = "audio/ogg; codecs=opus"
	case ".ogg":
		contentType = "audio/ogg"
	case ".m4a", ".mp4", ".aac":
		contentType = "audio/mp4"
	case ".mp3":
		contentType = "audio/mpeg"
	case ".flac":
		contentType = "audio/flac"
	case ".webm":
		contentType = "audio/webm"
	case ".wav":
		contentType = "audio/wav"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	http.ServeContent(w, r, filepath.Base(filePath), stat.ModTime(), file)
}

// ServeCover serves track cover art (JPEG) or redirects to YouTube CDN fallback
func (h *StreamHandler) ServeCover(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, _ := r.Context().Value("user_id").(string)
	track, err := h.trackUsecase.GetByID(id, userID)
	if err != nil || track == nil {
		http.Error(w, "cover not found", http.StatusNotFound)
		return
	}

	if track.CoverPath != "" {
		if file, err := os.Open(track.CoverPath); err == nil {
			defer file.Close()
			if stat, err := file.Stat(); err == nil && stat.Size() > 0 {
				w.Header().Set("Content-Type", "image/jpeg")
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				http.ServeContent(w, r, "cover.jpg", stat.ModTime(), file)
				return
			}
		}
	}

	// Fallback redirect to official YouTube thumbnail CDN
	if track.YouTubeID != "" {
		http.Redirect(w, r, fmt.Sprintf("https://i.ytimg.com/vi/%s/hqdefault.jpg", track.YouTubeID), http.StatusTemporaryRedirect)
		return
	}

	http.Error(w, "cover not found", http.StatusNotFound)
}

// DownloadAudio forces browser download with Content-Disposition
func (h *StreamHandler) DownloadAudio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, _ := r.Context().Value("user_id").(string)
	track, err := h.trackUsecase.GetByID(id, userID)
	if err != nil || track == nil || track.FilePath == "" {
		http.Error(w, "track not found", http.StatusNotFound)
		return
	}

	filePath := track.FilePath
	file, err := os.Open(filePath)
	if err != nil {
		dir := filepath.Dir(filePath)
		baseName := track.YouTubeID
		if baseName != "" {
			validExts := map[string]bool{
				".opus": true, ".m4a": true, ".mp3": true,
				".flac": true, ".webm": true, ".ogg": true,
				".aac": true, ".mp4": true, ".wav": true,
			}
			matches, _ := filepath.Glob(filepath.Join(dir, fmt.Sprintf("%s.*", baseName)))
			for _, m := range matches {
				ext := filepath.Ext(m)
				if validExts[ext] {
					if altFile, altErr := os.Open(m); altErr == nil {
						if fi, fiErr := altFile.Stat(); fiErr == nil && fi.Size() > 1024 {
							file = altFile
							filePath = m
							err = nil
							break
						} else {
							altFile.Close()
						}
					}
				}
			}
		}
	}

	if err != nil {
		http.Error(w, "audio file not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "failed to stat file", http.StatusInternalServerError)
		return
	}

	ext := filepath.Ext(filePath)
	contentType := "audio/ogg"
	switch ext {
	case ".opus":
		contentType = "audio/ogg; codecs=opus"
	case ".ogg":
		contentType = "audio/ogg"
	case ".m4a", ".mp4", ".aac":
		contentType = "audio/mp4"
	case ".mp3":
		contentType = "audio/mpeg"
	case ".flac":
		contentType = "audio/flac"
	case ".webm":
		contentType = "audio/webm"
	case ".wav":
		contentType = "audio/wav"
	}

	artist := track.Artist
	if artist == "" {
		artist = "Unknown Artist"
	}
	title := track.Title
	if title == "" {
		title = track.YouTubeID
	}

	rawFilename := fmt.Sprintf("%s - %s%s", artist, title, ext)
	// Remove filesystem reserved characters
	rawFilename = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, rawFilename)

	// ASCII fallback: strip non-ASCII for legacy headers
	asciiFallback := strings.Map(func(r rune) rune {
		if r > 127 || r < 32 || r == '"' || r == '\\' {
			return '_'
		}
		return r
	}, rawFilename)
	if strings.TrimSpace(strings.ReplaceAll(asciiFallback, "_", "")) == "" {
		asciiFallback = fmt.Sprintf("track_%s%s", track.YouTubeID, ext)
	}

	utf8Filename := url.PathEscape(rawFilename)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", asciiFallback, utf8Filename))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
	w.Header().Set("Accept-Ranges", "bytes")

	http.ServeContent(w, r, rawFilename, stat.ModTime(), file)
}
