package templates

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
	MsgDownloadBusy    = "⏳ Another download is in progress. Please try again later..."
	MsgDownloadSuccess = "✅ Podcast downloaded successfully!\n\n🎧 %s"

	// General error messages

	MsgSomethingWentWrong         = "⚠️ Something went wrong while downloading the podcast."
	MsgSomethingWentWrongWithCode = "⚠️ Something went wrong while downloading the podcast (error %d)."

	// Error messages for codes 100-199

	MsgNoMatchingPlatform = "⚠️ This URL is not supported. Please provide a valid video URL."  // services.ErrNoMatchingPlatform
	MsgDownloadFailed     = "⚠️ Download failed. The media might be unavailable or protected." // services.ErrDownloadFailed
	MsgEpisodeInProgress  = "⚠️ This episode is already being processed. Please wait."         // services.ErrEpisodeInProgress
	MsgEpisodeExists      = "⚠️ This episode has already been downloaded."                     // services.ErrEpisodeExists
	MsgProcessInterrupted = "⚠️ Download was interrupted. Please try again."                   // services.ErrProcessInterrupted
	MsgEmptyFeed          = "⚠️ The feed has no items to process."                             // services.ErrEmptyFeed
	MsgInvalidRequest     = "⚠️ This request is invalid."                                      // services.ErrInvalidRequest

	MsgBuildSuccess = "✅ RSS feed built successfully!"

	MsgFeedInfoBasic      = "📻 Podcast information\n\n<b>%s</b>\n\n%s\n\n"
	MsgFeedInfoAuthor     = "👨‍💻 By %s\n"
	MsgFeedInfoLanguage   = "🌐 Language: %s\n"
	MsgFeedInfoCategories = "📚 Categories: %s\n"
	MsgFeedInfoKeywords   = "🔑 Keywords: %s\n"
	MsgFeedInfoArtwork    = "🖼️ <a href=\"%s\">Artwork</a>\n"
	MsgFeedInfoWebsite    = "🔗 <a href=\"%s\">Website</a>\n"
	MsgFeedInfoEpisodes   = "🎧 Number of episodes: %d\n"
	MsgFeedInfoNoEpisodes = "📭 No episodes yet\n"
	MsgFeedInfoExplicit   = "🔞 Explicit content\n"
	MsgFeedInfoRSS        = "\n📡 RSS: %s"
)
