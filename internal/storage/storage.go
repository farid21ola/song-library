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
}

var (
	ErrNoSavedPages = errors.New("no saved pages")
	ErrSongNotFound = errors.New("song not found")
)

type Song struct {
	Artist      string
	Title       string
	ReleaseDate time.Time
	Lyrics      string
	Link        string
}