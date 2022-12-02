package httpserver

import (
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"strings"
)

// Metrics interface contains the methods httpserver's middleware expects to record performance metrics
type Metrics interface {
	GetRequestDurationMetric(method, path string) prometheus.Observer
	GetRequestCountMetric(method, path string, statusCode int) prometheus.Counter
	prometheus.Collector
}

// SLOMetrics uses a histogram to record request duration metrics. Use this to measure an SLO (e.g. 95% of all requests must be serviced below x seconds).
// SLOMetrics uses Prometheus' default buckets.
type SLOMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

var _ Metrics = &SLOMetrics{}

// NewSLOMetrics creates a new SLOMetrics, where latency is measured using a histogram with the provided list of buckets.
// If the list is empty, NewSLOMetrics will use prometheus.DefBuckets.
func NewSLOMetrics(name string, buckets []float64) Metrics {
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}
	return &SLOMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of http requests",
			ConstLabels: prometheus.Labels{"handler": name},
		}, []string{"method", "path", "code"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:        "http_requests_duration_seconds",
			Help:        "Request duration in seconds",
			ConstLabels: prometheus.Labels{"handler": name},
			Buckets:     buckets,
		}, []string{"method", "path"}),
	}
}

// GetRequestDurationMetric returns the Observer to record request duration
func (m *SLOMetrics) GetRequestDurationMetric(method, path string) prometheus.Observer {
	return m.duration.With(prometheus.Labels{
		"method": strings.ToLower(method),
		"path":   path,
	})
}

// GetRequestCountMetric returns the Counter to record request count
func (m *SLOMetrics) GetRequestCountMetric(method, path string, statusCode int) prometheus.Counter {
	return m.requests.With(prometheus.Labels{
		"method": strings.ToLower(method),
		"path":   path,
		"code":   strconv.Itoa(statusCode),
	})
}

func (m *SLOMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.duration.Describe(ch)
	m.requests.Describe(ch)
}

func (m *SLOMetrics) Collect(ch chan<- prometheus.Metric) {
	m.duration.Collect(ch)
	m.requests.Collect(ch)
}

// AvgMetrics uses a Summary to record request duration metrics. Use this if you are only interested in the average time to service requests.
type AvgMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.SummaryVec
}

var _ Metrics = &AvgMetrics{}

// NewAvgMetrics creates a new AvgMetrics.
func NewAvgMetrics(name string) Metrics {
	return &AvgMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of http requests",
			ConstLabels: prometheus.Labels{"handler": name},
		}, []string{"method", "path", "code"}),
		duration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:        "http_requests_duration_seconds",
			Help:        "Request duration in seconds",
			ConstLabels: prometheus.Labels{"handler": name},
		}, []string{"method", "path"}),
	}
}

// GetRequestDurationMetric returns the Observer to record request duration
func (m *AvgMetrics) GetRequestDurationMetric(method, path string) prometheus.Observer {
	return m.duration.With(prometheus.Labels{
		"method": strings.ToLower(method),
		"path":   path,
	})
}

// GetRequestCountMetric returns the Counter to record request count
func (m *AvgMetrics) GetRequestCountMetric(method, path string, statusCode int) prometheus.Counter {
	return m.requests.With(prometheus.Labels{
		"method": strings.ToLower(method),
		"path":   path,
		"code":   strconv.Itoa(statusCode),
	})
}

func (m *AvgMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.duration.Describe(ch)
	m.requests.Describe(ch)
}

func (m *AvgMetrics) Collect(ch chan<- prometheus.Metric) {
	m.duration.Collect(ch)
	m.requests.Collect(ch)
}
