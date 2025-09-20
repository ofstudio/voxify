package services

import (
	"errors"
)

var (
	// Business logic errors

	ErrNoMatchingPlatform = NewError(101, "no matching platform found")
	ErrDownloadFailed     = NewError(102, "failed to download episode")
	ErrEpisodeInProgress  = NewError(103, "episode already in progress")
	ErrEpisodeExists      = NewError(104, "episode already exists")
	ErrProcessInterrupted = NewError(105, "process was interrupted")
	ErrEmptyFeed          = NewError(106, "feed has no items")
	ErrInvalidRequest     = NewError(107, "invalid download request")

	// Store errors

	ErrProcessUpsert        = NewError(201, "failed to update process")
	ErrProcessGetByURL      = NewError(202, "failed to get process by URL")
	EpisodeGetByOriginalURL = NewError(203, "failed to get episode by original URL")
	ErrEpisodeCreate        = NewError(204, "failed to create episode")
	ErrProcessGetByStatus   = NewError(205, "failed to get process by status")

	// I/O errors

	ErrDownloadDir = NewError(301, "download directory error")
	ErrMoveFile    = NewError(303, "failed to move file")
)

type Error = struct {
	Code int
	error
}

func NewError(code int, msg string) Error {
	return Error{Code: code, error: errors.New(msg)}
}
