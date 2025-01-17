package postgres

import (
	"context"
	"errors"
	"fmt"
	"song-library/internal/storage"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(connectionString string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) GetSongs(ctx context.Context, filter *map[string]string, limit, offset int) ([]*storage.Song, error) {
	const op = "storage.postgres.GetSongs"

	query := `
		SELECT artist_name, title, release_date, lyrics, link
		FROM songs s
		JOIN artists a ON s.artist_id = a.artist_id
		`

	var conditions []string
	var args []interface{}

	if filter != nil {
		if artist, ok := (*filter)["artist"]; ok {
			conditions = append(conditions, fmt.Sprintf("a.artist_name ILIKE $%d", len(args)+1))
			args = append(args, "%"+artist+"%")
		}
		if title, ok := (*filter)["title"]; ok {
			conditions = append(conditions, fmt.Sprintf("s.title ILIKE $%d", len(args)+1))
			args = append(args, "%"+title+"%")
		}
		if releaseDate, ok := (*filter)["release_date"]; ok {
			dates := strings.Split(releaseDate, ",")
			if len(dates) == 1 {
				conditions = append(conditions, fmt.Sprintf("s.release_date = $%d", len(args)+1))
				args = append(args, dates[0])
			}
			if len(dates) == 2 {
				conditions = append(conditions, fmt.Sprintf("s.release_date BETWEEN $%d AND $%d", len(args)+1, len(args)+2))
				args = append(args, dates[0], dates[1])
			}
		}
		if lyrics, ok := (*filter)["lyrics"]; ok {
			if lyrics == "not_null" {
				conditions = append(conditions, "s.lyrics != ''")
			} else {
				conditions = append(conditions, fmt.Sprintf("s.lyrics ILIKE $%d", len(args)+1))
				args = append(args, "%"+lyrics+"%")
			}
		}
		if link, ok := (*filter)["link"]; ok {
			if link == "not_null" {
				conditions = append(conditions, "s.link != ''")
			}
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrSongNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var songs []*storage.Song
	for rows.Next() {
		var song storage.Song
		if err := rows.Scan(&song.Artist, &song.Title, &song.ReleaseDate, &song.Lyrics, &song.Link); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		songs = append(songs, &song)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return songs, nil
}

func (s *Storage) GetSongLyrics(ctx context.Context, artist, title string, limit, offset int) (string, error) {
	const op = "storage.postgres.GetSongLyrics"

	query := `
		SELECT s.lyrics 
		FROM songs s
		JOIN artists a ON s.artist_id = a.artist_id
		WHERE a.artist_name ILIKE $1 AND s.title ILIKE $2
	`

	var lyrics string
	err := s.db.QueryRow(ctx, query, artist, title).Scan(&lyrics)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrSongNotFound
		}
		return "", fmt.Errorf("%s execute statement: %w", op, err)
	}

	verses := strings.Split(lyrics, "\\n\\n")
	if offset >= len(verses) {
		return "", nil
	}

	end := offset + limit
	if end > len(verses) {
		end = len(verses)
	}

	return strings.Join(verses[offset:end], "\\n\\n"), nil
}

func (s *Storage) DeleteSong(ctx context.Context, artist, title string) error {
	const op = "storage.postgres.DeleteSong"

	query := `
		DELETE FROM songs s
		USING artists a
		WHERE s.artist_id = a.artist_id
		AND a.artist_name = $1 AND s.title = $2
	`

	res, err := s.db.Exec(ctx, query, artist, title)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return storage.ErrSongNotFound
	}

	return nil
}

func (s *Storage) UpdateSong(ctx context.Context, song *storage.Song) error {
	const op = "storage.postgres.UpdateSong"

	var (
		setClauses []string
		args       []interface{}
		argIndex   = 1
	)

	if song.Lyrics != "" {
		setClauses = append(setClauses, fmt.Sprintf("lyrics = $%d", argIndex))
		args = append(args, song.Lyrics)
		argIndex++
	}
	if song.Link != "" {
		setClauses = append(setClauses, fmt.Sprintf("link = $%d", argIndex))
		args = append(args, song.Link)
		argIndex++
	}
	if !song.ReleaseDate.IsZero() {
		setClauses = append(setClauses, fmt.Sprintf("release_date = $%d", argIndex))
		args = append(args, song.ReleaseDate)
		argIndex++
	}

	if len(setClauses) == 0 {
		return storage.NothingChanged
	}

	query := fmt.Sprintf(`UPDATE songs s
		SET %s
		FROM artists a
		WHERE s.artist_id = a.artist_id
		AND a.artist_name = $%d
		AND s.title = $%d`,
		strings.Join(setClauses, ", "), argIndex, argIndex+1)

	args = append(args, song.Artist, song.Title)

	_, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) AddSong(ctx context.Context, song *storage.Song) error {
	const op = "storage.postgres.AddSong"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback(ctx)

	var artistID int
	err = tx.QueryRow(ctx, "SELECT artist_id FROM artists WHERE artist_name = $1", song.Artist).Scan(&artistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, "INSERT INTO artists(artist_name) VALUES($1) RETURNING artist_id", song.Artist).Scan(&artistID)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		} else {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	query := `INSERT INTO songs(artist_id, title, release_date, lyrics, link) VALUES ($1,$2,$3,$4,$5)`

	_, err = tx.Exec(ctx, query, artistID, song.Title, song.ReleaseDate, song.Lyrics, song.Link)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 - уникальное ограничение нарушено
			return fmt.Errorf("%s: %w", op, storage.ErrSongExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetSong(ctx context.Context, artist, title string) (*storage.Song, error) {
	const op = "storage.postgres.GetSong"

	query := `
		SELECT a.artist_name, s.title, s.release_date, s.lyrics, s.link 
		FROM songs s
		JOIN artists a ON s.artist_id = a.artist_id
		WHERE a.artist_name ILIKE $1 AND s.title ILIKE $2`

	var song storage.Song
	err := s.db.QueryRow(ctx, query, artist, title).Scan(&song.Artist, &song.Title, &song.ReleaseDate, &song.Lyrics, &song.Link)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrSongNotFound
		}
		return nil, fmt.Errorf("%s execute statement: %w", op, err)
	}

	return &song, nil
}
