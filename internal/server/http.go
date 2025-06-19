package server

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"text/template"

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

	// add static resources
	contentRoot, err := fs.Sub(iconsFS, "templates/icons")
	if err != nil {
		logger.Error("Failed to create sub filesystem for icons", "error", err)
	} else {
		fileServer := http.FileServer(http.FS(contentRoot))
		mux.Handle("/icons/", http.StripPrefix("/icons/", fileServer))
	}

	for _, route := range listener.Routes {
		currentRoute := route
		handler := &routeHandler{
			route:    currentRoute,
			logger:   logger,
			listener: listener,
		}

		mux.Handle(currentRoute.Path, handler)
	}

	return mux
}

// implements http.Handler
type routeHandler struct {
	route    config.Route
	listener config.Listener
	logger   *slog.Logger
}

func (h *routeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	h.logger.Info(
		"Request received",
		"listener", h.listener.Name,
		"source_ip", r.RemoteAddr,
		"method", r.Method,
		"path", r.URL.Path,
	)

	for key, value := range h.route.Response.Headers {
		w.Header().Set(key, value)
	}

	if h.route.Response.Template != "" {
		tmpl, err := template.ParseFS(templateFS, "templates/"+h.route.Response.Template)
		if err != nil {
			h.logger.Error("Failed to parse template", "template", h.route.Response.Template, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(h.route.Response.StatusCode)

		err = tmpl.Execute(w, h.route.Response.Vars)
		if err != nil {
			h.logger.Error("Failed to execute template", "template", h.route.Response.Template, "error", err)
		}

	} else {
		w.WriteHeader(h.route.Response.StatusCode)
		fmt.Fprint(w, h.route.Response.Body)
	}
}
