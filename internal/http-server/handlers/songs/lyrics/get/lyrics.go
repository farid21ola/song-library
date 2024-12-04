package get

import (
	"context"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "song-library/internal/lib/api/response"
	"song-library/internal/lib/logger/sl"
	"strconv"
)

type Response struct {
	Lyrics string `json:"lyrics,omitempty"`
}

type LyricsGetter interface {
	GetSongLyrics(ctx context.Context, artist, title string, limit, offset int) (string, error)
}

func New(log *slog.Logger, lyricsGetter LyricsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		artist := r.URL.Query().Get("group")
		title := r.URL.Query().Get("song")

		if artist == "" || title == "" {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("missing required parameters: group and song"))
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

		render.JSON(w, r, Response{Lyrics: lyrics})
	}
}
