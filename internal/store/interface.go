package store

import (
	"context"
	"time"

	"github.com/ofstudio/voxify/internal/entities"
)

// Store defines the interface for a data store that manages episodes and processes.
type Store interface {
	// Close the store and release all resources.
	Close()
	// Begin starts a new transaction and returns a new Store instance.
	Begin(ctx context.Context) (Store, error)
	// Commit commits the current transaction.
	Commit() error
	// Rollback aborts the current transaction.
	Rollback() error

	// Episode methods

	// EpisodeCreate creates a new episode record in the store.
	EpisodeCreate(ctx context.Context, episode *entities.Episode) error
	// EpisodeListAll returns all episodes from the store in descending order by creation date.
	EpisodeListAll(ctx context.Context) ([]*entities.Episode, error)
	// EpisodeGetByOriginalUrl returns episodes matching the given original URL.
	EpisodeGetByOriginalUrl(ctx context.Context, url string) ([]*entities.Episode, error)
	// EpisodeGetLastTime returns the creation time of the most recently added episode.
	// If no episodes exist, it returns zero time.
	EpisodeGetLastTime(ctx context.Context) (time.Time, error)

	// Process methods

	// ProcessUpsert creates or updates a process record in the store.
	ProcessUpsert(ctx context.Context, process *entities.Process) error
	// ProcessGetByStatus returns processes matching the given status.
	ProcessGetByStatus(ctx context.Context, status entities.Status) ([]*entities.Process, error)
	// ProcessCountByUrlAndStatus returns the count of processes matching the given URL and status.
	ProcessCountByUrlAndStatus(ctx context.Context, url string, status entities.Status) (int, error)
}
