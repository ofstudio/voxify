package domain

import (
	"fmt"
	"log/slog"
)

type Event struct {
	t EventType
	p any
}

func (e Event) Type() EventType {
	return e.t
}

func (e Event) Payload() any {
	return e.p
}

// LogValue implements [slog.LogValuer] interface
func (e Event) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("type", e.t.String()),
		slog.Any("payload", e.p),
	)
}

// Constructors for different event types

// NewDownloadRequestEvent creates a new Event for DownloadRequest.
func NewDownloadRequestEvent(p DownloadRequest) Event { return Event{t: DownloadRequestEvent, p: p} }

// NewDownloadResponseEvent creates a new Event for DownloadResponse.
func NewDownloadResponseEvent(p DownloadResponse) Event { return Event{t: DownloadResponseEvent, p: p} }

// NewBuildRequestEvent creates a new Event for BuildRequest.
func NewBuildRequestEvent(p BuildRequest) Event { return Event{t: BuildRequestEvent, p: p} }

// NewBuildResponseEvent creates a new Event for BuildResponse.
func NewBuildResponseEvent(p BuildResponse) Event { return Event{t: BuildResponseEvent, p: p} }

// NewFeedInfoRequestEvent creates a new Event for FeedInfoRequest.
func NewFeedInfoRequestEvent(p FeedInfoRequest) Event { return Event{t: FeedInfoRequestEvent, p: p} }

// NewFeedInfoResponseEvent creates a new Event for FeedInfoResponse.
func NewFeedInfoResponseEvent(p FeedInfoResponse) Event { return Event{t: FeedInfoResponseEvent, p: p} }

// EventType represents the type of event.
type EventType int

const (
	DownloadRequestEvent EventType = iota + 1 // Start from 1 to avoid zero value
	DownloadResponseEvent
	BuildRequestEvent
	BuildResponseEvent
	FeedInfoRequestEvent
	FeedInfoResponseEvent
)

func (t EventType) String() string {
	switch t {
	case DownloadRequestEvent:
		return "download_request"
	case DownloadResponseEvent:
		return "download_response"
	case BuildRequestEvent:
		return "build_request"
	case BuildResponseEvent:
		return "build_response"
	case FeedInfoRequestEvent:
		return "feed_info_request"
	case FeedInfoResponseEvent:
		return "feed_info_response"
	default:
		return fmt.Sprintf("unknown_type_%d", int(t))
	}
}

// EventHandler is a function that processes an event.
type EventHandler func(event Event)
