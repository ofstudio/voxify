package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/domain"
	"github.com/ofstudio/voxify/internal/events"
	"github.com/ofstudio/voxify/internal/handlers"
	"github.com/ofstudio/voxify/internal/platforms"
	"github.com/ofstudio/voxify/internal/services"
	"github.com/ofstudio/voxify/internal/store"
	"github.com/ofstudio/voxify/internal/templates"
	"github.com/ofstudio/voxify/pkg/telegram"
)

// App represents the main application.
type App struct {
	cfg      config.Config
	log      *slog.Logger
	store    domain.Store
	bus      domain.EventBus
	bot      telegram.Bot
	services *services.Container
	handlers *handlers.Container
}

// New creates a new App instance.
func New(cfg config.Config, log *slog.Logger) *App {
	return &App{cfg: cfg, log: log}
}

// Start initializes and starts the application.
func (app *App) Start(ctx context.Context) error {
	// Initialize app components
	if err := app.init(ctx); err != nil {
		return fmt.Errorf("app initialization failed: %w", err)
	}

	//start the bot
	go func() {
		app.bot.Start(ctx)
		app.log.Info("[app] telegram bot stopped")
	}()

	app.log.Info("[app] telegram bot started", "bot_user_id", app.bot.ID())
	app.log.Info("[app] application started")

	// Wait for the shutdown signal
	<-ctx.Done()

	// Shutdown components
	app.handlers.Wait()
	app.bus.Wait()
	app.store.Close()

	app.log.Info("[app] application stopped")
	return nil
}

// init initializes the application components.
func (app *App) init(ctx context.Context) error {

	var err error
	// Initialize templates
	if err = templates.Init(ctx); err != nil {
		return fmt.Errorf("templates initialization failed: %w", err)
	}

	// Create event bus
	app.bus = events.NewAsyncBus(app.log)

	// Initialize store
	if app.store, err = app.initStore(); err != nil {
		return fmt.Errorf("store initialization failed: %w", err)
	}

	// Initialize services
	if app.services, err = app.initServices(ctx); err != nil {
		return fmt.Errorf("services initialization failed: %w", err)
	}

	// Create handlers
	app.handlers = handlers.New(app.cfg.Settings, app.log, app.bus).
		WithBuilder(app.services.Feed).
		WithDownloader(app.services.Episode)

	// Create bot
	app.bot, err = telegram.NewBot(
		app.cfg.Telegram.BotToken,
		bot.WithAllowedUpdates(bot.AllowedUpdates{"message"}),
		bot.WithErrorsHandler(app.handlers.Telegram.ErrorsHandler(app.log)),
		bot.WithMiddlewares(telegram.Middlewares(
			app.handlers.Telegram.AllowedUsersMiddleware(app.log, app.cfg.Telegram.AllowedUsers),
		)...),
	)
	if err != nil {
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}

	// Initialize handlers with the bot
	if err = app.handlers.WithBot(app.bot).Init(ctx); err != nil {
		return fmt.Errorf("handlers start failed: %w", err)
	}

	return nil
}

// initStore initializes the data store.
func (app *App) initStore() (domain.Store, error) {
	// Connect to the database
	db, err := store.NewSQLite(app.cfg.DB.Filepath, app.cfg.DB.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return store.NewSQLiteStore(db), nil
}

// initServices initializes the application services.
func (app *App) initServices(ctx context.Context) (*services.Container, error) {
	// Create platforms
	ytDlp := platforms.NewYtDlpPlatform(app.cfg.Settings, app.log)

	// Create services
	s := services.New(app.cfg.Settings, app.log, app.store, ytDlp)

	// Initialize services
	if err := s.Init(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}
	return s, nil
}
