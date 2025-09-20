package entities

import (
	"log/slog"
	"time"

	"github.com/ofstudio/voxify/pkg/feedcast"
)

// Episode represents a podcast episode.
type Episode struct {
	ID            int64
	Title         string
	Description   string
	ThumbnailFile string
	MediaFile     string
	MediaType     MediaType
	MediaDuration int64
	MediaSize     int64
	Author        string
	OriginalURL   string
	CanonicalURL  string
	CreatedAt     time.Time
}

// MediaType is the MIME type of the media file.
type MediaType = feedcast.EnclosureType

const (
	MediaMp3 = feedcast.Mp3
	MediaM4a = feedcast.M4a
)

// LogValue implements slog.LogValuer interface for automatic logging
func (p *Episode) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int64("id", p.ID),
		slog.String("thumbnail_filename", p.ThumbnailFile),
		slog.String("media_filename", p.MediaFile),
		slog.Int64("media_duration", p.MediaDuration),
		slog.String("media_type", string(p.MediaType)),
		slog.Int64("media_size", p.MediaSize),
		slog.String("original_url", p.OriginalURL),
		slog.String("canonical_url", p.CanonicalURL),
		slog.Time("created_at", p.CreatedAt),
	)
}
