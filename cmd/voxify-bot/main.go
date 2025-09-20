package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ofstudio/voxify/internal/app"
	"github.com/ofstudio/voxify/internal/config"
	"github.com/ofstudio/voxify/pkg/shutdown"
)

func main() {
	// Create logger
	log := slog.Default()
	log.Info("starting", "version", config.Version())

	// Load the configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("fatal error: failed to load config: " + err.Error())
		os.Exit(-1)
	}

	// Create application context
	ctx, cancel := shutdown.Context(context.Background(), func(s os.Signal) {
		log.Warn("received signal: " + s.String())
	})
	defer cancel()

	// Start the application
	if err = app.New(cfg, log).Start(ctx); err != nil {
		log.Error("fatal error: " + err.Error())
		os.Exit(-1)
	}

	log.Info("exiting...")
}
