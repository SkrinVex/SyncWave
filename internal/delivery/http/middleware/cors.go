package middleware

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/cors"
)

func NewCORSMiddleware() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Range"},
		ExposedHeaders:   []string{"Link", "Content-Range", "Content-Length", "Accept-Ranges"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (rw *responseWriter) ReadFrom(r io.Reader) (int64, error) {
	if rf, ok := rw.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}
	return io.Copy(rw.ResponseWriter, r)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, errors.New("hijacker not supported")
}

func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		// Don't clutter logs with SSE heartbeat or repeated cover requests
		if r.URL.Path != "/api/v1/sync/events" {
			log.Printf("[%s] %s %s -> %d (%s)", r.Method, r.RequestURI, r.RemoteAddr, wrapped.statusCode, time.Since(start))
		}
	})
}
