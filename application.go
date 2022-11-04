package httpserver

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Application runs an HTTP Server for one or more application Handlers
type Application struct {
	Name     string
	Port     int
	Handlers []Handler
	Metrics
	httpServer
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

func (a *Application) Run() error {
	a.Metrics.initialize(a.Name)
	r := mux.NewRouter()
	r.Use(a.Metrics.handle)
	for _, h := range a.Handlers {
		methods := h.Methods
		if len(methods) == 0 {
			methods = []string{http.MethodGet}
		}
		r.Path(h.Path).Handler(h.Handler).Methods(methods...)
	}
	return a.httpServer.Run(a.Port, r)
}

var (
	// DefBuckets contains the default buckets for the Duration histogram metric
	DefBuckets = []float64{.001, .01, .1, 1, 10}
)

// Metrics contains the metrics that will be captured while serving HTTP requests. If these are not provided then
// Server will create default metrics and register them with Prometheus' default registry.
type Metrics struct {
	// RequestCounter records the number of times a handler was called. This is a prometheus.CounterVec with three labels: "method", "path" and "code".
	// By default, a metric called "http_requests_totals" will be used
	RequestCounter *prometheus.CounterVec
	// DurationHistogram records the latency of each handler call. This is a prometheus.HistogramVec with two labels: "method" and "path".
	// By default, a metric called "http_requests_duration_seconds" will be used, with DefBuckets as the histogram's buckets.
	DurationHistogram *prometheus.HistogramVec
}

func (m *Metrics) initialize(name string) {
	if m.RequestCounter == nil {
		m.RequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of http requests",
			ConstLabels: prometheus.Labels{"handler": name},
		}, []string{"method", "path", "code"})

	}
	if m.DurationHistogram == nil {
		m.DurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:        "http_requests_duration_seconds",
			Help:        "Request duration in seconds",
			ConstLabels: prometheus.Labels{"handler": name},
			Buckets:     DefBuckets,
		}, []string{"method", "path"})
	}
}

func (m *Metrics) handle(next http.Handler) http.Handler {
	return instrumentHandlerCounter(m.RequestCounter,
		instrumentHandlerDuration(m.DurationHistogram,
			next,
		),
	)
}

func instrumentHandlerCounter(counter *prometheus.CounterVec, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		counter.With(prometheus.Labels{
			"method": strings.ToLower(r.Method),
			"path":   path,
			"code":   strconv.Itoa(lrw.statusCode),
		}).Inc()
	}
}

func instrumentHandlerDuration(histogram *prometheus.HistogramVec, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		histogram.With(prometheus.Labels{
			"method": strings.ToLower(r.Method),
			"path":   path,
		}).Observe(time.Since(start).Seconds())
	}
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
