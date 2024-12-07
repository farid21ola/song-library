package update

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

type SongUpdater interface {
	UpdateSong(ctx context.Context, song *storage.Song) error
}

// @Summary Update song details by artist and title.
// @Description Updates the details of a song by artist and title. Only the fields that are provided in the request body will be updated. Fields like lyrics, release date, and link are optional.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param request body models.Song true "New song info "
// @Success 200 {object} resp.Response "Song successfully updated"
// @Failure 400 {object} resp.Response "Bad Request"
// @Failure 404 {object} resp.Response "Not Found - Song not found"
// @Failure 500 {object} resp.Response "Internal Server Error"
// @Router /songs [put]
func New(log *slog.Logger, songUpdater SongUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.songs.update.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req models.Song

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("bad request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		if req.Text == "" && req.Link == "" && req.ReleaseDate == "" {
			log.Error("nothing to change")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("nothing to change"))
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

				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("internal error"))

				return
			}
			song.ReleaseDate = releaseDate
		}

		err = songUpdater.UpdateSong(r.Context(), &song)
		if errors.Is(err, storage.ErrSongNotFound) {
			log.Error("song not found", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, resp.Error("song not found"))

			return
		}
		if err != nil {
			log.Error("failed to update song", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("song updated",
			slog.String("artist", req.Artist),
			slog.String("title", req.Title),
		)

		render.JSON(w, r, resp.OK())
	}
}
