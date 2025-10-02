package domain

import "errors"

// Domain errors
var (
	ErrBuildFeedFailed     = errors.New("failed to build feed")
	ErrBuildLandingFailed  = errors.New("failed to build landing page")
	ErrDownloadFailed      = errors.New("failed to download episode")
	ErrDownloadInProgress  = errors.New("episode already in progress")
	ErrDownloadExists      = errors.New("episode already exists")
	ErrDownloadInterrupted = errors.New("episode download interrupted")
	ErrDownloadBusy        = errors.New("downloader is busy")
	ErrNoMatchingPlatform  = errors.New("no matching platform found")
	ErrInvalidUrl          = errors.New("invalid download URL")
	ErrInvalidFormat       = errors.New("unsupported download format")
	ErrDownloadQuality     = errors.New("unsupported download quality")
)
