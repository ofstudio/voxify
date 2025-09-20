package config

import (
	"net/url"
	"time"

	"github.com/ofstudio/voxify/internal/entities"
)

type Config struct {
	DB
	Telegram
	Settings
}

// DB is SQLite database configuration
type DB struct {
	Filepath string `env:"DB_FILEPATH,required"` // Path to database file
	Version  uint   // Required database schema version
}

// Telegram bot configuration
type Telegram struct {
	BotToken     string  `env:"TELEGRAM_BOT_TOKEN,required,unset"`
	AllowedUsers []int64 `env:"TELEGRAM_ALLOWED_USERS,required"`
}

// Settings - application settings
type Settings struct {
	PublicUrl       url.URL                 `env:"PUBLIC_URL,required"`      // Public URL where the public directory is accessible (used in feed)
	PublicDir       string                  `env:"PUBLIC_DIR,required"`      // Path to public directory where feed and media files are stored
	DownloadDir     string                  `env:"DOWNLOAD_DIR,required"`    // Path to temporary download directory
	DownloadTimeout time.Duration           `env:"DOWNLOAD_TIMEOUT"`         // Timeout for downloading media
	DownloadFormat  entities.DownloadFormat `env:"DOWNLOAD_FORMAT"`          // Media download format by default (e.g., mp3, m4a)
	DownloadQuality string                  `env:"DOWNLOAD_QUALITY"`         // Media download quality (e.g., 192k)
	ThumbnailSize   int                     `env:"THUMBNAIL_SIZE"`           // Size of the square thumbnail to generate (in pixels)
	YtDlpPath       string                  `env:"YT_DLP_PATH"`              // Path to yt-dlp executable
	FFMpegPath      string                  `env:"FFMPEG_PATH"`              // Path to ffmpeg executable
	FeedFileName    string                  `env:"FEED_FILENAME"`            // Name of the RSS feed file (will be created in PublicDir)
	FeedTitle       string                  `env:"FEED_TITLE"`               // Title of the RSS feed
	FeedDescription string                  `env:"FEED_DESC"`                // Description of the RSS feed
	FeedImage       string                  `env:"FEED_IMAGE,required"`      // URL of the RSS feed cover image
	FeedLanguage    string                  `env:"FEED_LANGUAGE,required"`   // Language of the RSS feed (e.g., en)
	FeedCategories  []string                `env:"FEED_CATEGORIES,required"` // Categories of the RSS feed
	FeedCategories2 []string                `env:"FEED_CATEGORIES2"`         // Additional categories of the RSS feed
	FeedCategories3 []string                `env:"FEED_CATEGORIES3"`         // Additional categories of the RSS feed
	FeedIsExplicit  bool                    `env:"FEED_IS_EXPLICIT"`         // Whether the feed contains explicit content
	FeedAuthor      string                  `env:"FEED_AUTHOR"`              // Author of the RSS feed
	FeedLink        string                  `env:"FEED_LINK"`                // Link to the website of the RSS feed
	FeedKeywords    string                  `env:"FEED_KEYWORDS"`            // Comma-separated keywords for the RSS feed
}
