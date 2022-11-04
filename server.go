package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// Server runs a Prometheus metrics server. If Application contains one or more Handlers, it will also run an HTTP server
// for those handlers. The two HTTP servers use different TCP ports.
type Server struct {
	Name string
	Application
	Prometheus
	Metrics
}

func (s *Server) initialize() (err error) {
	s.Metrics.initialize(s.Name)
	if err = s.Application.initialize(s.Metrics); err == nil {
		err = s.Prometheus.initialize()
	}
	return err
}

// Run starts the HTTP server(s)
func (s *Server) Run() (err error) {
	if err = s.initialize(); err == nil {
		go func() {
			if err2 := s.Application.Run(); err2 != nil && !errors.Is(err2, http.ErrServerClosed) {
				panic(err2)
			}
		}()
		err = s.Prometheus.Run()
	}
	return err
}

// Shutdown performs a graceful shutdown of the HTTP server(s).
func (s *Server) Shutdown(timeout time.Duration) (err error) {
	if err = s.Application.Shutdown(timeout); err == nil {
		err = s.Prometheus.Shutdown(timeout)
	}
	return err
}

// HTTPServer is a helper structure to run HTTP servers for Application & Prometheus servers
type HTTPServer struct {
	Port     int
	listener net.Listener
	server   *http.Server
	lock     sync.RWMutex
}

func (b *HTTPServer) initialize(r http.Handler) (err error) {
	if b.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", b.Port)); err != nil {
		return fmt.Errorf("server: %w", err)

	}
	if b.Port == 0 {
		b.lock.Lock()
		b.Port = b.listener.Addr().(*net.TCPAddr).Port
		b.lock.Unlock()
	}
	b.server = &http.Server{Handler: r}
	return nil
}

// GetPort returns the TCP port on which the HTTP server is accepting connections. Useful when creating the server for a dynamic listening port
func (b *HTTPServer) GetPort() int {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.Port
}

// Run starts the HTTP Server
func (b *HTTPServer) Run() (err error) {
	if b.server != nil {
		err = b.server.Serve(b.listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
	}
	return
}

// Shutdown performs a graceful shutdown of the HTTP Server
func (b *HTTPServer) Shutdown(timeout time.Duration) (err error) {
	if b.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err = b.server.Shutdown(ctx)
	}
	return err
}
