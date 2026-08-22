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
	track, err := h.trackUsecase.GetByID(id)
	if err != nil || track == nil || track.FilePath == "" {
		http.Error(w, "track not found or not ready", http.StatusNotFound)
		return
	}

	file, err := os.Open(track.FilePath)
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

	// Set audio Content-Type based on format
	contentType := "audio/ogg"
	switch track.Format {
	case "opus":
		contentType = "audio/ogg; codecs=opus"
	case "m4a", "aac":
		contentType = "audio/mp4"
	case "mp3":
		contentType = "audio/mpeg"
	case "flac":
		contentType = "audio/flac"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	http.ServeContent(w, r, filepath.Base(track.FilePath), stat.ModTime(), file)
}

// ServeCover serves track cover art (JPEG)
func (h *StreamHandler) ServeCover(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	track, err := h.trackUsecase.GetByID(id)
	if err != nil || track == nil || track.CoverPath == "" {
		http.Error(w, "cover not found", http.StatusNotFound)
		return
	}

	file, err := os.Open(track.CoverPath)
	if err != nil {
		http.Error(w, "cover file not found on disk", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "failed to stat cover file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	http.ServeContent(w, r, "cover.jpg", stat.ModTime(), file)
}

// DownloadAudio forces browser download with Content-Disposition
func (h *StreamHandler) DownloadAudio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	track, err := h.trackUsecase.GetByID(id)
	if err != nil || track == nil || track.FilePath == "" {
		http.Error(w, "track not found", http.StatusNotFound)
		return
	}

	file, err := os.Open(track.FilePath)
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

	filename := fmt.Sprintf("%s - %s.%s", track.Artist, track.Title, track.Format)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	http.ServeContent(w, r, filename, stat.ModTime(), file)
}
