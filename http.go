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

// httpServer is a helper structure to run HTTP servers for Application & Prometheus servers
type httpServer struct {
	listener net.Listener
	server   *http.Server
	lock     sync.RWMutex
}

// Run starts the HTTP Server
func (s *httpServer) Run(port int, r http.Handler) error {
	if err := s.initialize(port, r); err != nil {
		return err
	}
	err := s.server.Serve(s.listener)
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}

func (s *httpServer) initialize(port int, r http.Handler) (err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
		return fmt.Errorf("http server: %w", err)

	}
	s.server = &http.Server{Handler: r}
	return nil
}

// Shutdown performs a graceful shutdown of the HTTP Server
func (s *httpServer) Shutdown(timeout time.Duration) (err error) {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err = s.server.Shutdown(ctx)
		s.server = nil
	}
	return err
}

// GetPort returns the TCP port on which the HTTP server is accepting connections. Useful when creating the server for a dynamic listening port
func (s *httpServer) GetPort() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.listener.Addr().(*net.TCPAddr).Port
}
