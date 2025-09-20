package entities

import "log/slog"

// Request represents a download request.
type Request struct {
	ID              string
	UserID          int64
	ChatID          int64
	MessageID       int
	Url             string
	DownloadFormat  DownloadFormat
	DownloadQuality string
	Force           bool
}

// DownloadFormat is the format in which media should be downloaded.
type DownloadFormat string

const (
	DownloadMp3 DownloadFormat = "mp3"
)

// LogValue implements slog.LogValuer interface for automatic logging
func (p *Request) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", p.ID),
		slog.Int64("user_id", p.UserID),
		slog.Int64("chat_id", p.ChatID),
		slog.Int("message_id", p.MessageID),
		slog.String("url", p.Url),
	}
	if p.DownloadFormat != "" {
		attrs = append(attrs, slog.String("download_format", string(p.DownloadFormat)))
	}
	if p.DownloadQuality != "" {
		attrs = append(attrs, slog.String("download_quality", p.DownloadQuality))
	}
	if p.Force {
		attrs = append(attrs, slog.Bool("force", p.Force))
	}
	return slog.GroupValue(attrs...)
}
