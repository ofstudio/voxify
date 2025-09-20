package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/internal/platforms"
	"github.com/ofstudio/voxify/internal/services"
	"github.com/ofstudio/voxify/internal/store"
	"github.com/ofstudio/voxify/internal/telegram"
)

type App struct {
	cfg config.Config
	log *slog.Logger
}

func New(cfg config.Config, log *slog.Logger) *App {
	return &App{
		cfg: cfg,
		log: log,
	}
}

func (a *App) Start(ctx context.Context) error {

	// Connect to the database
	db, err := store.NewSQLite(a.cfg.DB.Filepath, a.cfg.DB.Version)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	a.log.Info("database connected", "filepath", a.cfg.DB.Filepath, "version", a.cfg.DB.Version)

	// Create store
	st := store.NewSQLiteStore(db)
	// Create platforms
	ytDlp := platforms.NewYtDlpPlatform(a.cfg.Settings, a.log)

	// Create services
	var (
		feedSrv    = services.NewFeedService(&a.cfg.Settings, a.log, st)
		episodeSrv = services.NewEpisodeService(&a.cfg.Settings, a.log, st, ytDlp)
		processSrv = services.NewProcessService(&a.cfg.Settings, a.log, st, episodeSrv, feedSrv)
	)

	// Initialize services
	if err = episodeSrv.Init(ctx); err != nil {
		return fmt.Errorf("episode service init failed: %w", err)
	}
	if err = feedSrv.Init(ctx); err != nil {
		return fmt.Errorf("feed service init failed: %w", err)
	}
	if err = processSrv.Init(ctx); err != nil {
		return fmt.Errorf("process service init failed: %w", err)
	}
	a.log.Info("services initialized")

	// Initialize Telegram bot
	middleware := telegram.NewMiddleware(a.cfg.Telegram, a.log)
	handlers := telegram.NewHandlers(a.cfg.Settings, a.log, processSrv, feedSrv)

	b, err := bot.New(a.cfg.Telegram.BotToken, []bot.Option{
		bot.WithMiddlewares(middleware.WithAllowedUsers()),
		bot.WithErrorsHandler(handlers.Error()),
	}...)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "start", bot.MatchTypeCommandStartOnly, handlers.CmdStart())
	b.RegisterHandler(bot.HandlerTypeMessageText, "build", bot.MatchTypeCommand, handlers.CmdBuild())
	b.RegisterHandler(bot.HandlerTypeMessageText, "https://", bot.MatchTypePrefix, handlers.Url())

	notifications := telegram.NewNotifications(b, processSrv, a.log)

	// Start bot and notifications
	ctxBot, cancelBot := context.WithCancel(ctx)
	defer cancelBot()
	go b.Start(ctxBot)
	a.log.Info("telegram bot started", "id", b.ID())

	// Start background services
	ctxSrv, cancelSrv := context.WithCancel(ctx)
	defer cancelSrv()
	notifications.Start(ctxSrv)
	processSrv.Start(ctxSrv)
	a.log.Info("background services started")

	// Wait for the context to be done
	a.log.Info("app is running")
	<-ctx.Done()

	// Shutdown

	// Stop background services
	cancelSrv()
	a.log.Info("background services stopped")

	// Stop the bot
	cancelBot()
	a.log.Info("telegram bot stopped")

	return nil
}
