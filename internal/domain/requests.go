package domain

import "log/slog"

// DownloadRequest represents a request to download episode from a URL.
type DownloadRequest struct {
	ID              string         // Unique request id
	Source          RequestSource  // User who made the request
	Url             string         // URL to download
	DownloadFormat  DownloadFormat // Format to download (e.g., mp3, m4a)
	DownloadQuality string         // Quality of the download (e.g., 128kbps)
}

// LogValue implements [slog.LogValuer] interface
func (r DownloadRequest) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", r.ID),
		slog.Any("source", r.Source),
		slog.String("url", r.Url),
		slog.String("download_format", string(r.DownloadFormat)),
		slog.String("download_quality", r.DownloadQuality),
	)
}

// BuildRequest represents a request to build the podcast feed and landing page.
type BuildRequest struct {
	ID     string         // Unique request id
	Source *RequestSource // User who made the request
}

// LogValue implements [slog.LogValuer] interface
func (r BuildRequest) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", r.ID),
	}
	if r.Source != nil {
		attrs = append(attrs, slog.Any("source", *r.Source))
	}
	return slog.GroupValue(attrs...)
}

// RequestSource contains data about user who made the RequestType.
type RequestSource struct {
	UserID    int64 // Telegram user ID
	ChatID    int64 // Telegram chat ID
	MessageID int   // Telegram message ID
}

// LogValue implements [slog.LogValuer] interface
func (r RequestSource) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int64("user_id", r.UserID),
		slog.Int64("chat_id", r.ChatID),
		slog.Int("message_id", r.MessageID),
	)
}

// DownloadFormat is the format in which media should be downloaded (e.g., mp3, m4a) via DownloadRequest.
type DownloadFormat string

const (
	DownloadMp3 DownloadFormat = "mp3"
	DownloadM4a DownloadFormat = "m4a"
)
