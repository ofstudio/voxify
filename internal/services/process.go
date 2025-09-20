package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ofstudio/voxify/internal/entities"
	"github.com/ofstudio/voxify/pkg/randtoken"
)

const (
	notifyBuffer   = 8 // Buffer size for the notify channel
	workerPoolSize = 2 // Number of concurrent workers for processing requests
)

// ProcessService handles the processing of download requests.
type ProcessService struct {
	log        *slog.Logger
	store      Store
	downloader Downloader
	builder    Builder
	in         chan *entities.Request
	notify     chan *entities.Process
}

// NewProcessService creates a new ProcessService instance.
func NewProcessService(log *slog.Logger, store Store, downloader Downloader, builder Builder) *ProcessService {
	return &ProcessService{
		log:        log,
		store:      store,
		downloader: downloader,
		builder:    builder,
		in:         make(chan *entities.Request),
		notify:     make(chan *entities.Process, notifyBuffer),
	}
}

// In returns the input channel for receiving download requests.
func (s *ProcessService) In() chan<- *entities.Request {
	return s.in
}

// Notify returns the output channel for sending notifications about process completion.
func (s *ProcessService) Notify() <-chan *entities.Process {
	return s.notify
}

// Init initializes the service before starting.
// It fails all in-progress processes to ensure a clean state.
func (s *ProcessService) Init(ctx context.Context) error {
	// Fail all in-progress processes
	processes, err := s.store.ProcessGetByStatus(ctx, entities.StatusInProgress)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProcessGetByStatus, err)
	}
	if len(processes) > 0 {
		s.log.Info("[process service] failing in-progress processes", "count", len(processes))
		for _, process := range processes {
			process.Status = entities.StatusFailed
			process.Error = ErrProcessInterrupted
			if err = s.store.ProcessUpsert(ctx, process); err != nil {
				return fmt.Errorf("%w: %w", ErrProcessUpsert, err)
			}
			s.log.Info("[process service] process failed", "process", process.LogValue())
			s.sendNotify(ctx, process)
		}
	}
	return nil
}

// Start begins processing download requests.
// It runs until the provided context is canceled.
func (s *ProcessService) Start(ctx context.Context) {
	// Start multiple workers for better concurrency
	for i := 0; i < workerPoolSize; i++ {
		go s.worker(ctx, i)
	}
}

// worker processes requests from the input channel
func (s *ProcessService) worker(ctx context.Context, workerID int) {
	s.log.Info("[process service] worker started", "worker_id", workerID)
	defer s.log.Info("[process service] worker stopped", "worker_id", workerID)

	for {
		select {
		case <-ctx.Done():
			return
		case req := <-s.in:
			req.ID = randtoken.New(10)
			s.log.Info("[process service] received new request", "worker_id", workerID, "request", req.LogValue())
			s.handle(ctx, req)
		}
	}
}

// handle processes a single download request.
// It updates the process status at each step and handles errors appropriately.
func (s *ProcessService) handle(ctx context.Context, req *entities.Request) {
	var err error
	process := &entities.Process{
		Request: *req,
		Step:    entities.StepCreating,
		Status:  entities.StatusInProgress,
	}

	// Create process
	if err = s.update(ctx, process); err != nil {
		s.fail(ctx, process, err)
		return
	}

	// Validate process
	if err = s.validate(ctx, process); err != nil {
		s.fail(ctx, process, err)
		return
	}

	// Download episode
	process.Step = entities.StepDownloading
	if err = s.update(ctx, process); err != nil {
		s.fail(ctx, process, err)
		return
	}

	if process.Episode, err = s.downloader.Download(ctx, process.Request); err != nil {
		s.fail(ctx, process, err)
		return
	}
	s.log.Info("[process service] episode downloaded", "process", process.LogValue())

	// Build podcast feed
	process.Step = entities.StepPublishing
	if err = s.update(ctx, process); err != nil {
		s.fail(ctx, process, err)
		return
	}
	if err = s.builder.Build(ctx); err != nil {
		s.fail(ctx, process, err)
		return
	}
	process.Status = entities.StatusSuccess
	if err = s.update(ctx, process); err != nil {
		s.fail(ctx, process, err)
		return
	}
}

// validate checks if the process can proceed.
// It checks for existing processes in progress and existing episodes.
// If the process is valid, it returns nil. Otherwise, it returns an appropriate error.
func (s *ProcessService) validate(ctx context.Context, process *entities.Process) error {
	// Init processes in progress
	count, err := s.store.ProcessCountByUrlAndStatus(ctx, process.Request.Url, entities.StatusInProgress)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrProcessGetByURL, err)
	}
	if count > 1 { // count > 1 because the current process is also counted
		return ErrEpisodeInProgress
	}

	// Init existing episodes
	if process.Request.Force {
		return nil
	}
	existing, err := s.store.EpisodeGetByOriginalUrl(ctx, process.Request.Url)
	if err != nil {
		return fmt.Errorf("%w: %w", EpisodeGetByOriginalURL, err)
	}
	if len(existing) > 0 {
		return ErrEpisodeExists
	}

	return nil
}

// sendNotify sends a notification.
func (s *ProcessService) sendNotify(ctx context.Context, process *entities.Process) {
	select {
	case s.notify <- process: // Successfully sent
	case <-ctx.Done(): // Context canceled
		s.log.Error("[process service] notification dropped due to context cancellation",
			"process", process.LogValue())
	default: // Channel full
		s.log.Error("[process service] notification buffer full, dropping notification",
			"process", process.LogValue())
	}
}

// fail marks the process as failed with the given error and notifies via the notify channel.
func (s *ProcessService) fail(ctx context.Context, process *entities.Process, e error) {
	process.Status = entities.StatusFailed
	process.Error = e
	if err := s.store.ProcessUpsert(ctx, process); err != nil {
		s.log.Error("[process service] failed to update process",
			"error", err, "process", process.LogValue())
	}
	s.log.Error("[process service] process failed",
		"process", process.LogValue())

	// Notify about process failure
	s.sendNotify(ctx, process)
}

// update creates or updates the process in the store and notifies via the notify channel.
func (s *ProcessService) update(ctx context.Context, process *entities.Process) error {
	var created bool
	if process.ID == 0 {
		created = true
	}
	if err := s.store.ProcessUpsert(ctx, process); err != nil {
		return fmt.Errorf("%w: %w", ErrProcessUpsert, err)
	}

	if created {
		s.log.Info("[process service] process created",
			"process", process.LogValue())
	} else {
		s.log.Info("[process service] process updated",
			"process", process.LogValue())
	}

	// Notify about process update
	s.sendNotify(ctx, process)
	return nil
}
