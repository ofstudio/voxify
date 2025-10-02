package domain

import (
	"fmt"
	"log/slog"
)

// DownloadResponse represents the response for a DownloadRequest.
type DownloadResponse struct {
	Status  ResponseStatus
	Error   error
	Episode *Episode
	Request DownloadRequest
}

// LogValue implements [slog.LogValuer] interface
func (r DownloadResponse) LogValue() slog.Value {
	attrs := []slog.Attr{slog.String("status", r.Status.String())}
	if r.Error != nil {
		attrs = append(attrs, slog.String("error", r.Error.Error()))
	}
	attrs = append(attrs, slog.Any("request", r.Request))
	if r.Episode != nil {
		attrs = append(attrs, slog.Any("episode", r.Episode))
	}
	attrs = append(attrs, slog.Any("request", r.Request))
	return slog.GroupValue(attrs...)
}

// BuildResponse represents the response for a BuildRequest.
type BuildResponse struct {
	Status  ResponseStatus
	Error   error
	Request BuildRequest
}

// LogValue implements [slog.LogValuer] interface
func (r BuildResponse) LogValue() slog.Value {
	attrs := []slog.Attr{slog.String("status", r.Status.String())}
	if r.Error != nil {
		attrs = append(attrs, slog.String("error", r.Error.Error()))
	}
	attrs = append(attrs, slog.Any("request", r.Request))
	return slog.GroupValue(attrs...)
}

type FeedInfoResponse struct {
	Status   ResponseStatus
	Error    error
	FeedInfo *FeedInfo
	Request  FeedInfoRequest
}

// LogValue implements [slog.LogValuer] interface
func (r FeedInfoResponse) LogValue() slog.Value {
	attrs := []slog.Attr{slog.String("status", r.Status.String())}
	if r.Error != nil {
		attrs = append(attrs, slog.String("error", r.Error.Error()))
	}
	if r.FeedInfo != nil {
		attrs = append(attrs, slog.Any("feed", r.FeedInfo))
	}
	attrs = append(attrs, slog.Any("request", r.Request))
	return slog.GroupValue(attrs...)
}

// ResponseStatus is the current status of the Response.
type ResponseStatus int

const (
	StatusPending ResponseStatus = iota // Being processed
	StatusSuccess                       // Successfully processed (final state)
	StatusFailed                        // Processing failed (final state)
)

func (s ResponseStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusSuccess:
		return "success"
	case StatusFailed:
		return "failed"
	default:
		return fmt.Sprintf("unknown_status_%d", int(s))
	}
}
