package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/entities"
)

// TestSQLiteStoreSuite is a test suite for SQLiteStore
type TestSQLiteStoreSuite struct {
	suite.Suite
	db    *sql.DB
	store *SQLiteStore
	ctx   context.Context
}

// SetupSuite is called once before the entire test suite runs
func (suite *TestSQLiteStoreSuite) SetupSuite() {
	var err error
	// Use in-memory SQLite database
	suite.db, err = NewSQLite(":memory:", 1)
	suite.Require().NoError(err, "Failed to create in-memory database")
}

// TearDownSuite is called once after the entire test suite runs
func (suite *TestSQLiteStoreSuite) TearDownSuite() {
	if suite.db != nil {
		suite.Require().NoError(suite.db.Close())
	}
}

// SetupTest is called before each test method
func (suite *TestSQLiteStoreSuite) SetupTest() {
	suite.store = NewSQLiteStore(suite.db)
	suite.ctx = context.Background()
}

// TearDownTest is called after each test in the suite
func (suite *TestSQLiteStoreSuite) TearDownTest() {
	suite.cleanupTestData()
}

// TearDownSubTest is called after each subtest in the suite
func (suite *TestSQLiteStoreSuite) TearDownSubTest() {
	suite.cleanupTestData()
}

func (suite *TestSQLiteStoreSuite) cleanupTestData() {
	// Clean up test data
	_, err := suite.db.Exec("DELETE FROM processes")
	suite.Require().NoError(err)
	_, err = suite.db.Exec("DELETE FROM episodes")
	suite.Require().NoError(err)
}

// Test Episode methods

func (suite *TestSQLiteStoreSuite) TestEpisodeCreate() {
	// Arrange
	episode := &entities.Episode{
		Title:         "Test Episode",
		Description:   "Test Description",
		ThumbnailFile: "thumb.jpg",
		MediaFile:     "audio.mp3",
		MediaDuration: 3600,
		MediaSize:     1024000,
		MediaType:     "audio/mpeg",
		Author:        "Test Author",
		OriginalURL:   "https://example.com/original",
		CanonicalURL:  "https://example.com/canonical",
	}

	// Act
	err := suite.store.EpisodeCreate(suite.ctx, episode)

	// Assert
	suite.Require().NoError(err)
	suite.NotZero(episode.ID, "ID should be set after creation")
	suite.NotZero(episode.CreatedAt, "CreatedAt should be set after creation")

	// Verify all fields were stored correctly in database
	var storedEpisode entities.Episode
	var mediaType string
	err = suite.db.QueryRow(`
		SELECT id, title, description, thumbnail_file, media_file,
			   media_duration, media_size, media_type, author, original_url, canonical_url, created_at
		FROM episodes WHERE id = ?`, episode.ID).Scan(
		&storedEpisode.ID,
		&storedEpisode.Title,
		&storedEpisode.Description,
		&storedEpisode.ThumbnailFile,
		&storedEpisode.MediaFile,
		&storedEpisode.MediaDuration,
		&storedEpisode.MediaSize,
		&mediaType,
		&storedEpisode.Author,
		&storedEpisode.OriginalURL,
		&storedEpisode.CanonicalURL,
		&storedEpisode.CreatedAt,
	)
	suite.Require().NoError(err)
	storedEpisode.MediaType = entities.MediaType(mediaType)

	// Verify all fields match exactly
	suite.Equal(episode.ID, storedEpisode.ID)
	suite.Equal("Test Episode", storedEpisode.Title)
	suite.Equal("Test Description", storedEpisode.Description)
	suite.Equal("thumb.jpg", storedEpisode.ThumbnailFile)
	suite.Equal("audio.mp3", storedEpisode.MediaFile)
	suite.Equal(int64(3600), storedEpisode.MediaDuration)
	suite.Equal(int64(1024000), storedEpisode.MediaSize)
	suite.Equal(entities.MediaType("audio/mpeg"), storedEpisode.MediaType)
	suite.Equal("Test Author", storedEpisode.Author)
	suite.Equal("https://example.com/original", storedEpisode.OriginalURL)
	suite.Equal("https://example.com/canonical", storedEpisode.CanonicalURL)
	suite.Equal(episode.CreatedAt, storedEpisode.CreatedAt)
}

func (suite *TestSQLiteStoreSuite) TestEpisodeListAll() {
	// Arrange - create test episodes
	episodes := []*entities.Episode{
		{
			Title:         "Episode 1",
			Description:   "Description 1",
			ThumbnailFile: "thumb1.jpg",
			MediaFile:     "audio1.mp3",
			MediaDuration: 1800,
			MediaSize:     512000,
			MediaType:     "audio/mpeg",
			Author:        "Author 1",
			OriginalURL:   "https://example.com/1",
			CanonicalURL:  "https://example.com/1",
		},
		{
			Title:         "Episode 2",
			Description:   "Description 2",
			ThumbnailFile: "thumb2.jpg",
			MediaFile:     "audio2.mp3",
			MediaDuration: 2400,
			MediaSize:     768000,
			MediaType:     "audio/mpeg",
			Author:        "Author 2",
			OriginalURL:   "https://example.com/2",
			CanonicalURL:  "https://example.com/2",
		},
	}

	for _, ep := range episodes {
		err := suite.store.EpisodeCreate(suite.ctx, ep)
		suite.Require().NoError(err)
	}

	// Act
	result, err := suite.store.EpisodeListAll(suite.ctx)

	// Assert
	suite.Require().NoError(err)
	suite.Len(result, 2, "Should return 2 episodes")

	// Should be ordered by created_at DESC (newest first)
	suite.True(result[0].CreatedAt.After(result[1].CreatedAt) ||
		result[0].CreatedAt.Equal(result[1].CreatedAt),
		"Episodes should be ordered by creation time DESC")

	// Verify all fields are read correctly
	for _, ep := range result {
		suite.NotZero(ep.ID)
		suite.NotEmpty(ep.Title)
		suite.NotEmpty(ep.Description)
		suite.NotEmpty(ep.ThumbnailFile)
		suite.NotEmpty(ep.MediaFile)
		suite.NotZero(ep.MediaDuration)
		suite.NotZero(ep.MediaSize)
		suite.NotEmpty(string(ep.MediaType))
		suite.NotEmpty(ep.Author)
		suite.NotEmpty(ep.OriginalURL)
		suite.NotEmpty(ep.CanonicalURL)
		suite.NotZero(ep.CreatedAt)

		// Verify specific values
		if ep.Title == "Episode 1" {
			suite.Equal("Description 1", ep.Description)
			suite.Equal("thumb1.jpg", ep.ThumbnailFile)
			suite.Equal("audio1.mp3", ep.MediaFile)
			suite.Equal(int64(1800), ep.MediaDuration)
			suite.Equal(int64(512000), ep.MediaSize)
			suite.Equal(entities.MediaType("audio/mpeg"), ep.MediaType)
			suite.Equal("Author 1", ep.Author)
			suite.Equal("https://example.com/1", ep.OriginalURL)
			suite.Equal("https://example.com/1", ep.CanonicalURL)
		} else if ep.Title == "Episode 2" {
			suite.Equal("Description 2", ep.Description)
			suite.Equal("thumb2.jpg", ep.ThumbnailFile)
			suite.Equal("audio2.mp3", ep.MediaFile)
			suite.Equal(int64(2400), ep.MediaDuration)
			suite.Equal(int64(768000), ep.MediaSize)
			suite.Equal(entities.MediaType("audio/mpeg"), ep.MediaType)
			suite.Equal("Author 2", ep.Author)
			suite.Equal("https://example.com/2", ep.OriginalURL)
			suite.Equal("https://example.com/2", ep.CanonicalURL)
		}
	}
}

func (suite *TestSQLiteStoreSuite) TestEpisodeGetByOriginalUrl() {
	suite.Run("Found", func() {
		// Arrange
		targetURL := "https://example.com/target"
		episodes := []*entities.Episode{
			{
				Title:        "Target Episode",
				MediaFile:    "target.mp3",
				MediaSize:    256000,
				MediaType:    "audio/mpeg",
				OriginalURL:  targetURL,
				CanonicalURL: "https://example.com/target",
			},
			{
				Title:        "Other Episode",
				MediaFile:    "other.mp3",
				MediaSize:    384000,
				MediaType:    "audio/mpeg",
				OriginalURL:  "https://example.com/other",
				CanonicalURL: "https://example.com/other",
			},
		}

		for _, ep := range episodes {
			err := suite.store.EpisodeCreate(suite.ctx, ep)
			suite.Require().NoError(err)
		}

		// Act
		result, err := suite.store.EpisodeGetByOriginalUrl(suite.ctx, targetURL)

		// Assert
		suite.Require().NoError(err)
		suite.Len(result, 1, "Should return exactly 1 episode")
		suite.Equal("Target Episode", result[0].Title)
		suite.Equal(targetURL, result[0].OriginalURL)
		suite.Equal(int64(256000), result[0].MediaSize)
		suite.Equal(entities.MediaType("audio/mpeg"), result[0].MediaType)
	})

	suite.Run("NotFound", func() {
		// Act
		result, err := suite.store.EpisodeGetByOriginalUrl(suite.ctx, "https://nonexistent.com")

		// Assert
		suite.Require().NoError(err)
		suite.Empty(result, "Should return empty slice for non-existent URL")
	})
}

// Test Process methods

func (suite *TestSQLiteStoreSuite) TestProcessUpsert() {
	suite.Run("Insert", func() {
		// Arrange
		process := &entities.Process{
			Request: entities.Request{
				UserID:    12345,
				ChatID:    67890,
				MessageID: 111,
				Url:       "https://example.com/video",
				Force:     false,
			},
			Step:   entities.StepDownloading,
			Status: entities.StatusInProgress,
		}

		// Act
		err := suite.store.ProcessUpsert(suite.ctx, process)

		// Assert
		suite.Require().NoError(err)
		suite.NotZero(process.ID, "ID should be set after creation")
		suite.NotZero(process.CreatedAt, "CreatedAt should be set")
		suite.NotZero(process.UpdatedAt, "UpdatedAt should be set")

		// Verify all fields were stored correctly in database
		var storedProcess entities.Process
		var errorText sql.NullString
		err = suite.db.QueryRow(`
			SELECT id, request_user_id, request_chat_id, request_message_id, 
				   request_url, request_force, step, status, error, 
				   created_at, updated_at
			FROM processes WHERE id = ?`, process.ID).Scan(
			&storedProcess.ID,
			&storedProcess.Request.UserID,
			&storedProcess.Request.ChatID,
			&storedProcess.Request.MessageID,
			&storedProcess.Request.Url,
			&storedProcess.Request.Force,
			&storedProcess.Step,
			&storedProcess.Status,
			&errorText,
			&storedProcess.CreatedAt,
			&storedProcess.UpdatedAt,
		)
		suite.Require().NoError(err)

		// Verify all fields match exactly
		suite.Equal(process.ID, storedProcess.ID)
		suite.Equal(int64(12345), storedProcess.Request.UserID)
		suite.Equal(int64(67890), storedProcess.Request.ChatID)
		suite.Equal(111, storedProcess.Request.MessageID)
		suite.Equal("https://example.com/video", storedProcess.Request.Url)
		suite.Equal(false, storedProcess.Request.Force)
		suite.Equal(string(entities.StepDownloading), string(storedProcess.Step))
		suite.Equal(string(entities.StatusInProgress), string(storedProcess.Status))
		suite.False(errorText.Valid, "Error should be null")
		suite.Equal(process.CreatedAt, storedProcess.CreatedAt)
		suite.Equal(process.UpdatedAt, storedProcess.UpdatedAt)
	})

	suite.Run("Update", func() {
		// Arrange - create initial process
		process := &entities.Process{
			Request: entities.Request{
				UserID:    12345,
				ChatID:    67890,
				MessageID: 111,
				Url:       "https://example.com/video",
				Force:     false,
			},
			Step:   entities.StepDownloading,
			Status: entities.StatusInProgress,
		}

		err := suite.store.ProcessUpsert(suite.ctx, process)
		suite.Require().NoError(err)

		originalUpdatedAt := process.UpdatedAt
		time.Sleep(1001 * time.Millisecond) // Ensure time difference

		// Act - update the process
		process.Step = entities.StepPublishing
		process.Status = entities.StatusSuccess
		err = suite.store.ProcessUpsert(suite.ctx, process)

		// Assert
		suite.Require().NoError(err)
		suite.True(process.UpdatedAt.After(originalUpdatedAt),
			"UpdatedAt should be newer after update")

		// Verify update was stored
		var storedStep, storedStatus string
		err = suite.db.QueryRow("SELECT step, status FROM processes WHERE id = ?",
			process.ID).Scan(&storedStep, &storedStatus)
		suite.Require().NoError(err)
		suite.Equal(string(entities.StepPublishing), storedStep)
		suite.Equal(string(entities.StatusSuccess), storedStatus)
	})

	suite.Run("WithEpisodeAndError", func() {
		// Arrange - create episode first
		episode := &entities.Episode{
			Title:        "Test Episode",
			MediaFile:    "test.mp3",
			MediaSize:    512000,
			MediaType:    "audio/mpeg",
			OriginalURL:  "https://example.com/test",
			CanonicalURL: "https://example.com/test",
		}
		err := suite.store.EpisodeCreate(suite.ctx, episode)
		suite.Require().NoError(err)

		// Create process with episode and error
		testError := errors.New("test error")
		process := &entities.Process{
			Request: entities.Request{
				UserID:    12345,
				ChatID:    67890,
				MessageID: 111,
				Url:       "https://example.com/video",
				Force:     true,
			},
			Step:    entities.StepDownloading,
			Status:  entities.StatusFailed,
			Error:   testError,
			Episode: episode,
		}

		// Act
		err = suite.store.ProcessUpsert(suite.ctx, process)

		// Assert
		suite.Require().NoError(err)

		// Verify all fields including episode_id and error were stored
		var storedProcess entities.Process
		var episodeID sql.NullInt64
		var errorText sql.NullString
		err = suite.db.QueryRow(`
			SELECT id, request_user_id, request_chat_id, request_message_id,
				   request_url, request_force, step, status, error, episode_id,
				   created_at, updated_at
			FROM processes WHERE id = ?`, process.ID).Scan(
			&storedProcess.ID,
			&storedProcess.Request.UserID,
			&storedProcess.Request.ChatID,
			&storedProcess.Request.MessageID,
			&storedProcess.Request.Url,
			&storedProcess.Request.Force,
			&storedProcess.Step,
			&storedProcess.Status,
			&errorText,
			&episodeID,
			&storedProcess.CreatedAt,
			&storedProcess.UpdatedAt,
		)
		suite.Require().NoError(err)

		// Verify all fields match
		suite.Equal(process.ID, storedProcess.ID)
		suite.Equal(int64(12345), storedProcess.Request.UserID)
		suite.Equal(int64(67890), storedProcess.Request.ChatID)
		suite.Equal(111, storedProcess.Request.MessageID)
		suite.Equal("https://example.com/video", storedProcess.Request.Url)
		suite.Equal(true, storedProcess.Request.Force)
		suite.Equal(string(entities.StepDownloading), string(storedProcess.Step))
		suite.Equal(string(entities.StatusFailed), string(storedProcess.Status))
		suite.True(episodeID.Valid, "Episode ID should be set")
		suite.Equal(episode.ID, episodeID.Int64)
		suite.True(errorText.Valid, "Error should be set")
		suite.Contains(errorText.String, "test error")
	})

	suite.Run("WithNewRequestFields", func() {
		// Arrange - create process with all new Request fields
		process := &entities.Process{
			Request: entities.Request{
				ID:              "req-12345-abcdef",
				UserID:          12345,
				ChatID:          67890,
				MessageID:       111,
				Url:             "https://example.com/video",
				DownloadFormat:  entities.DownloadMp3,
				DownloadQuality: "192",
				Force:           true,
			},
			Step:   entities.StepDownloading,
			Status: entities.StatusInProgress,
		}

		// Act
		err := suite.store.ProcessUpsert(suite.ctx, process)

		// Assert
		suite.Require().NoError(err)
		suite.NotZero(process.ID, "Process ID should be set after creation")
		suite.NotZero(process.CreatedAt, "CreatedAt should be set")
		suite.NotZero(process.UpdatedAt, "UpdatedAt should be set")

		// Verify all new fields were stored correctly in database
		var storedProcess entities.Process
		var requestID, requestDownloadFormat, requestDownloadQuality sql.NullString
		var errorText sql.NullString
		err = suite.db.QueryRow(`
			SELECT id, request_id, request_user_id, request_chat_id, request_message_id, 
				   request_url, request_download_format, request_download_quality, request_force, 
				   step, status, error, created_at, updated_at
			FROM processes WHERE id = ?`, process.ID).Scan(
			&storedProcess.ID,
			&requestID,
			&storedProcess.Request.UserID,
			&storedProcess.Request.ChatID,
			&storedProcess.Request.MessageID,
			&storedProcess.Request.Url,
			&requestDownloadFormat,
			&requestDownloadQuality,
			&storedProcess.Request.Force,
			&storedProcess.Step,
			&storedProcess.Status,
			&errorText,
			&storedProcess.CreatedAt,
			&storedProcess.UpdatedAt,
		)
		suite.Require().NoError(err)

		// Verify all fields match exactly, including new ones
		suite.Equal(process.ID, storedProcess.ID)
		suite.True(requestID.Valid, "Request ID should be stored")
		suite.Equal("req-12345-abcdef", requestID.String)
		suite.Equal(int64(12345), storedProcess.Request.UserID)
		suite.Equal(int64(67890), storedProcess.Request.ChatID)
		suite.Equal(111, storedProcess.Request.MessageID)
		suite.Equal("https://example.com/video", storedProcess.Request.Url)
		suite.True(requestDownloadFormat.Valid, "Download format should be stored")
		suite.Equal("mp3", requestDownloadFormat.String)
		suite.True(requestDownloadQuality.Valid, "Download quality should be stored")
		suite.Equal("192", requestDownloadQuality.String)
		suite.Equal(true, storedProcess.Request.Force)
		suite.Equal(string(entities.StepDownloading), string(storedProcess.Step))
		suite.Equal(string(entities.StatusInProgress), string(storedProcess.Status))
		suite.False(errorText.Valid, "Error should be null")
		suite.Equal(process.CreatedAt, storedProcess.CreatedAt)
		suite.Equal(process.UpdatedAt, storedProcess.UpdatedAt)

		// Test that ProcessGetByStatus properly loads the new fields
		result, err := suite.store.ProcessGetByStatus(suite.ctx, entities.StatusInProgress)
		suite.Require().NoError(err)
		suite.Len(result, 1)

		retrievedProcess := result[0]
		suite.Equal("req-12345-abcdef", retrievedProcess.Request.ID)
		suite.Equal(entities.DownloadMp3, retrievedProcess.Request.DownloadFormat)
		suite.Equal("192", retrievedProcess.Request.DownloadQuality)
		suite.Equal(true, retrievedProcess.Request.Force)
	})
}

func (suite *TestSQLiteStoreSuite) TestProcessGetByStatus() {
	suite.Run("WithResults", func() {
		// Arrange - create processes with different statuses
		processes := []*entities.Process{
			{
				Request: entities.Request{
					ID:     "req-1",
					UserID: 1, ChatID: 1, MessageID: 1,
					Url: "https://example.com/1", Force: false,
				},
				Step:   entities.StepDownloading,
				Status: entities.StatusInProgress,
			},
			{
				Request: entities.Request{
					ID:     "req-2",
					UserID: 2, ChatID: 2, MessageID: 2,
					Url: "https://example.com/2", Force: true,
				},
				Step:   entities.StepPublishing,
				Status: entities.StatusSuccess,
			},
			{
				Request: entities.Request{
					ID:     "req-3",
					UserID: 3, ChatID: 3, MessageID: 3,
					Url: "https://example.com/3", Force: false,
				},
				Step:   entities.StepDownloading,
				Status: entities.StatusInProgress,
			},
		}

		for _, p := range processes {
			err := suite.store.ProcessUpsert(suite.ctx, p)
			suite.Require().NoError(err)
		}

		// Act
		result, err := suite.store.ProcessGetByStatus(suite.ctx, entities.StatusInProgress)

		// Assert
		suite.Require().NoError(err)
		suite.Len(result, 2, "Should return 2 processes with in_progress status")

		for _, p := range result {
			// Verify all fields are read correctly
			suite.Equal(entities.StatusInProgress, p.Status)
			suite.NotZero(p.ID)
			suite.NotZero(p.Request.UserID)
			suite.NotZero(p.Request.ChatID)
			suite.NotZero(p.Request.MessageID)
			suite.NotEmpty(p.Request.Url)
			suite.Contains([]bool{true, false}, p.Request.Force) // Either true or false
			suite.Equal(entities.StepDownloading, p.Step)
			suite.NotZero(p.CreatedAt)
			suite.NotZero(p.UpdatedAt)
			suite.Nil(p.Error, "Error should be nil for in-progress processes")
			suite.Nil(p.Episode, "Episode should be nil for these test processes")
		}
	})

	suite.Run("WithEpisode", func() {
		// Arrange - create episode
		episode := &entities.Episode{
			Title:        "Test Episode",
			MediaFile:    "test.mp3",
			MediaSize:    768000,
			MediaType:    "audio/mpeg",
			OriginalURL:  "https://example.com/test",
			CanonicalURL: "https://example.com/test",
		}
		err := suite.store.EpisodeCreate(suite.ctx, episode)
		suite.Require().NoError(err)

		// Create process with episode
		process := &entities.Process{
			Request: entities.Request{
				UserID: 1, ChatID: 1, MessageID: 1,
				Url: "https://example.com/video",
			},
			Step:    entities.StepPublishing,
			Status:  entities.StatusSuccess,
			Episode: episode,
		}
		err = suite.store.ProcessUpsert(suite.ctx, process)
		suite.Require().NoError(err)

		// Act
		result, err := suite.store.ProcessGetByStatus(suite.ctx, entities.StatusSuccess)

		// Assert
		suite.Require().NoError(err)
		suite.Len(result, 1)

		retrievedProcess := result[0]
		suite.NotNil(retrievedProcess.Episode, "Episode should be loaded")
		suite.Equal(episode.ID, retrievedProcess.Episode.ID)
		suite.Equal("Test Episode", retrievedProcess.Episode.Title)
		suite.Equal("test.mp3", retrievedProcess.Episode.MediaFile)
		suite.Equal(int64(768000), retrievedProcess.Episode.MediaSize)
		suite.Equal(entities.MediaType("audio/mpeg"), retrievedProcess.Episode.MediaType)
	})

	suite.Run("NoResults", func() {
		// Act
		result, err := suite.store.ProcessGetByStatus(suite.ctx, entities.StatusInProgress)

		// Assert
		suite.Require().NoError(err)
		suite.Empty(result, "Should return empty slice when no processes match")
	})
}

func (suite *TestSQLiteStoreSuite) TestProcessCountByUrlAndStatus() {
	// Arrange
	url := "https://example.com/video"
	processes := []*entities.Process{
		{
			Request: entities.Request{
				ID: "aaa", UserID: 1, ChatID: 1, MessageID: 1, Url: url,
			},
			Step: entities.StepDownloading, Status: entities.StatusInProgress,
		},
		{
			Request: entities.Request{
				ID: "bbb", UserID: 2, ChatID: 2, MessageID: 2, Url: url,
			},
			Step: entities.StepPublishing, Status: entities.StatusInProgress,
		},
		{
			Request: entities.Request{
				ID: "ccc", UserID: 3, ChatID: 3, MessageID: 3, Url: "https://other.com",
			},
			Step: entities.StepDownloading, Status: entities.StatusInProgress,
		},
		{
			Request: entities.Request{
				ID: "ddd", UserID: 4, ChatID: 4, MessageID: 4, Url: url,
			},
			Step: entities.StepPublishing, Status: entities.StatusSuccess,
		},
	}

	for _, p := range processes {
		err := suite.store.ProcessUpsert(suite.ctx, p)
		suite.Require().NoError(err)
	}

	// Act
	count, err := suite.store.ProcessCountByUrlAndStatus(suite.ctx, url, entities.StatusInProgress)

	// Assert
	suite.Require().NoError(err)
	suite.Equal(2, count, "Should count 2 processes with matching URL and status")
}

// Test Transaction methods

func (suite *TestSQLiteStoreSuite) TestBegin() {
	suite.Run("Commit", func() {
		// Act
		txStore, err := suite.store.Begin(suite.ctx)
		suite.Require().NoError(err)

		// Create episode in transaction
		episode := &entities.Episode{
			Title:        "TX Episode",
			MediaFile:    "tx.mp3",
			MediaType:    "audio/mpeg",
			OriginalURL:  "https://example.com/tx",
			CanonicalURL: "https://example.com/tx",
		}
		err = txStore.EpisodeCreate(suite.ctx, episode)
		suite.Require().NoError(err)

		// Commit transaction
		err = txStore.Commit()
		suite.Require().NoError(err)

		// Assert - data should be persisted
		episodes, err := suite.store.EpisodeListAll(suite.ctx)
		suite.Require().NoError(err)
		suite.Len(episodes, 1)
		suite.Equal("TX Episode", episodes[0].Title)
	})

	suite.Run("Rollback", func() {

		// Act
		txStore, err := suite.store.Begin(suite.ctx)
		suite.Require().NoError(err)

		// Create episode in transaction
		episode := &entities.Episode{
			Title:        "TX Episode",
			MediaFile:    "tx.mp3",
			MediaType:    "audio/mpeg",
			OriginalURL:  "https://example.com/tx",
			CanonicalURL: "https://example.com/tx",
		}
		err = txStore.EpisodeCreate(suite.ctx, episode)
		suite.Require().NoError(err)

		// Rollback transaction
		err = txStore.Rollback()
		suite.Require().NoError(err)

		// Assert - data should not be persisted
		episodes, err := suite.store.EpisodeListAll(suite.ctx)
		suite.Require().NoError(err)
		suite.Empty(episodes, "No episodes should exist after rollback")
	})

	suite.Run("NestedTransaction", func() {
		// Act
		txStore, err := suite.store.Begin(suite.ctx)
		suite.Require().NoError(err)

		// Try to begin another transaction
		_, err = txStore.Begin(suite.ctx)

		// Assert
		suite.Error(err, "Should not allow nested transactions")
		suite.Contains(err.Error(), "unable to start a transaction within another transaction")
		suite.Require().NoError(txStore.Rollback())
	})
}

func (suite *TestSQLiteStoreSuite) TestCommit() {
	suite.Run("WithoutTransaction", func() {
		// Act
		err := suite.store.Commit()

		// Assert
		suite.Error(err, "Should error when committing without transaction")
		suite.Contains(err.Error(), "unable to commit outside of a transaction")
	})
}

func (suite *TestSQLiteStoreSuite) TestRollback() {
	suite.Run("WithoutTransaction", func() {

		// Act
		err := suite.store.Rollback()

		// Assert
		suite.Error(err, "Should error when rolling back without transaction")
		suite.Contains(err.Error(), "unable to rollback outside of a transaction")
	})
}

// TestStore is the entry point for running the test suite
func TestStore(t *testing.T) {
	suite.Run(t, new(TestSQLiteStoreSuite))
}
