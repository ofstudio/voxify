package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ofstudio/voxify/internal/entities"
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
func (s *SQLiteStore) Begin(ctx context.Context) (Store, error) {
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

// EpisodeCreate creates a new episode in the database
func (s *SQLiteStore) EpisodeCreate(ctx context.Context, episode *entities.Episode) error {
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

// EpisodeListAll returns all episodes from the database
func (s *SQLiteStore) EpisodeListAll(ctx context.Context) ([]*entities.Episode, error) {
	query := `
		SELECT id, title, description, thumbnail_file, media_file,
			   media_duration, media_size, media_type, author, original_url, canonical_url, created_at
		FROM episodes
		ORDER BY created_at DESC`

	rows, err := s.execer.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query episodes: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()

	var episodes []*entities.Episode
	for rows.Next() {
		episode := &entities.Episode{}
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
		episode.MediaType = entities.MediaType(mediaType)
		episodes = append(episodes, episode)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over episodes: %w", err)
	}

	return episodes, nil
}

// EpisodeGetByOriginalUrl returns episodes by original URL
func (s *SQLiteStore) EpisodeGetByOriginalUrl(ctx context.Context, url string) ([]*entities.Episode, error) {
	query := `
		SELECT id, title, description, thumbnail_file, media_file,
			   media_duration, media_size, media_type, author, original_url, canonical_url, created_at
		FROM episodes
		WHERE original_url = ?
		ORDER BY created_at DESC`

	rows, err := s.execer.QueryContext(ctx, query, url)
	if err != nil {
		return nil, fmt.Errorf("failed to query episodes by URL: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()

	var episodes []*entities.Episode
	for rows.Next() {
		episode := &entities.Episode{}
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
		episode.MediaType = entities.MediaType(mediaType)
		episodes = append(episodes, episode)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over episodes: %w", err)
	}

	return episodes, nil
}

// ProcessUpsert creates or updates a process in the database
func (s *SQLiteStore) ProcessUpsert(ctx context.Context, process *entities.Process) error {
	var episodeID *int64
	if process.Episode != nil {
		episodeID = &process.Episode.ID
	}

	var errorText *string
	if process.Error != nil {
		errStr := process.Error.Error()
		errorText = &errStr
	}

	if process.ID == 0 {
		// INSERT with RETURNING
		query := `
			INSERT INTO processes (
				request_id, request_user_id, request_chat_id, request_message_id, 
				request_url, request_download_format, request_download_quality, request_force,
				step, status, error, episode_id
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			RETURNING id, created_at, updated_at`

		var id int64
		var createdAt, updatedAt time.Time

		err := s.execer.QueryRowContext(ctx, query,
			process.Request.ID,
			process.Request.UserID,
			process.Request.ChatID,
			process.Request.MessageID,
			process.Request.Url,
			string(process.Request.DownloadFormat),
			process.Request.DownloadQuality,
			process.Request.Force,
			string(process.Step),
			string(process.Status),
			errorText,
			episodeID,
		).Scan(&id, &createdAt, &updatedAt)

		if err != nil {
			return fmt.Errorf("failed to insert process: %w", err)
		}

		process.ID = id
		process.CreatedAt = createdAt
		process.UpdatedAt = updatedAt
	} else {
		// UPDATE with RETURNING
		query := `
			UPDATE processes SET
				request_id = ?, request_user_id = ?, request_chat_id = ?, request_message_id = ?, 
				request_url = ?, request_download_format = ?, request_download_quality = ?, request_force = ?, 
				step = ?, status = ?, error = ?, episode_id = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
			RETURNING updated_at`

		var updatedAt time.Time

		err := s.execer.QueryRowContext(ctx, query,
			process.Request.ID,
			process.Request.UserID,
			process.Request.ChatID,
			process.Request.MessageID,
			process.Request.Url,
			string(process.Request.DownloadFormat),
			process.Request.DownloadQuality,
			process.Request.Force,
			string(process.Step),
			string(process.Status),
			errorText,
			episodeID,
			process.ID,
		).Scan(&updatedAt)

		if err != nil {
			return fmt.Errorf("failed to update process: %w", err)
		}

		process.UpdatedAt = updatedAt
	}

	return nil
}

func (s *SQLiteStore) ProcessGetByStatus(ctx context.Context, status entities.Status) ([]*entities.Process, error) {
	query := `
		SELECT p.id, p.request_id, p.request_user_id, p.request_chat_id, p.request_message_id, 
			   p.request_url, p.request_download_format, p.request_download_quality, p.request_force, 
			   p.step, p.status, p.error, p.episode_id, p.created_at, p.updated_at,
			   e.id, e.title, e.description, e.thumbnail_file, e.media_file,
			   e.media_duration, e.media_size, e.media_type, e.author, e.original_url, e.canonical_url, e.created_at
		FROM processes p
		LEFT JOIN episodes e ON p.episode_id = e.id
		WHERE p.status = ?
		ORDER BY p.created_at DESC`

	rows, err := s.execer.QueryContext(ctx, query, string(status))
	if err != nil {
		return nil, fmt.Errorf("failed to query processes by status: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()

	return s.scanProcesses(rows)
}

func (s *SQLiteStore) ProcessCountByUrlAndStatus(ctx context.Context, url string, status entities.Status) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM processes 
		WHERE request_url = ? AND status = ?`

	var count int
	err := s.execer.QueryRowContext(ctx, query, url, string(status)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count processes by URL and status: %w", err)
	}

	return count, nil
}

// scanProcesses scans process rows with joined episode data
func (s *SQLiteStore) scanProcesses(rows *sql.Rows) ([]*entities.Process, error) {
	var processes []*entities.Process

	for rows.Next() {
		process := &entities.Process{}
		var episodeID sql.NullInt64
		var episodeTitle, episodeDesc, episodeThumbnail, episodeMedia, mediaType, episodeAuthor sql.NullString
		var episodeDuration, episodeMediaSize sql.NullInt64
		var episodeOriginalURL, episodeCanonicalURL sql.NullString
		var episodeCreatedAt sql.NullTime
		var errorText sql.NullString
		var requestDownloadFormat, requestDownloadQuality sql.NullString

		err := rows.Scan(
			&process.ID,
			&process.Request.ID,
			&process.Request.UserID,
			&process.Request.ChatID,
			&process.Request.MessageID,
			&process.Request.Url,
			&requestDownloadFormat,
			&requestDownloadQuality,
			&process.Request.Force,
			&process.Step,
			&process.Status,
			&errorText,
			&episodeID,
			&process.CreatedAt,
			&process.UpdatedAt,
			&episodeID,
			&episodeTitle,
			&episodeDesc,
			&episodeThumbnail,
			&episodeMedia,
			&episodeDuration,
			&episodeMediaSize,
			&mediaType,
			&episodeAuthor,
			&episodeOriginalURL,
			&episodeCanonicalURL,
			&episodeCreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan process: %w", err)
		}

		// Set request download format and quality
		if requestDownloadFormat.Valid {
			process.Request.DownloadFormat = entities.DownloadFormat(requestDownloadFormat.String)
		}
		if requestDownloadQuality.Valid {
			process.Request.DownloadQuality = requestDownloadQuality.String
		}

		// Create error if exists
		if errorText.Valid {
			process.Error = fmt.Errorf("%s", errorText.String)
		}

		// Create associated episode if exists
		if episodeID.Valid && episodeTitle.Valid {
			process.Episode = &entities.Episode{
				ID:            episodeID.Int64,
				Title:         episodeTitle.String,
				Description:   episodeDesc.String,
				ThumbnailFile: episodeThumbnail.String,
				MediaFile:     episodeMedia.String,
				MediaDuration: episodeDuration.Int64,
				MediaSize:     episodeMediaSize.Int64,
				MediaType:     entities.MediaType(mediaType.String),
				Author:        episodeAuthor.String,
				OriginalURL:   episodeOriginalURL.String,
				CanonicalURL:  episodeCanonicalURL.String,
				CreatedAt:     episodeCreatedAt.Time,
			}
		}

		processes = append(processes, process)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over processes: %w", err)
	}

	return processes, nil
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
