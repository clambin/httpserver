package httpserver

import (
	"github.com/gorilla/mux"
	"net/http"
)

// ApplicationServer contains the configuration items for an HTTP Server
type ApplicationServer struct {
	Handlers []Handler
	BaseServer
}

// Handler contains an endpoint to be registered in the Server's HTTP server, using NewWithHandlers.
type Handler struct {
	// Path of the endpoint (e.g. "/health"). Must include the leading /
	Path string
	// Handler that implements the endpoint
	Handler http.Handler
	// Methods that the handler should support. If empty, http.MethodGet is the default
	Methods []string
}

func (s *ApplicationServer) initialize(metrics Metrics) (err error) {
	if len(s.Handlers) == 0 {
		return
	}
	r := mux.NewRouter()
	r.Use(metrics.Handle)
	for _, h := range s.Handlers {
		methods := h.Methods
		if len(methods) == 0 {
			methods = []string{http.MethodGet}
		}
		r.Path(h.Path).Handler(h.Handler).Methods(methods...)
	}
	return s.BaseServer.initialize(r)
}
