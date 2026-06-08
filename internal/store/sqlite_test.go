package store

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/domain"
)

// TestSQLiteStoreSuite is a test suite for SQLiteStore
type TestSQLiteStoreSuite struct {
	suite.Suite
	db    *sql.DB
	store *SQLiteStore
	ctx   context.Context
}

// SetupTest is called once before each test start
func (suite *TestSQLiteStoreSuite) SetupTest() {
	suite.SetupSubTest()
}

// SetupSubTest is called before each subtest in the suite
func (suite *TestSQLiteStoreSuite) SetupSubTest() {
	var err error
	suite.db, err = NewSQLite(":memory:", 1)
	suite.Require().NoError(err, "Failed to create in-memory database")
	suite.store = NewSQLiteStore(suite.db)
	suite.ctx = context.Background()
}

// TearDownTest is called after each test in the suite completes
func (suite *TestSQLiteStoreSuite) TearDownTest() {
	suite.TearDownSubTest()
}

// TearDownSubTest is called after each subtest in the suite
func (suite *TestSQLiteStoreSuite) TearDownSubTest() {
	if suite.db != nil {
		suite.Require().NoError(suite.db.Close())
	}
}

// Test Episode methods

func (suite *TestSQLiteStoreSuite) TestEpisodeCreate() {
	// Arrange
	episode := &domain.Episode{
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
	var storedEpisode domain.Episode
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
	storedEpisode.MediaType = domain.MediaType(mediaType)

	// Verify all fields match exactly
	suite.Equal(episode.ID, storedEpisode.ID)
	suite.Equal("Test Episode", storedEpisode.Title)
	suite.Equal("Test Description", storedEpisode.Description)
	suite.Equal("thumb.jpg", storedEpisode.ThumbnailFile)
	suite.Equal("audio.mp3", storedEpisode.MediaFile)
	suite.Equal(int64(3600), storedEpisode.MediaDuration)
	suite.Equal(int64(1024000), storedEpisode.MediaSize)
	suite.Equal(domain.MediaType("audio/mpeg"), storedEpisode.MediaType)
	suite.Equal("Test Author", storedEpisode.Author)
	suite.Equal("https://example.com/original", storedEpisode.OriginalURL)
	suite.Equal("https://example.com/canonical", storedEpisode.CanonicalURL)
	suite.Equal(episode.CreatedAt, storedEpisode.CreatedAt)
}

func (suite *TestSQLiteStoreSuite) TestEpisodeGet() {
	suite.Run("All episodes without pagination", func() {
		// Arrange - create test episodes
		episodes := []*domain.Episode{
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

		// Act - get all episodes without pagination (0, 0)
		result, err := suite.store.EpisodeGet(suite.ctx, 0, 0)

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
				suite.Equal(domain.MediaType("audio/mpeg"), ep.MediaType)
				suite.Equal("Author 1", ep.Author)
				suite.Equal("https://example.com/1", ep.OriginalURL)
				suite.Equal("https://example.com/1", ep.CanonicalURL)
			} else if ep.Title == "Episode 2" {
				suite.Equal("Description 2", ep.Description)
				suite.Equal("thumb2.jpg", ep.ThumbnailFile)
				suite.Equal("audio2.mp3", ep.MediaFile)
				suite.Equal(int64(2400), ep.MediaDuration)
				suite.Equal(int64(768000), ep.MediaSize)
				suite.Equal(domain.MediaType("audio/mpeg"), ep.MediaType)
				suite.Equal("Author 2", ep.Author)
				suite.Equal("https://example.com/2", ep.OriginalURL)
				suite.Equal("https://example.com/2", ep.CanonicalURL)
			}
		}
	})

	suite.Run("With pagination", func() {
		// Arrange - create 5 test episodes
		for i := 1; i <= 5; i++ {
			episode := &domain.Episode{
				Title:         fmt.Sprintf("Episode %d", i),
				Description:   fmt.Sprintf("Description %d", i),
				ThumbnailFile: fmt.Sprintf("thumb%d.jpg", i),
				MediaFile:     fmt.Sprintf("audio%d.mp3", i),
				MediaDuration: int64(1800 * i),
				MediaSize:     int64(512000 * i),
				MediaType:     "audio/mpeg",
				Author:        fmt.Sprintf("Author %d", i),
				OriginalURL:   fmt.Sprintf("https://example.com/%d", i),
				CanonicalURL:  fmt.Sprintf("https://example.com/%d", i),
			}
			err := suite.store.EpisodeCreate(suite.ctx, episode)
			suite.Require().NoError(err)
		}

		// Act - get first page with pageSize=2
		result, err := suite.store.EpisodeGet(suite.ctx, 2, 0)

		// Assert
		suite.Require().NoError(err)
		suite.Len(result, 2, "Should return 2 episodes for first page")

		// Act - get second page with pageSize=2
		result2, err := suite.store.EpisodeGet(suite.ctx, 2, 1)

		// Assert
		suite.Require().NoError(err)
		suite.Len(result2, 2, "Should return 2 episodes for second page")

		// Act - get third page with pageSize=2
		result3, err := suite.store.EpisodeGet(suite.ctx, 2, 2)

		// Assert
		suite.Require().NoError(err)
		suite.Len(result3, 1, "Should return 1 episode for third page")

		// Verify pagination doesn't return duplicates
		allIds := make(map[int64]bool)
		for _, ep := range append(append(result, result2...), result3...) {
			suite.False(allIds[ep.ID], "Episode ID should be unique across pages")
			allIds[ep.ID] = true
		}
	})

	suite.Run("Empty result", func() {
		// Act
		result, err := suite.store.EpisodeGet(suite.ctx, 0, 0)

		// Assert
		suite.Require().NoError(err)
		suite.Empty(result, "Should return empty slice when no episodes exist")
	})
}

func (suite *TestSQLiteStoreSuite) TestEpisodeGetByUrl() {
	suite.Run("Found", func() {
		// Arrange
		targetURL := "https://example.com/target"
		episodes := []*domain.Episode{
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
		result, err := suite.store.EpisodeGetByUrl(suite.ctx, targetURL)

		// Assert
		suite.Require().NoError(err)
		suite.Len(result, 1, "Should return exactly 1 episode")
		suite.Equal("Target Episode", result[0].Title)
		suite.Equal(targetURL, result[0].OriginalURL)
		suite.Equal(int64(256000), result[0].MediaSize)
		suite.Equal(domain.MediaType("audio/mpeg"), result[0].MediaType)
	})

	suite.Run("NotFound", func() {
		// Act
		result, err := suite.store.EpisodeGetByUrl(suite.ctx, "https://nonexistent.com")

		// Assert
		suite.Require().NoError(err)
		suite.Empty(result, "Should return empty slice for non-existent URL")
	})
}

func (suite *TestSQLiteStoreSuite) TestEpisodeCount() {
	suite.Run("Multiple", func() {
		// Insert several episodes directly (author is required)
		for i := 1; i <= 3; i++ {
			_, err := suite.db.ExecContext(suite.ctx, `
				INSERT INTO episodes (title, media_file, media_type, author, original_url, canonical_url)
				VALUES (?, ?, ?, ?, ?, ?)
			`,
				fmt.Sprintf("Ep %d", i),
				fmt.Sprintf("ep%d.mp3", i),
				"audio/mpeg",
				fmt.Sprintf("author%d", i),
				fmt.Sprintf("https://example.com/%d", i),
				fmt.Sprintf("https://example.com/%d", i),
			)
			suite.Require().NoError(err)
		}

		// Act
		count, err := suite.store.EpisodeCount(suite.ctx)

		// Assert
		suite.Require().NoError(err)
		suite.Equal(3, count)
	})

	suite.Run("Zero", func() {
		// Act
		count, err := suite.store.EpisodeCount(suite.ctx)

		// Assert
		suite.Require().NoError(err)
		suite.Equal(0, count)
	})

	suite.Run("DatabaseError", func() {
		// Simulate database error by closing the DB connection
		suite.Require().NoError(suite.db.Close())

		// Act
		count, err := suite.store.EpisodeCount(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Equal(0, count)
	})
}

func (suite *TestSQLiteStoreSuite) TestEpisodeGetOldest() {
	// Arrange
	for i := 1; i <= 3; i++ {
		_, err := suite.db.ExecContext(suite.ctx, `
			INSERT INTO episodes (title, description, thumbnail_file, media_file, media_type, author, original_url, canonical_url, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			fmt.Sprintf("Ep %d", i),
			fmt.Sprintf("Description %d", i),
			fmt.Sprintf("thumb%d.jpg", i),
			fmt.Sprintf("ep%d.mp3", i),
			"audio/mpeg",
			fmt.Sprintf("author%d", i),
			fmt.Sprintf("https://example.com/%d", i),
			fmt.Sprintf("https://example.com/%d", i),
			fmt.Sprintf("2026-01-0%d 00:00:00", i),
		)
		suite.Require().NoError(err)
	}

	// Act
	episodes, err := suite.store.EpisodeGetOldest(suite.ctx, 2)

	// Assert
	suite.Require().NoError(err)
	suite.Len(episodes, 2)
	suite.Equal("Ep 1", episodes[0].Title)
	suite.Equal("Ep 2", episodes[1].Title)
}

func (suite *TestSQLiteStoreSuite) TestEpisodeDelete() {
	suite.Run("Success", func() {
		// Arrange
		episode := &domain.Episode{
			Title:        "Deleted Episode",
			MediaFile:    "deleted.mp3",
			MediaType:    "audio/mpeg",
			Author:       "Author",
			OriginalURL:  "https://example.com/deleted",
			CanonicalURL: "https://example.com/deleted",
		}
		err := suite.store.EpisodeCreate(suite.ctx, episode)
		suite.Require().NoError(err)

		// Act
		err = suite.store.EpisodeDelete(suite.ctx, episode.ID)

		// Assert
		suite.Require().NoError(err)
		count, err := suite.store.EpisodeCount(suite.ctx)
		suite.Require().NoError(err)
		suite.Equal(0, count)
	})

	suite.Run("NotFound", func() {
		// Act
		err := suite.store.EpisodeDelete(suite.ctx, 404)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "episode not found")
	})
}

// Test Transaction methods

func (suite *TestSQLiteStoreSuite) TestBegin() {
	suite.Run("Commit", func() {
		// Act
		txStore, err := suite.store.Begin(suite.ctx)
		suite.Require().NoError(err)

		// Create episode in transaction
		episode := &domain.Episode{
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
		episodes, err := suite.store.EpisodeGet(suite.ctx, 0, 0)
		suite.Require().NoError(err)
		suite.Len(episodes, 1)
		suite.Equal("TX Episode", episodes[0].Title)
	})

	suite.Run("Rollback", func() {

		// Act
		txStore, err := suite.store.Begin(suite.ctx)
		suite.Require().NoError(err)

		// Create episode in transaction
		episode := &domain.Episode{
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
		episodes, err := suite.store.EpisodeGet(suite.ctx, 0, 0)
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
