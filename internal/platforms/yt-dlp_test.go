package platforms

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/entities"
)

// TestYtDlpSuite is a test suite for YtDlp platform
type TestYtDlpSuite struct {
	suite.Suite
	ctx       context.Context
	cfg       *config.Settings
	log       *slog.Logger
	platform  *YtDlp
	tempDir   string
	publicDir string
}

// SetupSuite is called once before the entire test suite runs
func (suite *TestYtDlpSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create temporary directories for testing
	var err error
	suite.tempDir, err = os.MkdirTemp("", "voxify_test_download_*")
	suite.Require().NoError(err)

	suite.publicDir, err = os.MkdirTemp("", "voxify_test_public_*")
	suite.Require().NoError(err)

	suite.cfg = &config.Settings{
		DownloadDir:     suite.tempDir,
		PublicDir:       suite.publicDir,
		DownloadTimeout: 30 * time.Second,
		DownloadFormat:  entities.DownloadMp3,
		DownloadQuality: "192k",
		ThumbnailSize:   300,
		YtDlpPath:       "yt-dlp", // Assuming yt-dlp is in PATH for tests
		FFMpegPath:      "ffmpeg", // Assuming ffmpeg is in PATH for tests
	}
}

// TearDownSuite is called once after the entire test suite runs
func (suite *TestYtDlpSuite) TearDownSuite() {
	_ = os.RemoveAll(suite.tempDir)
	_ = os.RemoveAll(suite.publicDir)
}

// SetupTest is called before each test method
func (suite *TestYtDlpSuite) SetupTest() {
	suite.platform = NewYtDlpPlatform(*suite.cfg, suite.log)
}

// TestNewYtDlpPlatform tests the constructor
func (suite *TestYtDlpSuite) TestNewYtDlpPlatform() {
	// Act
	platform := NewYtDlpPlatform(*suite.cfg, suite.log)

	// Assert
	suite.NotNil(platform)
	suite.Equal(suite.cfg, &platform.cfg)
	suite.Equal(suite.log, platform.log)
}

// TestID tests the ID method
func (suite *TestYtDlpSuite) TestID() {
	// Act
	id := suite.platform.ID()

	// Assert
	suite.Equal("yt-dlp", id)
}

// TestMatch tests the Match method
func (suite *TestYtDlpSuite) TestMatch() {
	testCases := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "YouTube full URL",
			url:      "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: true,
		},
		{
			name:     "YouTube short URL",
			url:      "https://youtu.be/dQw4w9WgXcQ",
			expected: true,
		},
		{
			name:     "Non-YouTube URL",
			url:      "https://example.com/video",
			expected: false,
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: false,
		},
		{
			name:     "Invalid URL",
			url:      "not-a-url",
			expected: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Act
			result := suite.platform.Match(tc.url)

			// Assert
			suite.Equal(tc.expected, result)
		})
	}
}

// TestInit tests the Init method
func (suite *TestYtDlpSuite) TestInit() {
	suite.Run("Success", func() {
		// Skip if yt-dlp or ffmpeg are not available
		if !suite.checkCommand("yt-dlp") || !suite.checkCommand("ffmpeg") {
			suite.T().Skip("yt-dlp or ffmpeg not available")
		}

		// Act
		err := suite.platform.Init(suite.ctx)

		// Assert
		suite.NoError(err)
	})

	suite.Run("YtDlpPathEmpty", func() {
		// Arrange
		platform := NewYtDlpPlatform(config.Settings{
			YtDlpPath:  "",
			FFMpegPath: "ffmpeg",
		}, suite.log)

		// Act
		err := platform.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "yt-dlp path is not configured")
	})

	suite.Run("FFMpegPathEmpty", func() {
		// Arrange
		platform := NewYtDlpPlatform(config.Settings{
			YtDlpPath:  "yt-dlp",
			FFMpegPath: "",
		}, suite.log)

		// Act
		err := platform.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "ffmpeg path is not configured")
	})

	suite.Run("FFMpegNotFound", func() {
		// Arrange
		platform := NewYtDlpPlatform(config.Settings{
			YtDlpPath:  "yt-dlp",
			FFMpegPath: "non-existent-ffmpeg",
		}, suite.log)

		// Act
		err := platform.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "ffmpeg not found or not working")
	})

	suite.Run("YtDlpNotFound", func() {
		// Arrange
		platform := NewYtDlpPlatform(config.Settings{
			YtDlpPath:  "non-existent-yt-dlp",
			FFMpegPath: "ffmpeg",
		}, suite.log)

		// Act
		err := platform.Init(suite.ctx)

		// Assert
		suite.Error(err)
		suite.Contains(err.Error(), "yt-dlp not found or not working")
	})
}

// TestFetchMeta tests the fetchMeta method
func (suite *TestYtDlpSuite) TestFetchMeta() {
	suite.Run("Success", func() {
		// Skip if yt-dlp is not available
		if !suite.checkCommand("yt-dlp") {
			suite.T().Skip("yt-dlp not available")
		}

		// Arrange
		req := entities.Request{
			ID:  "test123",
			Url: "https://www.youtube.com/watch?v=dQw4w9WgXcQ", // Rick Roll - should always be available
		}

		tempDir, err := os.MkdirTemp(suite.tempDir, ytDlpPattern)
		suite.Require().NoError(err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Act
		meta, err := suite.platform.fetchMeta(suite.ctx, req, tempDir)

		// Assert
		suite.NoError(err)
		suite.NotNil(meta)
		suite.NotEmpty(meta.Title)
		suite.NotEmpty(meta.WebpageURL)
		suite.Greater(meta.Duration, int64(0))
	})

	suite.Run("InvalidURL", func() {
		// Skip if yt-dlp is not available
		if !suite.checkCommand("yt-dlp") {
			suite.T().Skip("yt-dlp not available")
		}

		// Arrange
		req := entities.Request{
			ID:  "test123",
			Url: "https://www.youtube.com/watch?v=invalid-video-id",
		}

		tempDir, err := os.MkdirTemp(suite.tempDir, ytDlpPattern)
		suite.Require().NoError(err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Act
		meta, err := suite.platform.fetchMeta(suite.ctx, req, tempDir)

		// Assert
		suite.Error(err)
		suite.Nil(meta)
		suite.Contains(err.Error(), "yt-dlp command failed")
	})

	suite.Run("EmptyTitle", func() {
		// Test with mock JSON response that has empty title
		suite.testFetchMetaWithMockOutput(`{"title": "", "uploader": "TestUploader", "description": "", "thumbnail": "", "duration": 100, "webpage_url": "https://example.com"}`)
	})

	suite.Run("EmptyDescription", func() {
		// Test with mock JSON response that has empty description
		suite.testFetchMetaWithMockOutput(`{"title": "TestTitle", "uploader": "TestUploader", "description": "", "thumbnail": "", "duration": 100, "webpage_url": "https://example.com"}`)
	})
}

// TestDownload tests the Download method
func (suite *TestYtDlpSuite) TestDownload() {
	suite.Run("UnsupportedFormat", func() {
		// Arrange
		req := entities.Request{
			ID:             "test123",
			Url:            "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			DownloadFormat: "unsupported",
		}

		// Act
		episode, err := suite.platform.Download(suite.ctx, req)

		// Assert
		suite.Error(err)
		suite.Nil(episode)
		suite.Contains(err.Error(), "unsupported download format")
	})

	// Note: Full integration test would require actual yt-dlp and ffmpeg
	// and a stable test video, which is complex for unit tests
}

// Helper methods

// checkCommand checks if a command is available in PATH
func (suite *TestYtDlpSuite) checkCommand(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// testFetchMetaWithMockOutput tests fetchMeta with a mock yt-dlp script
func (suite *TestYtDlpSuite) testFetchMetaWithMockOutput(jsonOutput string) {
	// Create a mock yt-dlp script
	mockScript := filepath.Join(suite.tempDir, "mock-yt-dlp")
	scriptContent := fmt.Sprintf(`#!/bin/bash
echo '%s'
`, jsonOutput)

	err := os.WriteFile(mockScript, []byte(scriptContent), 0755)
	suite.Require().NoError(err)

	// Create platform with mock script
	cfg := *suite.cfg
	cfg.YtDlpPath = mockScript
	platform := NewYtDlpPlatform(cfg, suite.log)

	// Arrange
	req := entities.Request{
		ID:  "test123",
		Url: "https://example.com/test",
	}

	tempDir, err := os.MkdirTemp(suite.tempDir, ytDlpPattern)
	suite.Require().NoError(err)
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Act
	meta, err := platform.fetchMeta(suite.ctx, req, tempDir)

	// Assert
	suite.NoError(err)
	suite.NotNil(meta)

	// Parse expected output to verify
	var expected youtubeMeta
	err = json.Unmarshal([]byte(jsonOutput), &expected)
	suite.Require().NoError(err)

	if expected.Title == "" {
		suite.Equal(expected.Uploader, meta.Title) // Should use uploader as title
	} else {
		suite.Equal(expected.Title, meta.Title)
	}

	if expected.Description == "" {
		suite.Equal("-", meta.Description) // Should use "-" as default
	} else {
		suite.Equal(expected.Description, meta.Description)
	}
}

// TestFetchThumbnail tests the fetchThumbnail method
func (suite *TestYtDlpSuite) TestFetchThumbnail() {
	suite.Run("InvalidURL", func() {
		// Skip if ffmpeg is not available
		if !suite.checkCommand("ffmpeg") {
			suite.T().Skip("ffmpeg not available")
		}

		// Arrange
		req := entities.Request{ID: "test123"}
		thumbUrl := "https://invalid-url.com/image.jpg"

		tempDir, err := os.MkdirTemp(suite.tempDir, ytDlpPattern)
		suite.Require().NoError(err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Act
		fileName, err := suite.platform.fetchThumbnail(suite.ctx, req, thumbUrl, tempDir)

		// Assert
		suite.Error(err)
		suite.Empty(fileName)
		suite.Contains(err.Error(), "ffmpeg command failed")
	})
}

// TestFetchMedia tests the fetchMedia method
func (suite *TestYtDlpSuite) TestFetchMedia() {
	suite.Run("InvalidURL", func() {
		// Skip if yt-dlp is not available
		if !suite.checkCommand("yt-dlp") {
			suite.T().Skip("yt-dlp not available")
		}

		// Arrange
		req := entities.Request{
			ID:             "test123",
			Url:            "https://invalid-url.com/video",
			DownloadFormat: entities.DownloadMp3,
		}

		tempDir, err := os.MkdirTemp(suite.tempDir, ytDlpPattern)
		suite.Require().NoError(err)
		defer func() { _ = os.RemoveAll(tempDir) }()

		// Act
		fileName, size, err := suite.platform.fetchMedia(suite.ctx, req, tempDir)

		// Assert
		suite.Error(err)
		suite.Empty(fileName)
		suite.Zero(size)
		suite.Contains(err.Error(), "yt-dlp command failed")
	})
}

// TestMediaTypes tests the mediaTypes mapping
func (suite *TestYtDlpSuite) TestMediaTypes() {
	// Verify that supported formats are mapped correctly
	mediaType, supported := mediaTypes[entities.DownloadMp3]
	suite.True(supported)
	suite.Equal(entities.MediaMp3, mediaType)

	// Verify that unsupported formats return false
	_, supported = mediaTypes["unsupported"]
	suite.False(supported)
}

// TestConstants tests package constants
func (suite *TestYtDlpSuite) TestConstants() {
	suite.Equal("yt-dlp-*", ytDlpPattern)
	suite.Equal("failed to create temporary directory", errTempDir.Error())
}

// Run the test suite
func TestYtDlp(t *testing.T) {
	suite.Run(t, new(TestYtDlpSuite))
}
