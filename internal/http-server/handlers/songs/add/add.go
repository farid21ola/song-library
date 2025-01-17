package add

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"song-library/internal/lib/api/resp"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/models"
	"song-library/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type SongSaver interface {
	AddSong(ctx context.Context, song *storage.Song) error
}

// @Summary Add song
// @Description The request contains details about the song, including the artist's name, song title, release date, lyrics, and a link.
// @Accept  json
// @Tags songs
// @Produce  json
// @Param   request  body models.Song true "Song info"
// @Success 200 {object} resp.Response  "Song successfully added"
// @Failure 400 {object} resp.Response  "Bad request"
// @Failure 409 {object} resp.Response  "Song already exists"
// @Failure 500 {object} resp.Response  "Internal server error"
// @Router /songs [post]
func New(log *slog.Logger, songSaver SongSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.songs.add.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req models.Song

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Debug("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		song := storage.Song{
			Artist: req.Artist,
			Title:  req.Title,
			Lyrics: req.Text,
			Link:   req.Link,
		}

		if req.ReleaseDate != "" {
			releaseDate, err := time.Parse("02.01.2006", req.ReleaseDate)
			if err != nil {
				log.Error("failed to parse song release date", sl.Err(err))

				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error("bad release date"))

				return
			}
			song.ReleaseDate = releaseDate
		}

		err = songSaver.AddSong(r.Context(), &song)
		if errors.Is(err, storage.ErrSongExists) {
			log.Error("song already exists", sl.Err(err))

			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, resp.Error("song already exists"))

			return
		}
		if err != nil {
			log.Error("failed to add song", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Debug("song added",
			slog.String("artist", req.Artist),
			slog.String("title", req.Title),
		)

		render.JSON(w, r, resp.OK())
	}
}
