package get

import (
	"context"
	"log/slog"
	"net/http"
	"song-library/internal/models"
	"strconv"

	"github.com/go-chi/render"

	"song-library/internal/lib/api/resp"
	"song-library/internal/lib/logger/sl"
)

type LyricsGetter interface {
	GetSongLyrics(ctx context.Context, artist, title string, limit, offset int) (string, error)
}

// @Summary Get song lyrics with optional pagination.
// @Description Fetches lyrics of a song by artist and title with optional pagination support for limit and offset.
// @Tags lyrics
// @Accept  json
// @Produce  json
// @Param group query string true "Artist Name" Example("The Beatles")
// @Param song query string true "Song Title" Example("Hey Jude")
// @Param limit query int false "Limit the number of lyrics lines to retrieve" Default(10)
// @Param offset query int false "Offset for pagination" Default(0)
// @Success 200 {object} models.Lyrics "Song lyrics successfully retrieved"
// @Failure 400 {object} resp.Response "Bad Request - Missing required parameters"
// @Failure 500 {object} resp.Response "Internal Server Error"
// @Router /songs/lyrics [get]
func New(log *slog.Logger, lyricsGetter LyricsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		artist := r.URL.Query().Get("group")
		title := r.URL.Query().Get("song")

		if artist == "" || title == "" {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("missing required parameters"))
			return
		}

		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit <= 0 {
			limit = 10
		}
		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil || offset < 0 {
			offset = 0
		}

		lyrics, err := lyricsGetter.GetSongLyrics(r.Context(), artist, title, limit, offset)
		if err != nil {
			log.Error("failed to fetch lyrics ", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Debug("song lyrics fetched",
			slog.String("artist", artist),
			slog.String("title", title),
		)

		render.JSON(w, r, models.Lyrics{Text: lyrics})
	}
}
