package locales

const (
	// Message templates

	MsgStart = `🎧 **Welcome to Voxify Bot!**

I help you convert videos into audio RSS feeds. Simply send me a URL from YouTube, and I'll:

🔽 Download the audio content
🎵 Generate a high-quality audio file
📡 Add it to your personal RSS feed
🔔 Send you a notification when ready

Just paste any video or podcast URL to get started! 

Perfect for creating your own podcast collection or listening to content offline.`

	MsgDownloadStarted = "🔄 Started downloading podcast..."
	MsgDownloadSuccess = "✅ Podcast downloaded successfully!\n\n📖 **%s**"

	// General error messages

	MsgSomethingWentWrong         = "⚠️ Something went wrong while downloading the podcast."
	MsgSomethingWentWrongWithCode = "⚠️ Something went wrong while downloading the podcast (error %d)."

	// Error messages for codes 100-199

	MsgNoMatchingPlatform = "⚠️ This URL is not supported. Please provide a valid video URL."
	MsgDownloadFailed     = "⚠️ Download failed. The media might be unavailable or protected."
	MsgEpisodeInProgress  = "⚠️ This episode is already being processed. Please wait."
	MsgEpisodeExists      = "⚠️ This episode has already been downloaded."
	MsgProcessInterrupted = "⚠️ Download was interrupted. Please try again."

	MsgBuildSuccess = "✅ RSS feed built successfully!"
	MsgBuildError   = "⚠️ Failed to build RSS feed."
)
