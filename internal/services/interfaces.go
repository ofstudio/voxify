package services

import (
	"context"

	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/internal/store"
)

// Store is a store interface.
type Store = store.Store

// Platform is an interface for different download platforms.
type Platform interface {
	ID() string
	Init(ctx context.Context) error
	Match(url string) bool
	Download(ctx context.Context, req entities.Request) (*entities.Episode, error)
}

// Downloader is an interface for downloading episodes.
type Downloader interface {
	Download(ctx context.Context, req entities.Request) (*entities.Episode, error)
}

// Builder is an interface for building the podcast feed.
type Builder interface {
	Build(ctx context.Context) error
}
