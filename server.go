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

type Server struct {
	Name string
	ApplicationServer
	Prometheus
	Metrics
}

func (s *Server) initialize() (err error) {
	s.Metrics.initialize(s.Name)
	if err = s.ApplicationServer.initialize(s.Metrics); err == nil {
		err = s.Prometheus.initialize()
	}
	return err
}

// Run starts the Server
func (s *Server) Run() (err error) {
	if err = s.initialize(); err == nil {
		go func() {
			if err2 := s.ApplicationServer.Run(); err2 != nil && !errors.Is(err2, http.ErrServerClosed) {
				panic(err2)
			}
		}()
		err = s.Prometheus.Run()
	}
	return err
}

// Shutdown performs a graceful shutdown of the HTTP Server.
func (s *Server) Shutdown(timeout time.Duration) (err error) {
	if err = s.ApplicationServer.Shutdown(timeout); err == nil {
		err = s.Prometheus.Shutdown(timeout)
	}
	return err
}

type BaseServer struct {
	Port     int
	listener net.Listener
	server   *http.Server
	lock     sync.RWMutex
}

func (b *BaseServer) initialize(r http.Handler) (err error) {
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

func (b *BaseServer) GetPort() int {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return b.Port
}

// Run starts the HTTP Server. This calls server's http.Server's Serve method and returns that method's return value.
func (b *BaseServer) Run() (err error) {
	if b.server != nil {
		err = b.server.Serve(b.listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
	}
	return
}

// Shutdown performs a graceful shutdown of the HTTP Server.
func (b *BaseServer) Shutdown(timeout time.Duration) (err error) {
	if b.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err = b.server.Shutdown(ctx)
	}
	return err
}
