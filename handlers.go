package httpserver

import (
	"github.com/gorilla/mux"
	"net/http"
)

// Application runs an HTTP Server for one or more application Handlers
type Application struct {
	Handlers []Handler
	HTTPServer
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

func (s *Application) initialize(metrics Metrics) (err error) {
	if len(s.Handlers) == 0 {
		return
	}
	r := mux.NewRouter()
	r.Use(metrics.handle)
	for _, h := range s.Handlers {
		methods := h.Methods
		if len(methods) == 0 {
			methods = []string{http.MethodGet}
		}
		r.Path(h.Path).Handler(h.Handler).Methods(methods...)
	}
	return s.HTTPServer.initialize(r)
}
