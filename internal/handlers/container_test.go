package handlers

import (
	"context"
	"log/slog"
	"testing"

	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestContainerInitUsesAppContextForNotifications(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := mocks.NewMockEventBus(t)
	botMock := mocks.NewMockBot(t)
	builder := mocks.NewMockFeedBuilder(t)
	downloader := mocks.NewMockEpisodeDownloader(t)

	bus.On("Subscribe", mock.AnythingOfType("domain.EventType"), mock.AnythingOfType("domain.EventHandler")).
		Return().
		Times(12)
	botMock.On("RegisterHandler",
		mock.AnythingOfType("bot.HandlerType"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("bot.MatchType"),
		mock.Anything,
	).Return("handler-id").Times(4)

	handlers := New(config.Settings{DownloadWorkers: 0}, slog.Default(), bus).
		WithBuilder(builder).
		WithDownloader(downloader).
		WithBot(botMock)

	err := handlers.Init(ctx)
	require.NoError(t, err)
	require.Equal(t, ctx, handlers.Notification.ctx)

	cancel()
	require.ErrorIs(t, handlers.Notification.ctx.Err(), context.Canceled)

	bus.AssertExpectations(t)
	botMock.AssertExpectations(t)
}
