package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/domain"
)

type EventHandlers struct {
	cfg        config.Settings
	log        *slog.Logger
	bus        domain.EventBus
	builder    domain.FeedBuilder
	downloader domain.EpisodeDownloader
	queue      chan domain.DownloadRequest
	active     sync.Map // map[string]struct{} to track ongoing downloads by URL
}

// NewEventHandlers creates a new EventHandlers handler instance.
func NewEventHandlers(
	cfg config.Settings,
	log *slog.Logger,
	bus domain.EventBus,
	b domain.FeedBuilder,
	d domain.EpisodeDownloader,
) *EventHandlers {
	return &EventHandlers{
		cfg:        cfg,
		log:        log,
		bus:        bus,
		builder:    b,
		downloader: d,
		queue:      make(chan domain.DownloadRequest),
	}
}

func (h *EventHandlers) Start(ctx context.Context) {
	// subscribe to events
	h.bus.Subscribe(domain.DownloadRequestEvent, h.downloadHandler(ctx))
	h.bus.Subscribe(domain.BuildRequestEvent, h.buildHandler(ctx))
	// logging handlers
	h.bus.Subscribe(domain.DownloadRequestEvent, h.logDownloadRequestHandler)
	h.bus.Subscribe(domain.DownloadResponseEvent, h.logDownloadResponseHandler)
	h.bus.Subscribe(domain.BuildRequestEvent, h.logBuildRequestHandler)
	h.bus.Subscribe(domain.BuildResponseEvent, h.logBuildResponseHandler)

	// start workers
	for i := 0; i < h.cfg.DownloadWorkers; i++ {
		go h.downloadWorker(ctx, i)
	}
}

// downloadHandler returns event handler for handling episode download requests.
func (h *EventHandlers) downloadHandler(ctx context.Context) domain.EventHandler {
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
func (h *EventHandlers) downloadValidate(ctx context.Context, req domain.DownloadRequest) error {
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

// downloadWorker processes download requests from the queue.
func (h *EventHandlers) downloadWorker(ctx context.Context, id int) {
	h.log.Info("[event handlers] download worker started", "id", id)
	defer h.log.Info("[event handlers] download worker stopped", "id", id)

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
func (h *EventHandlers) download(ctx context.Context, req domain.DownloadRequest) {
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
		h.failRequest(req, err)
		return
	}

	// publish success response
	h.bus.Publish(domain.NewDownloadResponseEvent(domain.DownloadResponse{
		Status:  domain.StatusSuccess,
		Episode: episode,
		Request: req,
	}))
}

// buildHandler returns event handler for handling feed build requests.
func (h *EventHandlers) buildHandler(ctx context.Context) domain.EventHandler {
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

// failRequest publishes a failed response event based on the request type.
func (h *EventHandlers) failRequest(req any, err error) {
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
		h.log.Error("[event handlers] unknown request type", "req", req)
		return
	}
	h.bus.Publish(event)
}

// logging handlers

// logDownloadRequestHandler logs download request events.
func (h *EventHandlers) logDownloadRequestHandler(event domain.Event) {
	req := event.Payload().(domain.DownloadRequest)
	h.log.Info("[event handlers] download request", "request", req)
}

// logDownloadResponseHandler logs download response events.
func (h *EventHandlers) logDownloadResponseHandler(event domain.Event) {
	resp := event.Payload().(domain.DownloadResponse)
	h.log.Info("[event handlers] download response", "response", resp)
}

// logBuildRequestHandler logs build request events.
func (h *EventHandlers) logBuildRequestHandler(event domain.Event) {
	req := event.Payload().(domain.BuildRequest)
	h.log.Info("[event handlers] build request", "request", req)
}

// logBuildResponseHandler logs build response events.
func (h *EventHandlers) logBuildResponseHandler(event domain.Event) {
	resp := event.Payload().(domain.BuildResponse)
	h.log.Info("[event handlers] build response", "response", resp)
}
