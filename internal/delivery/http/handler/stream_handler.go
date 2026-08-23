package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

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
			matches, _ := filepath.Glob(filepath.Join(dir, fmt.Sprintf("%s.*", baseName)))
			for _, m := range matches {
				if ext := filepath.Ext(m); ext != ".jpg" && ext != ".webp" && ext != ".png" {
					if altFile, altErr := os.Open(m); altErr == nil {
						file = altFile
						filePath = m
						err = nil
						break
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
	case ".opus", ".ogg":
		contentType = "audio/ogg; codecs=opus"
	case ".m4a", ".mp4", ".aac":
		contentType = "audio/mp4"
	case ".mp3":
		contentType = "audio/mpeg"
	case ".flac":
		contentType = "audio/flac"
	case ".webm":
		contentType = "audio/webm"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "public, max-age=3600")

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
		http.Error(w, "audio file not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "failed to stat file", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("%s - %s%s", track.Artist, track.Title, filepath.Ext(filePath))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	http.ServeContent(w, r, filename, stat.ModTime(), file)
}
