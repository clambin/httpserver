package httpserver

import (
	"github.com/clambin/httpserver/testtools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer(t *testing.T) {
	r := prometheus.NewRegistry()
	requestCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "foo_http_requests_total",
		Help:        "Total number of http requests",
		ConstLabels: prometheus.Labels{"handler": "test"},
	}, []string{"method", "path", "code"})
	durationHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "foo_http_requests_duration_seconds",
		Help:        "Request duration in seconds",
		ConstLabels: prometheus.Labels{"handler": "test"},
		Buckets:     DefBuckets,
	}, []string{"method", "path"})
	r.MustRegister(requestCounter, durationHistogram)

	s := Server{
		Name: "test",
		ApplicationServer: ApplicationServer{
			Handlers: []Handler{{
				Path: "/foo",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					_, _ = w.Write([]byte("OK"))
				}),
				Methods: []string{http.MethodGet},
			}},
		},
		Metrics: Metrics{
			RequestCounter:    requestCounter,
			DurationHistogram: durationHistogram,
		},
	}

	err := s.initialize()
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
	resp := httptest.NewRecorder()

	s.ApplicationServer.server.Handler.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "OK", resp.Body.String())

	req, _ = http.NewRequest(http.MethodPut, "/foo", nil)
	resp = httptest.NewRecorder()

	s.ApplicationServer.server.Handler.ServeHTTP(resp, req)
	require.Equal(t, http.StatusMethodNotAllowed, resp.Code)

	metrics, err := r.Gather()
	require.NoError(t, err)
	assert.Len(t, metrics, 2)

	for _, metric := range metrics {
		switch metric.GetName() {
		case "foo_http_requests_total":
			testtools.CheckRequests(t, metric, "test", "get", "/foo", "200")
		case "foo_http_requests_duration_seconds":
			testtools.CheckDuration(t, metric, "test", "get", "/foo")
		default:
			t.Fatalf("unexpected metric name: %s", metric.GetName())
		}
	}
}

func TestServer_Default_Metrics(t *testing.T) {
	s := Server{
		Name: "test",
		ApplicationServer: ApplicationServer{
			Handlers: []Handler{{
				Path: "/foo",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("OK"))
				}),
			}},
		},
	}

	err := s.initialize()
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
	resp := httptest.NewRecorder()

	s.ApplicationServer.server.Handler.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "OK", resp.Body.String())

	req, _ = http.NewRequest(http.MethodPut, "/foo", nil)
	resp = httptest.NewRecorder()

	s.ApplicationServer.server.Handler.ServeHTTP(resp, req)
	require.Equal(t, http.StatusMethodNotAllowed, resp.Code)

	metrics, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	var count int
	for _, metric := range metrics {
		switch metric.GetName() {
		case "http_requests_total":
			testtools.CheckRequests(t, metric, "test", "get", "/foo", "200")
			count++
		case "http_requests_duration_seconds":
			testtools.CheckDuration(t, metric, "test", "get", "/foo")
			count++
		}
	}
	assert.Equal(t, 2, count)
}
