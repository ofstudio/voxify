package entities

import (
	"log/slog"
	"time"
)

// Process represents a media processing task.
type Process struct {
	ID        int64
	Request   Request
	Step      Step
	Status    Status
	Error     error
	Episode   *Episode
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Step is the current step of the processing task.
type Step string

const (
	StepCreating    Step = "creating"
	StepDownloading Step = "downloading"
	StepPublishing  Step = "publishing"
)

// Status is the current status of the processing task.
type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusSuccess    Status = "success"
	StatusFailed     Status = "failed"
)

// LogValue implements slog.LogValuer interface for automatic logging
func (p *Process) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Int64("id", p.ID),
		slog.String("step", string(p.Step)),
		slog.String("status", string(p.Status)),
	}
	if p.Error != nil {
		attrs = append(attrs, slog.String("error", p.Error.Error()))
	}
	attrs = append(attrs, slog.Any("request", p.Request.LogValue()))
	if p.Episode != nil {
		attrs = append(attrs, slog.Any("episode", p.Episode.LogValue()))
	}
	attrs = append(attrs,
		slog.Time("created_at", p.CreatedAt),
		slog.Time("updated_at", p.UpdatedAt),
	)
	return slog.GroupValue(attrs...)
}
