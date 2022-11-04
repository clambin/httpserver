package httpserver

import (
	"time"
)

// Server runs a Prometheus metrics server. If Application contains one or more Handlers, it will also run an HTTP server
// for those handlers. The two HTTP servers use different TCP ports.
type Server struct {
	Application
	Prometheus
}

// Run starts the HTTP server(s)
func (s *Server) Run() []error {
	var chs []chan error
	for _, f := range []func() error{s.Prometheus.Run, s.Application.Run} {
		ch := make(chan error)
		go func(f func() error, ch chan error) {
			ch <- f()
		}(f, ch)
		chs = append(chs, ch)
	}

	var errs []error
	for _, ch := range chs {
		if err := <-ch; err != nil {
			errs = append(errs, err)
		}

	}
	return errs
}

// Shutdown performs a graceful shutdown of the HTTP server(s).
func (s *Server) Shutdown(timeout time.Duration) []error {
	var chs []chan error
	for _, f := range []func(duration time.Duration) error{s.Prometheus.Shutdown, s.Application.Shutdown} {
		ch := make(chan error)
		go func(f func(duration time.Duration) error, ch chan error) {
			ch <- f(timeout)
		}(f, ch)
		chs = append(chs, ch)
	}

	var errs []error
	for _, ch := range chs {
		if err := <-ch; err != nil {
			errs = append(errs, err)
		}

	}
	return errs
}
