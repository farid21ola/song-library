package add

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "song-library/internal/lib/api/response"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/storage"
	"time"
)

type Request struct {
	Artist      string `json:"group" validate:"required"`
	Title       string `json:"song" validate:"required"`
	ReleaseDate string `json:"release_date,omitempty"`
	Text        string `json:"text,omitempty"`
	Link        string `json:"link,omitempty"`
}

type Response struct {
	resp.Response
}

type SongSaver interface {
	AddSong(ctx context.Context, song *storage.Song) error
}

func New(log *slog.Logger, songSaver SongSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.songs.add.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if req.Artist == "" || req.Title == "" {
			log.Error("missing required parameters",
				slog.String("artist", req.Artist),
				slog.String("title", req.Title),
			)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("missing required parameters: group and song"))
			return
		}

		releaseDate, err := time.Parse("02.01.2006", req.ReleaseDate)
		if err != nil {
			log.Info("failed to parse song release date",
				slog.String("artist", req.Artist),
				slog.String("title", req.Title),
				slog.String("release date", req.ReleaseDate),
			)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		song := storage.Song{
			Artist:      req.Artist,
			Title:       req.Title,
			ReleaseDate: releaseDate,
			Lyrics:      req.Text,
			Link:        req.Link,
		}

		err = songSaver.AddSong(r.Context(), &song)
		if errors.Is(err, storage.ErrSongExists) {
			log.Info("song already exists",
				slog.String("artist", req.Artist),
				slog.String("title", req.Title),
			)

			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, resp.Error("song already exists"))

			return
		}
		if err != nil {
			log.Info("failed to add song",
				slog.String("artist", req.Artist),
				slog.String("title", req.Title),
			)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("song deleted",
			slog.String("artist", req.Artist),
			slog.String("title", req.Title),
		)

		render.JSON(w, r, Response{
			Response: resp.OK(),
		})
	}
}
