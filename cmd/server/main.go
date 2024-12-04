package main

import (
	"log/slog"
	"net/http"
	"os"
	"song-library/internal/http-server/handlers/songs/add"
	"song-library/internal/http-server/handlers/songs/delete"
	"song-library/internal/http-server/handlers/songs/update"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"song-library/internal/config"
	"song-library/internal/http-server/handlers/info"
	getSongs "song-library/internal/http-server/handlers/songs/get"
	getLyrics "song-library/internal/http-server/handlers/songs/lyrics/get"
	mwLogger "song-library/internal/http-server/middleware/logger"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/lib/logger/slogpretty"
	"song-library/internal/storage/postgres"
)

const (
	envLocal = "local"
	envProd  = "prod"
	envDev   = "dev"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting song-library", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	storage, err := postgres.New(cfg.ConnectionString)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	log.Info("storage connected")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log)) // TODO: fix pretty middleware logger
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/songs", func(r chi.Router) {
		r.Get("/", getSongs.New(log, storage))        // Список песен с фильтрацией и пагинацией
		r.Get("/lyrics", getLyrics.New(log, storage)) // Текст песни с пагинацией
		r.Delete("/", delete.New(log, storage))       // Удаление песни
		r.Put("/", update.New(log, storage))          // Изменение данных песни
		r.Post("/", add.New(log, storage))            // Добавление новой песни
	})

	router.Get("/info", info.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)

	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}
	handler := slogpretty.NewPrettyHandler(os.Stdout, opts)
	return slog.New(handler)
}
