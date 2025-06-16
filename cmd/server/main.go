package main

import (
	"log/slog"
	"os"
	"os/signal"
	"sync"

	"github.com/RobertRM/go-deception/internal/config"
	"github.com/RobertRM/go-deception/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mydir, err := os.Getwd()
	if err != nil {
		logger.Error("failed to get working directory", "error", err)
	}
	logger.Info("working directory", "path", mydir)

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
		currentSrv := srv
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("starting server", "name", currentSrv.Start())
			if err := currentSrv.Start(); err != nil {
				logger.Error("failed to start server", "error", err)
			}
			logger.Info("server stopped", "name", currentSrv.Stop())
		}()
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt)
	<-shutdownChan

	logger.Info("shutting down")

	wg.Wait()

	logger.Info("shutdown complete")
}
