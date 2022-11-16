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
}

// SLOMetrics uses a histogram to record request duration metrics. Use this to measure an SLO (e.g. 95% of all requests must be serviced below x seconds).
// SLOMetrics uses Prometheus' default buckets.
type SLOMetrics struct {
	// RequestCounter records the number of times a handler was called. This is a prometheus.CounterVec with three labels: "method", "path" and "code".
	// By default, a metric called "http_requests_totals" will be used
	RequestCounter *prometheus.CounterVec
	// DurationHistogram records the latency of each handler call. This is a prometheus.HistogramVec with two labels: "method" and "path".
	// By default, a metric called "http_requests_duration_seconds" will be used, with DefBuckets as the histogram's buckets.
	DurationHistogram *prometheus.HistogramVec
}

var _ Metrics = &SLOMetrics{}

// NewSLOMetrics creates a new SLOMetrics and registers it with the provided prometheus.Registerer. If no registerer is provided,
// the metrics will be registered with prometheus.DefaultRegisterer.
// NewSLOMetrics uses the standard prometheus default buckets (prometheus.DefBuckets). To override the default buckets, use NewSLOMetricsWithBuckets().
func NewSLOMetrics(name string, r prometheus.Registerer) Metrics {
	return NewSLOMetricsWithBuckets(name, prometheus.DefBuckets, r)
}

// NewSLOMetricsWithBuckets creates a new SLOMetrics with the specified buckets and registers it with the provided prometheus.Registerer. If no registerer is provided,
// the metrics will be registered with prometheus.DefaultRegisterer.
func NewSLOMetricsWithBuckets(name string, buckets []float64, r prometheus.Registerer) Metrics {
	metrics := &SLOMetrics{
		RequestCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of http requests",
			ConstLabels: prometheus.Labels{"handler": name},
		}, []string{"method", "path", "code"}),
		DurationHistogram: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:        "http_requests_duration_seconds",
			Help:        "Request duration in seconds",
			ConstLabels: prometheus.Labels{"handler": name},
			Buckets:     buckets,
		}, []string{"method", "path"}),
	}
	if r == nil {
		r = prometheus.DefaultRegisterer
	}
	r.MustRegister(metrics.RequestCounter, metrics.DurationHistogram)

	return metrics

}

// GetRequestDurationMetric returns the Observer to record request duration
func (m *SLOMetrics) GetRequestDurationMetric(method, path string) prometheus.Observer {
	return m.DurationHistogram.With(prometheus.Labels{
		"method": strings.ToLower(method),
		"path":   path,
	})
}

// GetRequestCountMetric returns the Counter to record request count
func (m *SLOMetrics) GetRequestCountMetric(method, path string, statusCode int) prometheus.Counter {
	return m.RequestCounter.With(prometheus.Labels{
		"method": strings.ToLower(method),
		"path":   path,
		"code":   strconv.Itoa(statusCode),
	})
}

// AvgMetrics uses a Summary to record request duration metrics. Use this if you are only interested in the average time to service requests.
type AvgMetrics struct {
	// Requests records the number of times a handler was called. This is a prometheus.CounterVec with three labels: "method", "path" and "code".
	Requests *prometheus.CounterVec
	// Duration records the latency of each handler call. This is a prometheus.HistogramVec with two labels: "method" and "path".
	Duration *prometheus.SummaryVec
}

var _ Metrics = &AvgMetrics{}

// NewAvgMetrics creates a new AvgMetrics and registers it with the provided prometheus.Registerer. If no registerer is provided,
// the metrics will be registered with prometheus.DefaultRegisterer.
func NewAvgMetrics(name string, r prometheus.Registerer) Metrics {
	metrics := &AvgMetrics{
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total number of http requests",
			ConstLabels: prometheus.Labels{"handler": name},
		}, []string{"method", "path", "code"}),
		Duration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:        "http_requests_duration_seconds",
			Help:        "Request duration in seconds",
			ConstLabels: prometheus.Labels{"handler": name},
		}, []string{"method", "path"}),
	}
	if r == nil {
		r = prometheus.DefaultRegisterer
	}
	r.MustRegister(metrics.Requests, metrics.Duration)

	return metrics
}

// GetRequestDurationMetric returns the Observer to record request duration
func (m *AvgMetrics) GetRequestDurationMetric(method, path string) prometheus.Observer {
	return m.Duration.With(prometheus.Labels{
		"method": strings.ToLower(method),
		"path":   path,
	})
}

// GetRequestCountMetric returns the Counter to record request count
func (m *AvgMetrics) GetRequestCountMetric(method, path string, statusCode int) prometheus.Counter {
	return m.Requests.With(prometheus.Labels{
		"method": strings.ToLower(method),
		"path":   path,
		"code":   strconv.Itoa(statusCode),
	})
}
