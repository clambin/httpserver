package httpserver_test

import (
	"fmt"
	"github.com/clambin/httpserver"
	"github.com/clambin/httpserver/testtools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestServer_Run(t *testing.T) {
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
		Buckets:     httpserver.DefBuckets,
	}, []string{"method", "path"})
	r.MustRegister(requestCounter, durationHistogram)

	s := httpserver.Server{
		Name: "test",
		ApplicationServer: httpserver.ApplicationServer{
			Handlers: []httpserver.Handler{
				{
					Path: "/foo",
					Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						_, _ = w.Write([]byte("OK"))
					}),
				},
			},
		},
		Metrics: httpserver.Metrics{
			RequestCounter:    requestCounter,
			DurationHistogram: durationHistogram,
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := s.Run()
		require.NoError(t, err)
		wg.Done()
	}()

	assert.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", s.Prometheus.GetPort()))
		if err == nil {
			_ = resp.Body.Close()
		}
		return err == nil && resp.StatusCode == http.StatusOK
	}, time.Second, 10*time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/foo", s.ApplicationServer.GetPort()))
	assert.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	metrics, err := r.Gather()
	require.NoError(t, err)

	for _, metric := range metrics {
		switch metric.GetName() {
		case "http_requests_total":
			testtools.CheckRequests(t, metric, "test", "get", "/foo", "200")
		case "http_requests_duration_seconds":
			testtools.CheckDuration(t, metric, "test", "get", "/foo")
		}
	}

	err = s.Shutdown(30 * time.Second)
	require.NoError(t, err)

	wg.Wait()
}
