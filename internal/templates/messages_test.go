package templates

import (
	"errors"
	"testing"

	"github.com/ofstudio/voxify/internal/domain"
	"github.com/stretchr/testify/suite"
)

// TestMessages runs the test suite
func TestMessages(t *testing.T) {
	suite.Run(t, new(TestMessagesSuite))
}

// TestMessagesSuite defines the test suite for message templates
type TestMessagesSuite struct {
	suite.Suite
}

// TestMsgDownloadSuccess tests the MsgDownloadSuccess function
func (suite *TestMessagesSuite) TestMsgDownloadSuccess() {
	suite.Run("simple title", func() {
		result := MsgDownloadSuccess("My Podcast Episode")
		suite.Contains(result, "✅ Podcast downloaded successfully!")
		suite.Contains(result, "🎧 My Podcast Episode")
	})

	suite.Run("title with HTML special characters", func() {
		result := MsgDownloadSuccess("<script>alert('xss')</script>")
		suite.Contains(result, "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;")
		suite.NotContains(result, "<script>")
	})

	suite.Run("title with ampersand", func() {
		result := MsgDownloadSuccess("Rock & Roll")
		suite.Contains(result, "Rock &amp; Roll")
	})

	suite.Run("title with quotes", func() {
		result := MsgDownloadSuccess(`The "Best" Episode`)
		suite.Contains(result, "&#34;Best&#34;")
	})

	suite.Run("empty title", func() {
		result := MsgDownloadSuccess("")
		suite.Contains(result, "✅ Podcast downloaded successfully!")
		suite.Contains(result, "🎧")
	})

	suite.Run("unicode characters", func() {
		result := MsgDownloadSuccess("Привет мир 🎵")
		suite.Contains(result, "Привет мир 🎵")
	})
}

// TestMsgError tests the MsgError function
func (suite *TestMessagesSuite) TestMsgError() {
	suite.Run("nil error", func() {
		result := MsgError(nil)
		suite.Equal(MsgUnknownError, result)
	})

	suite.Run("download interrupted error", func() {
		result := MsgError(domain.ErrDownloadInterrupted)
		suite.Equal(MsgDownloadInterrupted, result)
	})

	suite.Run("download failed error", func() {
		result := MsgError(domain.ErrDownloadFailed)
		suite.Equal(MsgDownloadFailed, result)
	})

	suite.Run("download in progress error", func() {
		result := MsgError(domain.ErrDownloadInProgress)
		suite.Equal(MsgDownloadInProgress, result)
	})

	suite.Run("download exists error", func() {
		result := MsgError(domain.ErrDownloadExists)
		suite.Equal(MsgDownloadExists, result)
	})

	suite.Run("download busy error", func() {
		result := MsgError(domain.ErrDownloadBusy)
		suite.Equal(MsgDownloadBusy, result)
	})

	suite.Run("no matching platform error", func() {
		result := MsgError(domain.ErrNoMatchingPlatform)
		suite.Equal(MsgNoMatchingPlatform, result)
	})

	suite.Run("invalid URL error", func() {
		result := MsgError(domain.ErrInvalidUrl)
		suite.Equal(MsgInvalidRequest, result)
	})

	suite.Run("invalid format error", func() {
		result := MsgError(domain.ErrInvalidFormat)
		suite.Equal(MsgInvalidRequest, result)
	})

	suite.Run("download quality error", func() {
		result := MsgError(domain.ErrDownloadQuality)
		suite.Equal(MsgInvalidRequest, result)
	})

	suite.Run("build feed failed error", func() {
		result := MsgError(domain.ErrBuildFeedFailed)
		suite.Equal(MsgBuildFeedFailed, result)
	})

	suite.Run("build landing failed error", func() {
		result := MsgError(domain.ErrBuildLandingFailed)
		suite.Equal(MsgBuildLandingFailed, result)
	})

	suite.Run("wrapped domain error", func() {
		wrappedErr := errors.Join(domain.ErrDownloadFailed, errors.New("additional context"))
		result := MsgError(wrappedErr)
		suite.Equal(MsgDownloadFailed, result)
	})

	suite.Run("unknown error", func() {
		unknownErr := errors.New("some random error")
		result := MsgError(unknownErr)
		suite.Equal(MsgSomethingWentWrong, result)
	})
}
