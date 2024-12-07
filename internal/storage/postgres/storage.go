package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"song-library/internal/storage"
	"strings"
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

	query := "SELECT artist, title, release_date, lyrics, link FROM songs"
	var conditions []string
	var args []interface{}

	if filter != nil {
		for key, value := range *filter {
			conditions = append(conditions, fmt.Sprintf("%s ILIKE $%d", key, len(args)+1))
			args = append(args, "%"+value+"%")
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

	query := "SELECT lyrics FROM songs WHERE artist = $1 AND  title = $2"

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

	query := `DELETE FROM songs WHERE artist = $1 AND title = $2`

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

	query := `UPDATE songs SET release_date = $1, lyrics = $2, link = $3 WHERE artist = $4 AND title = $5`

	_, err := s.db.Exec(ctx, query, song.ReleaseDate, song.Lyrics, song.Link, song.Artist, song.Title)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) AddSong(ctx context.Context, song *storage.Song) error {
	const op = "storage.postgres.AddSong"

	query := `INSERT INTO songs(artist, title, release_date, lyrics, link) VALUES ($1,$2,$3,$4,$5)`

	_, err := s.db.Exec(ctx, query, song.Artist, song.Title, song.ReleaseDate, song.Lyrics, song.Link)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 - уникальное ограничение нарушено
			return fmt.Errorf("%s: %w", op, storage.ErrSongExists)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetSong(ctx context.Context, artist, title string) (*storage.Song, error) {
	const op = "storage.postgres.GetSong"

	query := "SELECT * FROM songs WHERE artist = $1 AND  title = $2"

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
