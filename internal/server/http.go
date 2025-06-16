package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/RobertRM/go-deception/internal/config"
)

type HTTPServer struct {
	listener config.Listener
	logger   *slog.Logger
	server   *http.Server
}

func NewHTTPServer(listener config.Listener, logger *slog.Logger, server *http.Server) Server {
	if server == nil {
		mux := BuildMuxForListener(listener, logger)
		server = &http.Server{
			Addr:    fmt.Sprintf(":%d", listener.Port),
			Handler: mux,
		}
	}

	return &HTTPServer{
		listener: listener,
		logger:   logger,
		server:   server,
	}
}

func (s *HTTPServer) Name() string {
	return s.listener.Name
}

func (s *HTTPServer) Start() error {
	if s.server == nil {
		return fmt.Errorf("server not initialized")
	}

	return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	return s.server.Shutdown(ctx)
}

func BuildMuxForListener(listener config.Listener, logger *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	for _, route := range listener.Routes {
		currentRoute := route
		handler := &routeHandler{
			route:  currentRoute,
			logger: logger,
		}

		mux.Handle(currentRoute.Path, handler)
	}

	return mux
}

// implements http.Handler
type routeHandler struct {
	route  config.Route
	logger *slog.Logger
}

func (h *routeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Request received", "path", r.URL.Path, "method", r.Method, "source_ip", r.RemoteAddr)
	fmt.Fprintf(w, "Response for %s", h.route.Path)
}
