package templates

import (
	"errors"
	"fmt"

	"github.com/ofstudio/voxify/internal/domain"
)

const (
	// Message templates

	MsgStart = `🎧 <b>Welcome to Voxify Bot!</b>

I help you convert videos into audio RSS feeds. Simply send me a URL from YouTube, and I'll:

🔽 Download the audio content
🎵 Generate a high-quality audio file
📡 Add it to your personal RSS feed
🔔 Send you a notification when ready

Just paste any video or podcast URL to get started! 

Perfect for creating your own podcast collection or listening to content offline.`

	MsgDownloadStarted = "🔄 Started downloading podcast..."
	msgDownloadSuccess = "✅ Podcast downloaded successfully!\n\n🎧 %s"
	MsgBuildSuccess    = "✅ RSS feed built successfully!"

	// Error messages

	MsgSomethingWentWrong = "⚠️ Something went wrong."
	MsgUnknownError       = "⚠️ An unknown error occurred."

	MsgBuildFeedFailed     = "⚠️ Failed to build the RSS feed."
	MsgBuildLandingFailed  = "⚠️ Failed to build the landing page."
	MsgDownloadFailed      = "⚠️ Download failed. The media might be unavailable or protected."
	MsgDownloadInProgress  = "⚠️ This episode is already being processed. Please wait."
	MsgDownloadExists      = "⚠️ This episode has already been downloaded."
	MsgDownloadInterrupted = "⚠️ Download was interrupted. Please try again."
	MsgNoMatchingPlatform  = "⚠️ This URL is not supported. Please provide a valid video URL."
	MsgDownloadBusy        = "⏳ Another download is in progress. Please try again later..."
	MsgInvalidRequest      = "⚠️ This request is invalid."
)

// MsgDownloadSuccess formats the download success message with the episode title.
func MsgDownloadSuccess(episodeTitle string) string {
	return fmt.Sprintf(msgDownloadSuccess, episodeTitle)
}

// MsgError maps an error to a user-friendly message.
func MsgError(err error) string {
	switch {
	case err == nil:
		return MsgUnknownError
	case errors.Is(err, domain.ErrBuildFeedFailed):
		return MsgBuildFeedFailed
	case errors.Is(err, domain.ErrBuildLandingFailed):
		return MsgBuildLandingFailed
	case errors.Is(err, domain.ErrDownloadFailed):
		return MsgDownloadFailed
	case errors.Is(err, domain.ErrDownloadInProgress):
		return MsgDownloadInProgress
	case errors.Is(err, domain.ErrDownloadExists):
		return MsgDownloadExists
	case errors.Is(err, domain.ErrDownloadInterrupted):
		return MsgDownloadInterrupted
	case errors.Is(err, domain.ErrDownloadBusy):
		return MsgDownloadBusy
	case errors.Is(err, domain.ErrNoMatchingPlatform):
		return MsgNoMatchingPlatform
	case
		errors.Is(err, domain.ErrInvalidUrl),
		errors.Is(err, domain.ErrInvalidFormat),
		errors.Is(err, domain.ErrDownloadQuality):
		return MsgInvalidRequest
	default:
		return MsgSomethingWentWrong
	}
}
