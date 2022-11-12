package httpserver

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Option specified configuration options for Server
type Option interface {
	apply(server *Server)
}

// WithPort specifies the Server's listening port. If no port is specified, Server will listen on a random port.
// Use GetPort() to determine the actual listening port
type WithPort struct {
	Port int
}

func (o WithPort) apply(s *Server) {
	s.port = o.Port
}

// WithPrometheus adds a Prometheus metrics endpoint to the server at the specified Path. Default path is "/metrics"
type WithPrometheus struct {
	Path string
}

func (o WithPrometheus) apply(s *Server) {
	if o.Path == "" {
		o.Path = "/metrics"
	}
	s.handlers = append(s.handlers, Handler{
		Path:    o.Path,
		Handler: promhttp.Handler(),
		Methods: []string{http.MethodGet},
	})
}

// WithHandlers adds the specified handlers to the server
type WithHandlers struct {
	Handlers []Handler
}

func (o WithHandlers) apply(s *Server) {
	s.handlers = append(s.handlers, o.Handlers...)
}

// WithMetrics will collect the specified metrics to instrument the Server's Handlers.
type WithMetrics struct {
	Metrics *Metrics
}

func (o WithMetrics) apply(s *Server) {
	if o.Metrics == nil {
		o.Metrics = NewMetrics("default")
		o.Metrics.Register(prometheus.DefaultRegisterer)
	}
	s.metrics = o.Metrics
}
