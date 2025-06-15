package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/RobertRM/go-deception/internal/config"
)

type HTTPServer struct {
	listener config.Listener
	logger   *slog.Logger
	mux      *http.ServeMux
}

func NewHTTPServer(listener config.Listener, logger *slog.Logger, mux *http.ServeMux) *HTTPServer {
	if mux == nil {
		mux = http.NewServeMux()
	}

	server := &HTTPServer{
		listener: listener,
		logger:   logger,
		mux:      mux,
	}

	// register routes defined in listener
	server.registerRoutes()

	return server
}

func (s *HTTPServer) Start() error {
	return nil
}

func (s *HTTPServer) Stop() error {
	return nil
}

func (s *HTTPServer) registerRoutes() {
	for _, route := range s.listener.Routes {
		currentRoute := route
		handler := func(w http.ResponseWriter, r *http.Request) {
			s.handleRequest(w, r, currentRoute)
		}
		s.mux.HandleFunc(currentRoute.Path, handler)
	}
}

func (s *HTTPServer) handleRequest(w http.ResponseWriter, r *http.Request, route config.Route) {
	// Need to implement the body, template and headers in the response
	s.logger.Info("Request received", "path", r.URL.Path, "method", r.Method, "source_ip", r.RemoteAddr)
	fmt.Fprintf(w, "Response for %s on listener %s", route.Path, s.listener.Name)
}
