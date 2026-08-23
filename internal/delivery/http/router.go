package http

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/syncwave/syncwave/internal/delivery/http/handler"
	"github.com/syncwave/syncwave/internal/delivery/http/middleware"
)

type RouterConfig struct {
	AuthHandler      *handler.AuthHandler
	AdminHandler     *handler.AdminHandler
	TrackHandler     *handler.TrackHandler
	PlaylistHandler  *handler.PlaylistHandler
	SyncHandler      *handler.SyncHandler
	SettingsHandler  *handler.SettingsHandler
	StreamHandler    *handler.StreamHandler
	BlacklistHandler *handler.BlacklistHandler
	AuthMiddleware   *middleware.AuthMiddleware
	EmbedFS          fs.FS
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.NewCORSMiddleware())

	// Create Rate Limiters
	authLimiter := middleware.NewIPRateLimiter(20, 1*time.Minute)      // 20 requests / min for auth (bruteforce protection)
	generalLimiter := middleware.NewIPRateLimiter(1000, 1*time.Minute) // 1000 requests / min for standard API (generous headroom)

	// API Subrouter
	r.Route("/api/v1", func(api chi.Router) {
		// Public Auth routes (with brute-force protection)
		api.Route("/auth", func(auth chi.Router) {
			auth.Use(generalLimiter.Limit)
			auth.Get("/status", cfg.AuthHandler.Status)
			auth.With(authLimiter.Limit).Post("/setup", cfg.AuthHandler.Setup)
			auth.With(authLimiter.Limit).Post("/login", cfg.AuthHandler.Login)
			auth.With(authLimiter.Limit).Post("/register", cfg.AuthHandler.Register)
		})

		// Protected routes
		api.Group(func(protected chi.Router) {
			protected.Use(cfg.AuthMiddleware.RequireAuth)

			// Fast unthrottled media & stream routes (HTML5 range requests and covers)
			protected.Get("/tracks/{id}/stream", cfg.StreamHandler.StreamAudio)
			protected.Get("/tracks/{id}/cover", cfg.StreamHandler.ServeCover)
			protected.Get("/tracks/{id}/download", cfg.StreamHandler.DownloadAudio)
			protected.Get("/sync/events", cfg.SyncHandler.StreamEvents)

			// Standard API routes (with general rate limiter)
			protected.Group(func(rateLimited chi.Router) {
				rateLimited.Use(generalLimiter.Limit)

				// Current user profile
				rateLimited.Get("/auth/me", cfg.AuthHandler.Me)

				// Admin operations (Require admin role)
				if cfg.AdminHandler != nil {
					rateLimited.Group(func(admin chi.Router) {
						admin.Use(cfg.AuthMiddleware.RequireAdmin)
						admin.Route("/admin", func(adm chi.Router) {
							adm.Get("/users", cfg.AdminHandler.ListUsers)
							adm.Put("/users/{id}/quota", cfg.AdminHandler.UpdateUserQuota)
							adm.Delete("/users/{id}", cfg.AdminHandler.DeleteUser)
							adm.Post("/registration", cfg.AdminHandler.SetRegistration)
							adm.Post("/global-limit", cfg.AdminHandler.SetGlobalLimit)
						})
					})
				}

				// Tracks
				rateLimited.Route("/tracks", func(tracks chi.Router) {
					tracks.Get("/", cfg.TrackHandler.List)
					tracks.Get("/stats", cfg.TrackHandler.GetStats)
					tracks.Get("/ready", cfg.TrackHandler.GetAllReady)
					tracks.Post("/upload", cfg.TrackHandler.Upload)
					tracks.Post("/batch-delete", cfg.TrackHandler.BatchDelete)
					tracks.Get("/{id}", cfg.TrackHandler.GetByID)
					tracks.Delete("/{id}", cfg.TrackHandler.Delete)
				})

				// Playlists
				rateLimited.Route("/playlists", func(pl chi.Router) {
					pl.Get("/", cfg.PlaylistHandler.List)
					pl.Post("/", cfg.PlaylistHandler.Create)
					pl.Get("/{id}", cfg.PlaylistHandler.Get)
					pl.Put("/{id}", cfg.PlaylistHandler.Update)
					pl.Delete("/{id}", cfg.PlaylistHandler.Delete)
					pl.Post("/{id}/sync", cfg.PlaylistHandler.Sync)
				})

				// Sync Operations
				rateLimited.Route("/sync", func(sync chi.Router) {
					sync.Post("/trigger", cfg.SyncHandler.TriggerAll)
					sync.Post("/cancel", cfg.SyncHandler.Cancel)
					sync.Get("/progress", cfg.SyncHandler.GetProgress)
					sync.Get("/logs", cfg.SyncHandler.GetLogs)
					sync.Delete("/logs", cfg.SyncHandler.ClearLogs)
				})

				// Settings & Maintenance
				rateLimited.Route("/settings", func(settings chi.Router) {
					settings.Get("/", cfg.SettingsHandler.Get)
					settings.Put("/", cfg.SettingsHandler.Update)
					settings.Post("/cookies", cfg.SettingsHandler.UploadCookies)
					settings.Delete("/cookies", cfg.SettingsHandler.DeleteCookies)
					settings.Post("/test-proxy", cfg.SettingsHandler.TestProxy)
				})

				// Blacklist
				rateLimited.Route("/blacklist", func(bl chi.Router) {
					bl.Get("/", cfg.BlacklistHandler.List)
					bl.Delete("/{id}", cfg.BlacklistHandler.Delete)
				})
			})
		})
	})

	// Static Web SPA handler
	staticHandler := handler.NewStaticHandler(cfg.EmbedFS)
	r.NotFound(staticHandler.ServeHTTP)

	return r
}
