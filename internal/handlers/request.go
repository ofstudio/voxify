package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/domain"
)

// RequestHandlers handles requests for downloading episodes and building feeds.
type RequestHandlers struct {
	cfg        config.Settings
	log        *slog.Logger
	bus        domain.EventBus
	builder    domain.FeedBuilder
	downloader domain.EpisodeDownloader
	queue      chan domain.DownloadRequest
	active     sync.Map
	wg         sync.WaitGroup
}

// NewRequestHandlers creates a new RequestHandlers handler instance.
func NewRequestHandlers(
	cfg config.Settings,
	log *slog.Logger,
	bus domain.EventBus,
) *RequestHandlers {
	return &RequestHandlers{
		cfg:   cfg,
		log:   log,
		bus:   bus,
		queue: make(chan domain.DownloadRequest),
	}
}

// WithBuilder sets the FeedBuilder instance.
func (h *RequestHandlers) WithBuilder(b domain.FeedBuilder) *RequestHandlers {
	h.builder = b
	return h
}

// WithDownloader sets the EpisodeDownloader instance.
func (h *RequestHandlers) WithDownloader(d domain.EpisodeDownloader) *RequestHandlers {
	h.downloader = d
	return h
}

// Init initializes event handlers and starts workers.
func (h *RequestHandlers) Init(ctx context.Context) error {
	// check dependencies
	if h.builder == nil {
		return errors.New("feed builder is not set")
	}
	if h.downloader == nil {
		return errors.New("episode downloader is not set")
	}
	if h.bus == nil {
		return errors.New("event bus is not set")
	}

	// subscribe to events
	h.bus.Subscribe(domain.DownloadRequestEvent, h.downloadHandler(ctx))
	h.bus.Subscribe(domain.BuildRequestEvent, h.buildHandler(ctx))
	h.bus.Subscribe(domain.FeedInfoRequestEvent, h.feedInfoHandler(ctx))

	// subscribe logging handlers
	h.bus.Subscribe(domain.DownloadRequestEvent, h.logDownloadRequestHandler)
	h.bus.Subscribe(domain.DownloadResponseEvent, h.logDownloadResponseHandler)
	h.bus.Subscribe(domain.BuildRequestEvent, h.logBuildRequestHandler)
	h.bus.Subscribe(domain.BuildResponseEvent, h.logBuildResponseHandler)
	h.bus.Subscribe(domain.FeedInfoRequestEvent, h.logFeedInfoRequestHandler)
	h.bus.Subscribe(domain.FeedInfoResponseEvent, h.logFeedInfoResponseHandler)

	// channel to synchronize worker startup
	ready := make(chan struct{})

	// start download workers
	for i := 0; i < h.cfg.DownloadWorkers; i++ {
		h.wg.Add(1)
		go h.downloadWorker(ctx, i, ready)
	}

	// wait until all workers signal they are ready
	for i := 0; i < h.cfg.DownloadWorkers; i++ {
		<-ready
	}
	close(ready)

	return nil
}

// Wait waits for all workers to finish.
func (h *RequestHandlers) Wait() {
	h.wg.Wait()
}

// downloadWorker processes download requests from the queue.
func (h *RequestHandlers) downloadWorker(ctx context.Context, id int, ready chan<- struct{}) {
	h.log.Info("[request handlers] download worker started", "id", id)
	defer h.log.Info("[request handlers] download worker stopped", "id", id)
	defer h.wg.Done()

	// signal that the worker is ready to process requests
	ready <- struct{}{}

	for {
		select {
		case req := <-h.queue:
			h.download(ctx, req)
		case <-ctx.Done():
			return
		}
	}
}

// download processes a single episode download request.
func (h *RequestHandlers) download(ctx context.Context, req domain.DownloadRequest) {
	// mark URL as being downloaded
	if _, exists := h.active.LoadOrStore(req.Url, struct{}{}); exists {
		// already downloading
		h.failRequest(req, domain.ErrDownloadInProgress)
		return
	}
	defer h.active.Delete(req.Url)

	// perform download
	episode, err := h.downloader.Download(ctx, req)
	if err != nil {
		// check if context was cancelled
		if ctx.Err() != nil {
			err = fmt.Errorf("%w: %w; %w", domain.ErrDownloadInterrupted, ctx.Err(), err)
		}

		h.failRequest(req, err)
		return
	}

	// publish success response
	h.bus.Publish(domain.NewDownloadResponseEvent(domain.DownloadResponse{
		Status:  domain.StatusSuccess,
		Episode: episode,
		Request: req,
	}))

	// publish build request
	h.bus.Publish(domain.NewBuildRequestEvent(domain.BuildRequest{ID: req.ID}))
}

// downloadHandler returns event handler for handling episode download requests.
func (h *RequestHandlers) downloadHandler(ctx context.Context) domain.EventHandler {
	return func(event domain.Event) {
		req := event.Payload().(domain.DownloadRequest)

		// validate request
		if err := h.downloadValidate(ctx, req); err != nil {
			h.failRequest(req, err)
			return
		}

		// enqueue request
		select {
		case h.queue <- req:
			// successfully enqueued
			h.bus.Publish(domain.NewDownloadResponseEvent(domain.DownloadResponse{
				Status:  domain.StatusPending,
				Request: req,
			}))
		default:
			// queue is full
			h.failRequest(req, domain.ErrDownloadBusy)
		}
	}
}

// downloadValidate performs  validation of the download request.
func (h *RequestHandlers) downloadValidate(ctx context.Context, req domain.DownloadRequest) error {
	// check if URL is already being downloaded
	if _, exists := h.active.Load(req.Url); exists {
		return domain.ErrDownloadInProgress
	}
	// validate using downloader
	if err := h.downloader.Validate(ctx, req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil

}

// buildHandler returns event handler for handling feed build requests.
func (h *RequestHandlers) buildHandler(ctx context.Context) domain.EventHandler {
	return func(event domain.Event) {
		req := event.Payload().(domain.BuildRequest)
		if err := h.builder.Build(ctx); err != nil {
			h.failRequest(req, err)
			return
		}
		h.bus.Publish(domain.NewBuildResponseEvent(domain.BuildResponse{
			Status:  domain.StatusSuccess,
			Request: req,
		}))
	}
}

// feedInfoHandler returns event handler for handling feed info requests.
func (h *RequestHandlers) feedInfoHandler(ctx context.Context) domain.EventHandler {
	return func(event domain.Event) {
		req := event.Payload().(domain.FeedInfoRequest)
		info, err := h.builder.Info(ctx)
		if err != nil {
			h.bus.Publish(domain.NewFeedInfoResponseEvent(domain.FeedInfoResponse{
				Status:  domain.StatusFailed,
				Error:   err,
				Request: req,
			}))
			return
		}
		h.bus.Publish(domain.NewFeedInfoResponseEvent(domain.FeedInfoResponse{
			Status:   domain.StatusSuccess,
			FeedInfo: info,
			Request:  req,
		}))
	}
}

// failRequest publishes a failed response event based on the request type.
func (h *RequestHandlers) failRequest(req any, err error) {
	var event domain.Event
	switch r := req.(type) {
	case domain.DownloadRequest:
		event = domain.NewDownloadResponseEvent(domain.DownloadResponse{
			Status:  domain.StatusFailed,
			Error:   err,
			Request: r,
		})
	case domain.BuildRequest:
		event = domain.NewBuildResponseEvent(domain.BuildResponse{
			Status:  domain.StatusFailed,
			Error:   err,
			Request: r,
		})
	default:
		h.log.Error("[request handlers] unknown request type", "req", req)
		return
	}
	h.bus.Publish(event)
}

// logging handlers

// logDownloadRequestHandler logs download request events.
func (h *RequestHandlers) logDownloadRequestHandler(event domain.Event) {
	req := event.Payload().(domain.DownloadRequest)
	h.log.Info("[request handlers] download request", "request", req)
}

// logDownloadResponseHandler logs download response events.
func (h *RequestHandlers) logDownloadResponseHandler(event domain.Event) {
	resp := event.Payload().(domain.DownloadResponse)
	h.log.Info("[request handlers] download response", "response", resp)
}

// logBuildRequestHandler logs build request events.
func (h *RequestHandlers) logBuildRequestHandler(event domain.Event) {
	req := event.Payload().(domain.BuildRequest)
	h.log.Info("[request handlers] build request", "request", req)
}

// logBuildResponseHandler logs build response events.
func (h *RequestHandlers) logBuildResponseHandler(event domain.Event) {
	resp := event.Payload().(domain.BuildResponse)
	h.log.Info("[request handlers] build response", "response", resp)
}

// logFeedInfoRequestHandler logs feed info request events.
func (h *RequestHandlers) logFeedInfoRequestHandler(event domain.Event) {
	req := event.Payload().(domain.FeedInfoRequest)
	h.log.Info("[request handlers] feed info request", "request", req)
}

// logFeedInfoResponseHandler logs feed info response events.
func (h *RequestHandlers) logFeedInfoResponseHandler(event domain.Event) {
	resp := event.Payload().(domain.FeedInfoResponse)
	h.log.Info("[request handlers] feed info response", "response", resp)
}
