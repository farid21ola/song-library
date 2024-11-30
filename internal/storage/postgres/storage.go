package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
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

	return []*storage.Song{}, nil
}

func (s *Storage) GetSongLyrics(ctx context.Context, artist, title string, limit, offset int) (string, error) {
	const op = "storage.postgres.GetSongLyrics"

	query := "SELECT lyrics FROM songs WHERE artist = $1 AND  title = $2 LIMIT $3 OFFSET $4"

	var lyrics string
	err := s.db.QueryRow(ctx, query, artist, title, limit, offset).Scan(&lyrics)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrSongNotFound
		}
		return "", fmt.Errorf("%s execute statement: %w", op, err)
	}

	verses := strings.Split(lyrics, "\n\n")
	if offset >= len(verses) {
		return "", nil
	}

	end := offset + limit
	if end > len(verses) {
		end = len(verses)
	}

	return strings.Join(verses[offset:end], "\n\n"), nil
}

func (s *Storage) DeleteSong(ctx context.Context, artist, title string) error {
	const op = "storage.postgres.DeleteSong"
}

func (s *Storage) UpdateSong(ctx context.Context, song *storage.Song) error {
	const op = "storage.postgres.UpdateSong"
}

func (s *Storage) AddSong(ctx context.Context, song *storage.Song) error {
	const op = "storage.postgres.AddSong"
}
