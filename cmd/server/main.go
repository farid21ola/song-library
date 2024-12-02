package main

import (
	"log/slog"
	"os"
	"song-library/internal/http-server/handlers/info"
	"song-library/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"song-library/internal/config"
	mwLogger "song-library/internal/http-server/middleware/logger"
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

	_ = storage

	log.Info("storage connected")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/songs", func(r chi.Router) {
		r.Get("/", handlers.GetSongs)                         // Получение списка песен с фильтрацией и пагинацией
		r.Get("/{artist}/{title}/lyrics", handlers.GetLyrics) // Получение текста песни с пагинацией по куплетам
		r.Delete("/{artist}/{title}", handlers.DeleteSong)    // Удаление песни
		r.Put("/{artist}/{title}", handlers.UpdateSong)       // Изменение данных песни
		r.Post("/", handlers.AddSong)                         // Добавление новой песни
	})

	router.Get("/info", info.New(log, storage))

	//ToDo: server
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
