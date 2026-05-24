package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"effective-mobile-subscription-server/internal/app"
	"effective-mobile-subscription-server/internal/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger := logrus.New()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer func() {
		logger.Info("Received an interrupt, shutting down...")
		stop()
	}()

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal(err)
	}

	subsApp, err := app.NewApp(ctx, cfg, logger)
	if err != nil {
		logger.Fatal(err)
	}

	g, gCtx := errgroup.WithContext(ctx)

	// Server goroutine
	g.Go(func() error {
		logger.WithFields(logrus.Fields{
			"address": cfg.GetAddress(),
		}).Info("Starting HTTP server")
		if err := subsApp.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("subscriptionApp error: %w", err)
		}
		logger.Info("Server stopped")
		return nil
	})

	// App stop goroutine
	g.Go(func() error {
		<-gCtx.Done()
		return subsApp.Stop(ctx)
	})

	if err := g.Wait(); err != nil {
		logger.WithError(err).Error("exit reason: ")
	}
}
