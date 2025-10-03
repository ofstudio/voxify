package domain

import (
	"context"
)

// FeedBuilder should build podcast RSS feed and landing page.
type FeedBuilder interface {
	Build(ctx context.Context) error
	Info(ctx context.Context) (*FeedInfo, error)
}

// EpisodeDownloader should download episode.
type EpisodeDownloader interface {
	Validate(ctx context.Context, request DownloadRequest) error
	Download(ctx context.Context, request DownloadRequest) (*Episode, error)
}

// Platform should be able to download episodes from a specific platform.
type Platform interface {
	ID() string
	Init(ctx context.Context) error
	Match(request DownloadRequest) bool
	Download(ctx context.Context, request DownloadRequest) (*Episode, error)
}

// EventBus is an event bus for processing requests.
type EventBus interface {
	Publish(event Event)
	Subscribe(eventType EventType, handler EventHandler)
	Wait()
}

// Store defines the interface for a data store that manages episodes.
type Store interface {
	// Close the store and release all resources.
	Close()
	// Begin starts a new transaction and returns a new Store instance.
	Begin(ctx context.Context) (Store, error)
	// Commit commits the current transaction.
	Commit() error
	// Rollback aborts the current transaction.
	Rollback() error

	// EpisodeCreate creates a new episode in the store.
	EpisodeCreate(ctx context.Context, episode *Episode) error
	// EpisodeGet returns all episodes from the store in descending order by creation date.
	// Supports pagination via pageSize and pageNumber parameters.
	// Zero values for pageSize and pageNumber will return all episodes without pagination.
	EpisodeGet(ctx context.Context, pageSize, pageNumber int) ([]*Episode, error)
	// EpisodeCount returns the total count of episodes in the store.
	EpisodeCount(ctx context.Context) (int, error)
	// EpisodeGetByUrl returns episodes matching the given original or canonical URL.
	EpisodeGetByUrl(ctx context.Context, url string) ([]*Episode, error)
}
