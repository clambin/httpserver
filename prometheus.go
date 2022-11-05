package httpserver

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// Prometheus runs a Prometheus metrics server
type Prometheus struct {
	Path string
	Port int
	httpServer
}

// Run starts the HTTP Server
func (p *Prometheus) Run() error {
	if p.Path == "" {
		p.Path = "/metrics"
	}
	r := mux.NewRouter()
	r.Path(p.Path).Handler(promhttp.Handler()).Methods(http.MethodGet)
	return p.httpServer.Run(p.Port, r)
}
