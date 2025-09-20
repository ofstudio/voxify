package telegram

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ofstudio/voxify/internal/locales"
	"github.com/ofstudio/voxify/internal/services"
)

func TestMsgErr(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "NilError",
			err:      nil,
			expected: "⚠️ Something went wrong while downloading the podcast (error 0).",
		},
		{
			name:     "ProcessorBusyError",
			err:      errProcessorBusy,
			expected: locales.MsgDownloadBusy,
		},
		{
			name:     "NoMatchingPlatformError",
			err:      services.NewError(101, "no matching platform"),
			expected: locales.MsgNoMatchingPlatform,
		},
		{
			name:     "DownloadFailedError",
			err:      services.NewError(102, "download failed"),
			expected: locales.MsgDownloadFailed,
		},
		{
			name:     "EpisodeInProgressError",
			err:      services.NewError(103, "episode in progress"),
			expected: locales.MsgEpisodeInProgress,
		},
		{
			name:     "EpisodeExistsError",
			err:      services.NewError(104, "episode exists"),
			expected: locales.MsgEpisodeExists,
		},
		{
			name:     "ProcessInterruptedError",
			err:      services.NewError(105, "process interrupted"),
			expected: locales.MsgProcessInterrupted,
		},
		{
			name:     "EmptyFeedError",
			err:      services.NewError(106, "empty feed"),
			expected: locales.MsgEmptyFeed,
		},
		{
			name:     "InvalidRequestError",
			err:      services.NewError(107, "invalid request"),
			expected: locales.MsgInvalidRequest,
		},
		{
			name:     "ProcessUpsertError",
			err:      services.NewError(201, "failed to upsert process"),
			expected: "⚠️ Something went wrong while downloading the podcast (error 201).",
		},
		{
			name:     "ProcessGetByURLError",
			err:      services.NewError(202, "failed to get process by URL"),
			expected: "⚠️ Something went wrong while downloading the podcast (error 202).",
		},
		{
			name:     "EpisodeGetByOriginalURLError",
			err:      services.NewError(203, "failed to get episode by original URL"),
			expected: "⚠️ Something went wrong while downloading the podcast (error 203).",
		},
		{
			name:     "EpisodeCreateError",
			err:      services.NewError(204, "failed to create episode"),
			expected: "⚠️ Something went wrong while downloading the podcast (error 204).",
		},
		{
			name:     "ProcessGetByStatusError",
			err:      services.NewError(205, "failed to get process by status"),
			expected: "⚠️ Something went wrong while downloading the podcast (error 205).",
		},
		{
			name:     "UnknownServiceError",
			err:      services.NewError(999, "unknown error"),
			expected: "⚠️ Something went wrong while downloading the podcast (error 999).",
		},
		{
			name:     "GenericError",
			err:      errors.New("generic error"),
			expected: locales.MsgSomethingWentWrong,
		},
		{
			name:     "WrappedServiceError",
			err:      errors.Join(services.NewError(102, "download failed"), errors.New("network error")),
			expected: locales.MsgDownloadFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := msgErr(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
