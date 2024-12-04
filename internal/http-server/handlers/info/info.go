package info

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	resp "song-library/internal/lib/api/response"
	"song-library/internal/lib/logger/sl"

	"song-library/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	Artist string `json:"group" validate:"required"`
	Title  string `json:"song" validate:"required"`
}

type Response struct {
	//resp.Response
	ReleaseDate string `json:"release_date,omitempty"`
	Text        string `json:"text,omitempty"`
	Link        string `json:"link,omitempty"`
}

type SongGetter interface {
	GetSong(ctx context.Context, artist string, title string) (*storage.Song, error)
}

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

		render.JSON(w, r, Response{
			//Response:    resp.OK(),
			ReleaseDate: song.ReleaseDate.Format("02.01.2006"),
			Text:        song.Lyrics,
			Link:        song.Link,
		})

	}
}
