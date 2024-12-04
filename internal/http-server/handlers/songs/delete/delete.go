package delete

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
)

type Request struct {
	Artist string `json:"group" validate:"required"`
	Title  string `json:"song" validate:"required"`
}

type Response struct {
	resp.Response
}

type SongRemover interface {
	DeleteSong(ctx context.Context, artist, title string) error
}

func New(log *slog.Logger, songRemover SongRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.songs.delete.New"

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

		err = songRemover.DeleteSong(r.Context(), req.Artist, req.Title)
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
			log.Info("failed to delete song",
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
