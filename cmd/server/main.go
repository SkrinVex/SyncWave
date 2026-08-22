package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/syncwave/syncwave/internal/config"
	deliveryhttp "github.com/syncwave/syncwave/internal/delivery/http"
	"github.com/syncwave/syncwave/internal/delivery/http/handler"
	"github.com/syncwave/syncwave/internal/delivery/http/middleware"
	"github.com/syncwave/syncwave/internal/infrastructure/auth"
	"github.com/syncwave/syncwave/internal/infrastructure/worker"
	"github.com/syncwave/syncwave/internal/infrastructure/ytdlp"
	"github.com/syncwave/syncwave/internal/repository/sqlite"
	"github.com/syncwave/syncwave/internal/usecase"
	"github.com/syncwave/syncwave/web"
)

func main() {
	log.Println("==================================================")
	log.Println("           SyncWave Music Daemon v1.0.0           ")
	log.Println("==================================================")

	cfg := config.Load()

	// 1. Database & Repositories
	db, err := sqlite.NewDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}
	defer db.Close()
	log.Printf("[DB] SQLite initialized at %s (WAL mode enabled)", cfg.DBPath)

	userRepo := sqlite.NewUserRepository(db)
	trackRepo := sqlite.NewTrackRepository(db)
	playlistRepo := sqlite.NewPlaylistRepository(db)
	settingsRepo := sqlite.NewSettingsRepository(db)
	logRepo := sqlite.NewSyncLogRepository(db)

	// 2. Infrastructure Services
	hasher := auth.NewPasswordHasher()
	jwtService := auth.NewJWTService(cfg.JWTSecret)
	ytdlpClient := ytdlp.NewClient(cfg.YtDlpPath, cfg.FFmpegPath, cfg.CookiesPath, cfg.MusicDir, cfg.CoversDir)
	eventHub := worker.NewEventHub()

	// 3. Worker & Scheduler
	workerQueue := worker.NewWorkerQueue(ytdlpClient, trackRepo, playlistRepo, logRepo, eventHub, 50)
	workerQueue.Start()
	defer workerQueue.Stop()
	log.Println("[Worker] Background sync queue worker active")

	scheduler := worker.NewScheduler(playlistRepo, workerQueue)
	scheduler.Start()
	defer scheduler.Stop()
	log.Println("[Scheduler] Cron auto-sync scheduler active")

	// 4. Usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, hasher, jwtService)
	trackUsecase := usecase.NewTrackUsecase(trackRepo)
	playlistUsecase := usecase.NewPlaylistUsecase(playlistRepo, ytdlpClient, workerQueue)
	syncUsecase := usecase.NewSyncUsecase(playlistRepo, logRepo, workerQueue)
	settingsUsecase := usecase.NewSettingsUsecase(settingsRepo, trackRepo, playlistRepo, ytdlpClient, cfg.DataDir, cfg.DBPath)
	settingsUsecase.InitFromDB()

	// 5. Delivery Handlers & Middlewares
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	authHandler := handler.NewAuthHandler(authUsecase)
	trackHandler := handler.NewTrackHandler(trackUsecase)
	playlistHandler := handler.NewPlaylistHandler(playlistUsecase, syncUsecase)
	syncHandler := handler.NewSyncHandler(syncUsecase, eventHub)
	settingsHandler := handler.NewSettingsHandler(settingsUsecase)
	streamHandler := handler.NewStreamHandler(trackUsecase)

	// 6. Router Setup
	httpRouter := deliveryhttp.NewRouter(deliveryhttp.RouterConfig{
		AuthHandler:     authHandler,
		TrackHandler:    trackHandler,
		PlaylistHandler: playlistHandler,
		SyncHandler:     syncHandler,
		SettingsHandler: settingsHandler,
		StreamHandler:   streamHandler,
		AuthMiddleware:  authMiddleware,
		EmbedFS:         web.DistFS,
	})

	serverAddr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      httpRouter,
		ReadTimeout:  30 * time.Minute, // Large timeout to support long downloads / audio streaming
		WriteTimeout: 30 * time.Minute,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Graceful Shutdown Listener
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[Server] SyncWave daemon listening on http://%s", serverAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Fatal: HTTP server failed: %v", err)
		}
	}()

	<-stopChan
	log.Println("[Server] Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[Server] Shutdown error: %v", err)
	}

	log.Println("[Server] SyncWave daemon terminated cleanly")
}
