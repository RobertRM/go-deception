package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/RobertRM/go-deception/internal/config"
	"github.com/RobertRM/go-deception/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	logger.Info("getting config")
	config, err := config.Load("config.yaml")
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	logger.Info("config loaded")

	var wg sync.WaitGroup

	var runningServers []server.Server

	for _, listener := range config.Listeners {
		currentListener := listener
		// skip disabled listeners
		if !currentListener.Enabled {
			logger.Info("listener disabled", "name", currentListener.Name)
			continue
		}

		switch currentListener.Protocol {
		case "http":
			srvr := server.NewHTTPServer(currentListener, logger, nil)
			logger.Info("http server created", "name", currentListener.Name, "port", currentListener.Port)
			runningServers = append(runningServers, srvr)
		default:
			logger.Error("unknown protocol", "protocol", currentListener.Protocol)
		}
	}

	for _, srv := range runningServers {
		serverInstance := srv
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("http server starting", "name", serverInstance.Name())

			if err := serverInstance.Start(); err != nil && err != http.ErrServerClosed {
				logger.Error("http server failed unexpectedly", "name", serverInstance.Name(), "error", err)
			}
		}()
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt)
	<-shutdownChan

	logger.Info("shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	for _, srv := range runningServers {

		go func(s server.Server) {
			wg.Add(1)
			defer wg.Done()
			logger.Info("http server stopping", "name", s.Name())
			if err := s.Stop(shutdownCtx); err != nil {
				logger.Error("failed to stop server", "error", err)
			}
			logger.Info("http server stopped", "name", s.Name())
		}(srv)
	}

	wg.Wait()

	logger.Info("shutdown complete")
}
