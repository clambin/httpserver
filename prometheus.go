package httpserver

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Prometheus contains the different Prometheus configuration options
type Prometheus struct {
	Path string
	BaseServer
}

func (p *Prometheus) initialize() (err error) {
	if p.Path == "" {
		p.Path = "/metrics"
	}
	r := mux.NewRouter()
	r.Path(p.Path).Handler(promhttp.Handler()).Methods(http.MethodGet)
	return p.BaseServer.initialize(r)
}

var (
	DefBuckets = []float64{.001, .01, .1, 1, 10}
)

// Metrics contains the metrics that need to be captured while serving HTTP requests. If these are not provided then
// Server will create default metrics and register them with Prometheus' default registry.
type Metrics struct {
	RequestCounter    *prometheus.CounterVec
	DurationHistogram *prometheus.HistogramVec
}

func (m *Metrics) initialize(name string) {
	if m.RequestCounter != nil && m.DurationHistogram != nil {
		return
	}

	m.RequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "http_requests_total",
		Help:        "Total number of http requests",
		ConstLabels: prometheus.Labels{"handler": name},
	}, []string{"method", "path", "code"})
	m.DurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "http_requests_duration_seconds",
		Help:        "Request duration in seconds",
		ConstLabels: prometheus.Labels{"handler": name},
		Buckets:     DefBuckets,
	}, []string{"method", "path"})
	prometheus.DefaultRegisterer.MustRegister(m.RequestCounter, m.DurationHistogram)
}

func (m *Metrics) Handle(next http.Handler) http.Handler {
	return InstrumentHandlerCounter(m.RequestCounter,
		InstrumentHandlerDuration(m.DurationHistogram,
			next,
		),
	)
}

func InstrumentHandlerCounter(counter *prometheus.CounterVec, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		counter.With(prometheus.Labels{
			"method": strings.ToLower(r.Method),
			"path":   r.URL.Path,
			"code":   strconv.Itoa(lrw.statusCode),
		}).Inc()
	}
}

func InstrumentHandlerDuration(histogram *prometheus.HistogramVec, next http.Handler) http.HandlerFunc {
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

// loggingResponseWriter records the HTTP status code of a ResponseWriter, so we can use it to log response times for
// individual status codes.
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
