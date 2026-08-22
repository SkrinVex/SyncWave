package http

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/syncwave/syncwave/internal/delivery/http/handler"
	"github.com/syncwave/syncwave/internal/delivery/http/middleware"
)

type RouterConfig struct {
	AuthHandler     *handler.AuthHandler
	TrackHandler    *handler.TrackHandler
	PlaylistHandler *handler.PlaylistHandler
	SyncHandler     *handler.SyncHandler
	SettingsHandler *handler.SettingsHandler
	StreamHandler   *handler.StreamHandler
	AuthMiddleware  *middleware.AuthMiddleware
	EmbedFS         fs.FS
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.NewCORSMiddleware())

	// API Subrouter
	r.Route("/api/v1", func(api chi.Router) {
		// Public Auth routes
		api.Route("/auth", func(auth chi.Router) {
			auth.Get("/status", cfg.AuthHandler.GetStatus)
			auth.Post("/setup", cfg.AuthHandler.SetupAdmin)
			auth.Post("/login", cfg.AuthHandler.Login)
		})

		// Protected routes
		api.Group(func(protected chi.Router) {
			protected.Use(cfg.AuthMiddleware.RequireAuth)

			// Current user profile
			protected.Get("/auth/me", cfg.AuthHandler.GetMe)

			// Tracks & Streaming
			protected.Route("/tracks", func(tracks chi.Router) {
				tracks.Get("/", cfg.TrackHandler.List)
				tracks.Get("/stats", cfg.TrackHandler.GetStats)
				tracks.Get("/ready", cfg.TrackHandler.GetAllReady)
				tracks.Get("/{id}", cfg.TrackHandler.GetByID)
				tracks.Delete("/{id}", cfg.TrackHandler.Delete)
				tracks.Get("/{id}/stream", cfg.StreamHandler.StreamAudio)
				tracks.Get("/{id}/cover", cfg.StreamHandler.ServeCover)
				tracks.Get("/{id}/download", cfg.StreamHandler.DownloadTrackFile)
			})

			// Playlists
			protected.Route("/playlists", func(pl chi.Router) {
				pl.Get("/", cfg.PlaylistHandler.List)
				pl.Post("/", cfg.PlaylistHandler.Create)
				pl.Get("/{id}", cfg.PlaylistHandler.GetByID)
				pl.Put("/{id}", cfg.PlaylistHandler.Update)
				pl.Delete("/{id}", cfg.PlaylistHandler.Delete)
				pl.Post("/{id}/sync", cfg.PlaylistHandler.TriggerSync)
			})

			// Sync Operations & Live Events
			protected.Route("/sync", func(sync chi.Router) {
				sync.Post("/trigger", cfg.SyncHandler.TriggerAll)
				sync.Get("/progress", cfg.SyncHandler.GetProgress)
				sync.Get("/logs", cfg.SyncHandler.GetLogs)
				sync.Delete("/logs", cfg.SyncHandler.ClearLogs)
				sync.Get("/events", cfg.SyncHandler.StreamEvents)
			})

			// Settings & Maintenance
			protected.Route("/settings", func(settings chi.Router) {
				settings.Get("/", cfg.SettingsHandler.GetSettings)
				settings.Put("/", cfg.SettingsHandler.UpdateSettings)
				settings.Post("/cookies", cfg.SettingsHandler.UploadCookies)
				settings.Delete("/cookies", cfg.SettingsHandler.DeleteCookies)
				settings.Post("/test-proxy", cfg.SettingsHandler.TestProxy)
			})
		})
	})

	// Static Web SPA handler
	staticHandler := handler.NewStaticHandler(cfg.EmbedFS)
	r.NotFound(staticHandler.ServeHTTP)

	return r
}
