package update

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

type SongUpdater interface {
	UpdateSong(ctx context.Context, song *storage.Song) error
}

func New(log *slog.Logger, songUpdater SongUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.songs.update.New"

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

		if req.Text == "" && req.Link == "" && req.ReleaseDate == "" {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("nothing to change"))
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

		err = songUpdater.UpdateSong(r.Context(), &song)
		if errors.Is(err, storage.ErrSongNotFound) {
			log.Info("song not found",
				slog.String("artist", req.Artist),
				slog.String("title", req.Title),
			)
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, resp.Error("song not found"))

			return
		}
		if err != nil {
			log.Info("failed to update song",
				slog.String("artist", req.Artist),
				slog.String("title", req.Title),
			)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("song updated",
			slog.String("artist", req.Artist),
			slog.String("title", req.Title),
		)

		render.JSON(w, r, Response{
			Response: resp.OK(),
		})
	}
}
