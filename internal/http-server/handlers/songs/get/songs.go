package get

import (
	"context"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "song-library/internal/lib/api/response"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/storage"
	"strconv"
)

type SongsGetter interface {
	GetSongs(ctx context.Context, filter *map[string]string, limit, offset int) ([]*storage.Song, error)
}

func New(log *slog.Logger, songsGetter SongsGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		filter := map[string]string{}
		if group := r.URL.Query().Get("group"); group != "" {
			filter["artist"] = group
		}
		if title := r.URL.Query().Get("song"); title != "" {
			filter["title"] = title
		}

		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit <= 0 {
			limit = 10
		}
		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil || offset < 0 {
			offset = 0
		}

		songs, err := songsGetter.GetSongs(r.Context(), &filter, limit, offset)
		if err != nil {
			log.Error("failed to get song ", sl.Err(err))

			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		render.JSON(w, r, songs)
	}
}
