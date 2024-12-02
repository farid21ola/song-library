package storage

import (
	"context"
	"errors"
	"time"
)

type Storage interface {
	GetSongs(ctx context.Context, filter *map[string]string, limit, offset int) ([]*Song, error)
	GetSongLyrics(ctx context.Context, artist, title string, limit, offset int) (string, error)
	DeleteSong(ctx context.Context, artist, title string) error
	UpdateSong(ctx context.Context, song *Song) error
	AddSong(ctx context.Context, song *Song) error
	GetSong(ctx context.Context, artist, title string) (*Song, error)
}

var (
	ErrSongNotFound = errors.New("song not found")
	ErrSongExists   = errors.New("song exists")
)

type Song struct {
	Artist      string
	Title       string
	ReleaseDate time.Time
	Lyrics      string
	Link        string
}
