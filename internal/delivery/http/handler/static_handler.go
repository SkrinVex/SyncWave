package handler

import (
	"io/fs"
	"net/http"
	"strings"
)

type StaticHandler struct {
	fileSystem http.FileSystem
}

func NewStaticHandler(embedFS fs.FS) *StaticHandler {
	distFS, err := fs.Sub(embedFS, "dist")
	if err != nil {
		distFS = embedFS
	}
	return &StaticHandler{
		fileSystem: http.FS(distFS),
	}
}

func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// Check if the requested file exists in embedded FS
	f, err := h.fileSystem.Open(path)
	if err != nil {
		// Fallback to index.html for SPA client-side routing
		indexFile, err := h.fileSystem.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer indexFile.Close()

		stat, _ := indexFile.Stat()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile)
		return
	}
	defer f.Close()

	stat, _ := f.Stat()
	if stat.IsDir() {
		// Serve index.html for directories
		indexFile, err := h.fileSystem.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer indexFile.Close()

		istat, _ := indexFile.Stat()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "index.html", istat.ModTime(), indexFile)
		return
	}

	// Cache static assets aggressively
	if strings.HasPrefix(path, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}

	http.FileServer(h.fileSystem).ServeHTTP(w, r)
}
