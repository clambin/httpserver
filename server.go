package httpserver

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"time"
)

// Server implements a configurable HTTP Server. See the different WithXXX structs for available options.
type Server struct {
	port     int
	handlers []Handler
	metrics  Metrics
	listener net.Listener
	server   *http.Server
}

// Handler contains an endpoint to be registered in the Server's HTTP server
type Handler struct {
	// Path of the endpoint (e.g. "/health"). Can be any path that's valid for gorilla/mux router's Path().
	Path string
	// Handler that implements the endpoint
	Handler http.Handler
	// Methods that the handler should support. If empty, defaults to http.MethodGet
	Methods []string
}

// New returns a Server with the specified options
func New(options ...Option) (s *Server, err error) {
	s = new(Server)
	for _, o := range options {
		o.apply(s)
	}

	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return nil, fmt.Errorf("http server: %w", err)
	}

	r := mux.NewRouter()
	if s.metrics != nil {
		r.Use(s.handle)
	}
	for _, h := range s.handlers {
		if len(h.Methods) == 0 {
			h.Methods = []string{http.MethodGet}
		}
		r.Path(h.Path).Handler(h.Handler).Methods(h.Methods...)
	}
	s.server = &http.Server{Handler: r}
	return
}

// Run starts the HTTP server
func (s *Server) Run() error {
	err := s.server.Serve(s.listener)
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}

// Shutdown performs a graceful shutdown of the HTTP server
func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// GetPort returns the HTTP Server's listening port
func (s *Server) GetPort() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

// ServeHTTP calls the server's handler. Mainly intended to be used in unit tests without starting the underlying HTTP server.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.server.Handler.ServeHTTP(w, req)
}

func (s *Server) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		obs := s.metrics.GetRequestDurationMetric(r.Method, path)

		start := time.Now()
		next.ServeHTTP(lrw, r)

		counter := s.metrics.GetRequestCountMetric(r.Method, path, lrw.statusCode)
		counter.Inc()
		obs.Observe(time.Since(start).Seconds())
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
	statusCode  int
}

// WriteHeader implements the http.ResponseWriter interface.
func (w *loggingResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
	w.wroteHeader = true
}

// Write implements the http.ResponseWriter interface.
func (w *loggingResponseWriter) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}
