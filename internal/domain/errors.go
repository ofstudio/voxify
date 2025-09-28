package domain

import (
	"errors"
)

var (
	// Business logic errors

	ErrDownloadPlatform    = NewError(101, "no matching platform found")
	ErrDownloadFailed      = NewError(102, "failed to download episode")
	ErrDownloadInProgress  = NewError(103, "episode already in progress")
	ErrDownloadExists      = NewError(104, "episode already exists")
	ErrDownloadInterrupted = NewError(105, "episode download interrupted")
	ErrDownloadBusy        = NewError(106, "downloader is busy")
	ErrDownloadRequest     = NewError(107, "invalid download request")

	// Store errors

	ErrStoreBegin      = NewError(201, "failed to begin transaction")
	ErrStoreCommit     = NewError(202, "failed to commit transaction")
	ErrStoreRollback   = NewError(203, "failed to rollback transaction")
	ErrEpisodeCreate   = NewError(204, "failed to create episode")
	ErrEpisodeGet      = NewError(205, "failed to get episodes")
	ErrEpisodeCount    = NewError(206, "failed to count episodes")
	ErrEpisodeGetByUrl = NewError(207, "failed to get episode by URL")

	// I/O errors

	ErrFeedSave = NewError(301, "failed to save feed to file")
)

type Error = struct {
	Code int
	error
}

func NewError(code int, msg string) Error {
	return Error{Code: code, error: errors.New(msg)}
}
