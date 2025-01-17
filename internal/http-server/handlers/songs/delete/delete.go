package delete

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"song-library/internal/lib/api/resp"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type SongRemover interface {
	DeleteSong(ctx context.Context, artist, title string) error
}

// @Summary Delete a song by artist and title.
// @Description Deletes the specified song by artist and title from the database. Requires both "group" and "song" path parameters.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param group path string true "Artist Name" Example("The Beatles")
// @Param song path string true "Song Title" Example("Hey Jude")
// @Success 200 {object} resp.Response "Song successfully deleted"
// @Failure 400 {object} resp.Response "Bad Request"
// @Failure 404 {object} resp.Response "Not Found - Song not found"
// @Failure 500 {object} resp.Response "Internal Server Error"
// @Router /songs/{group}/{song} [delete]
func New(log *slog.Logger, songRemover SongRemover) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.songs.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		group := chi.URLParam(r, "group")
		song := chi.URLParam(r, "song")

		if group == "" || song == "" {
			log.Error("missing required path parameters")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("missing required path parameters"))
			return
		}

		err := songRemover.DeleteSong(r.Context(), group, song)
		if errors.Is(err, storage.ErrSongNotFound) {
			log.Error("song not found", sl.Err(err))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, resp.Error("song not found"))

			return
		}
		if err != nil {
			log.Error("failed to delete song", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Debug("song deleted",
			slog.String("artist", group),
			slog.String("title", song),
		)

		render.JSON(w, r, resp.OK())
	}
}
