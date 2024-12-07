package get

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"song-library/internal/lib/api/resp"
	"song-library/internal/lib/logger/sl"
	"song-library/internal/models"
	"song-library/internal/storage"

	"github.com/go-chi/render"
)

type SongsGetter interface {
	GetSongs(ctx context.Context, filter *map[string]string, limit, offset int) ([]*storage.Song, error)
}

// @Summary Get a list of songs with optional filters and pagination.
// @Description Fetches a list of songs with optional filters for artist and song title, and supports pagination via limit and offset.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param group query string false "Artist Name" Example("The Beatles")
// @Param song query string false "Song Title" Example("Hey Jude")
// @Param limit query int false "Limit of songs to retrieve" Default(10)
// @Param offset query int false "Offset for pagination" Default(0)
// @Success 200 {array} models.Song "A list of songs"
// @Failure 400 {object} resp.Response "Bad Request"
// @Failure 500 {object} resp.Response "Internal Server Error"
// @Router /songs [get]
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

		response := formatSongs(songs)

		render.JSON(w, r, response)
	}
}

func formatSongs(songs []*storage.Song) []*models.Song {
	formattedSongs := make([]*models.Song, len(songs))
	for i, song := range songs {
		formattedSongs[i] = &models.Song{
			Artist:      song.Artist,
			Title:       song.Title,
			ReleaseDate: song.ReleaseDate.Format("02.01.2006"), // Форматирование даты
			Text:        song.Lyrics,
			Link:        song.Link,
		}
	}
	return formattedSongs
}
