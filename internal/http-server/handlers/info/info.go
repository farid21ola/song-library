package info

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"song-library/internal/lib/api/resp"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/models"
	"song-library/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	Artist string `json:"group" validate:"required"`
	Title  string `json:"song" validate:"required"`
}

type SongGetter interface {
	GetSong(ctx context.Context, artist string, title string) (*storage.Song, error)
}

// @Summary Get song detail
// @Description Fetches the details of a song given an artist's name and song title.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param group query string true "Artist/group Name"
// @Param song query string true "Song Title"
// @Success 200 {object} models.SongDetail "Song details"
// @Failure 404 {object} resp.Response "Bad request"
// @Failure 404 {object} resp.Response "Not Found - Song not found"
// @Failure 500 {object} resp.Response "Internal Server Error"
// @Router /info [get]
func New(log *slog.Logger, songGetter SongGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.info.info.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		artist := r.URL.Query().Get("group")
		title := r.URL.Query().Get("song")

		if artist == "" || title == "" {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("missing required parameters: group and song"))
			return
		}

		song, err := songGetter.GetSong(r.Context(), artist, title)
		if errors.Is(err, storage.ErrSongNotFound) {
			log.Info("song not found",
				slog.String("artist", artist),
				slog.String("title", title),
			)

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}
		if err != nil {
			log.Error("failed to get song ", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		log.Info("song found",
			slog.String("artist", artist),
			slog.String("title", title),
		)

		songInfo := models.SongDetail{
			ReleaseDate: song.ReleaseDate.Format("02.01.2006"),
			Text:        song.Lyrics,
			Link:        song.Link,
		}
		render.JSON(w, r, songInfo)

	}
}
