package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ofstudio/voxify/internal/domain"
	_ "modernc.org/sqlite"
)

// SQLiteStore is a SQLite implementation of the Store interface.
type SQLiteStore struct {
	db     *sql.DB
	execer execer
}

// NewSQLiteStore creates a new SQLiteStore instance.
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{
		db:     db,
		execer: db,
	}
}

// Close closes the database connection.
func (s *SQLiteStore) Close() {
	_ = s.db.Close()
}

// Begin returns a new SQLiteStore within a transaction
func (s *SQLiteStore) Begin(ctx context.Context) (domain.Store, error) {
	if s.execer != s.db {
		return nil, fmt.Errorf("unable to start a transaction within another transaction")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &SQLiteStore{
		db:     s.db,
		execer: tx,
	}, nil
}

// Commit commits the transaction.
func (s *SQLiteStore) Commit() error {
	tx, ok := s.execer.(txer)
	if !ok {
		return fmt.Errorf("unable to commit outside of a transaction")
	}
	return tx.Commit()
}

// Rollback aborts the transaction.
func (s *SQLiteStore) Rollback() error {
	tx, ok := s.execer.(txer)
	if !ok {
		return fmt.Errorf("unable to rollback outside of a transaction")
	}
	return tx.Rollback()
}

// EpisodeCreate creates a new episode in the store.
func (s *SQLiteStore) EpisodeCreate(ctx context.Context, episode *domain.Episode) error {
	query := `
		INSERT INTO episodes (
			title, description, thumbnail_file, media_file, 
			media_duration, media_size, media_type, author, original_url, canonical_url
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at`

	var id int64
	var createdAt time.Time

	err := s.execer.QueryRowContext(ctx, query,
		episode.Title,
		episode.Description,
		episode.ThumbnailFile,
		episode.MediaFile,
		episode.MediaDuration,
		episode.MediaSize,
		string(episode.MediaType),
		episode.Author,
		episode.OriginalURL,
		episode.CanonicalURL,
	).Scan(&id, &createdAt)

	if err != nil {
		return fmt.Errorf("failed to create episode: %w", err)
	}

	episode.ID = id
	episode.CreatedAt = createdAt
	return nil
}

// EpisodeGet returns all episodes from the store in descending order by creation date.
// Supports pagination via pageSize and pageNumber parameters.
// Zero values for pageSize and pageNumber will return all episodes without pagination.
func (s *SQLiteStore) EpisodeGet(ctx context.Context, pageSize, pageNumber int) ([]*domain.Episode, error) {
	query := `
		SELECT id, title, description, thumbnail_file, media_file,
			   media_duration, media_size, media_type, author, original_url, canonical_url, created_at
		FROM episodes
		ORDER BY created_at DESC`

	// Add pagination if both pageSize and pageNumber are provided
	var args []interface{}
	if pageSize > 0 && pageNumber >= 0 {
		offset := pageNumber * pageSize
		query += ` LIMIT ? OFFSET ?`
		args = []interface{}{pageSize, offset}
	}

	rows, err := s.execer.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query episodes: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()

	var episodes []*domain.Episode
	for rows.Next() {
		episode := &domain.Episode{}
		var mediaType string
		err = rows.Scan(
			&episode.ID,
			&episode.Title,
			&episode.Description,
			&episode.ThumbnailFile,
			&episode.MediaFile,
			&episode.MediaDuration,
			&episode.MediaSize,
			&mediaType,
			&episode.Author,
			&episode.OriginalURL,
			&episode.CanonicalURL,
			&episode.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan episode: %w", err)
		}
		episode.MediaType = domain.MediaType(mediaType)
		episodes = append(episodes, episode)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over episodes: %w", err)
	}

	return episodes, nil
}

// EpisodeCount returns the total count of episodes in the store.
func (s *SQLiteStore) EpisodeCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM episodes`

	var count int
	err := s.execer.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count episodes: %w", err)
	}
	return count, nil
}

// EpisodeGetByUrl returns episodes matching the given original or canonical URL.
func (s *SQLiteStore) EpisodeGetByUrl(ctx context.Context, url string) ([]*domain.Episode, error) {
	query := `
		SELECT id, title, description, thumbnail_file, media_file,
			   media_duration, media_size, media_type, author, original_url, canonical_url, created_at
		FROM episodes
		WHERE original_url = ? OR canonical_url = ?
		ORDER BY created_at DESC`

	rows, err := s.execer.QueryContext(ctx, query, url, url)
	if err != nil {
		return nil, fmt.Errorf("failed to query episodes by URL: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()

	var episodes []*domain.Episode
	for rows.Next() {
		episode := &domain.Episode{}
		var mediaType string
		err = rows.Scan(
			&episode.ID,
			&episode.Title,
			&episode.Description,
			&episode.ThumbnailFile,
			&episode.MediaFile,
			&episode.MediaDuration,
			&episode.MediaSize,
			&mediaType,
			&episode.Author,
			&episode.OriginalURL,
			&episode.CanonicalURL,
			&episode.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan episode: %w", err)
		}
		episode.MediaType = domain.MediaType(mediaType)
		episodes = append(episodes, episode)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over episodes: %w", err)
	}

	return episodes, nil
}

// EpisodeGetOldest returns the oldest episodes from the store.
func (s *SQLiteStore) EpisodeGetOldest(ctx context.Context, limit int) ([]*domain.Episode, error) {
	query := `
		SELECT id, title, description, thumbnail_file, media_file,
			   media_duration, media_size, media_type, author, original_url, canonical_url, created_at
		FROM episodes
		ORDER BY created_at ASC
		LIMIT ?`

	rows, err := s.execer.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query oldest episodes: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()

	var episodes []*domain.Episode
	for rows.Next() {
		episode := &domain.Episode{}
		var mediaType string
		err = rows.Scan(
			&episode.ID,
			&episode.Title,
			&episode.Description,
			&episode.ThumbnailFile,
			&episode.MediaFile,
			&episode.MediaDuration,
			&episode.MediaSize,
			&mediaType,
			&episode.Author,
			&episode.OriginalURL,
			&episode.CanonicalURL,
			&episode.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan episode: %w", err)
		}
		episode.MediaType = domain.MediaType(mediaType)
		episodes = append(episodes, episode)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over episodes: %w", err)
	}

	return episodes, nil
}

// EpisodeDelete deletes an episode from the store.
func (s *SQLiteStore) EpisodeDelete(ctx context.Context, id int64) error {
	query := `DELETE FROM episodes WHERE id = ?`

	result, err := s.execer.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete episode: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get deleted episode count: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("episode not found: %d", id)
	}
	return nil
}

// execer defines the interface for executing SQL queries and commands.
// It abstracts the common database operations that can be performed
// on both regular database connections and transactions.
type execer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// txer extends execer with transaction-specific operations.
// It represents a database transaction that can be committed or rolled back.
type txer interface {
	execer
	Commit() error
	Rollback() error
}
