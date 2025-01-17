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
// @Description Fetches a list of songs with optional filters for artist, song title, release date, lyrics and link presence.
// @Tags songs
// @Accept  json
// @Produce  json
// @Param group query string false "Artist Name" Example("The Beatles")
// @Param song query string false "Song Title" Example("Hey Jude")
// @Param release_date query string false "Release Date (single date or range: 'DD-MM-YYYY' or 'DD-MM-YYYY,DD-MM-YYYY')" Example("01-01-1970,31-12-1979")
// @Param lyrics query string false "Lyrics content or 'not_null' to filter songs with lyrics" Example("love")
// @Param link query string false "Use 'not_null' to filter songs with links" Example("not_null")
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

		if releaseDate := r.URL.Query().Get("release_date"); releaseDate != "" {
			filter["release_date"] = releaseDate
		}

		if lyrics := r.URL.Query().Get("lyrics"); lyrics != "" {
			filter["lyrics"] = lyrics
		}

		if link := r.URL.Query().Get("link"); link != "" {
			filter["link"] = link
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

		log.Debug("songs fetched", slog.Any("filter", filter), slog.Any("limit", limit), slog.Any("offset", offset))

		render.JSON(w, r, response)
	}
}

func formatSongs(songs []*storage.Song) []*models.Song {
	formattedSongs := make([]*models.Song, len(songs))
	for i, song := range songs {
		releaseDate := ""
		if !song.ReleaseDate.IsZero() {
			releaseDate = song.ReleaseDate.Format("02.01.2006")
		}
		formattedSongs[i] = &models.Song{
			Artist:      song.Artist,
			Title:       song.Title,
			ReleaseDate: releaseDate,
			Text:        song.Lyrics,
			Link:        song.Link,
		}
	}
	return formattedSongs
}
