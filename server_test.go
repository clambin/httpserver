package httpserver_test

import (
	"fmt"
	"github.com/clambin/httpserver"
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
		Application: httpserver.Application{
			Name: "test",
			Handlers: []httpserver.Handler{
				{
					Path: "/foo",
					Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						_, _ = w.Write([]byte("OK"))
					}),
				},
			},
			Metrics: httpserver.Metrics{
				RequestCounter:    requestCounter,
				DurationHistogram: durationHistogram,
			},
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := s.Run()
		require.Empty(t, err)
		wg.Done()
	}()

	assert.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", s.Prometheus.GetPort()))
		if err == nil {
			_ = resp.Body.Close()
		}
		return err == nil && resp.StatusCode == http.StatusOK
	}, time.Second, 10*time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/foo", s.Application.GetPort()))
	assert.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	info := getMetricInfo(t, r, "foo_http_requests_total")
	require.Len(t, info, 1)
	assert.Equal(t, 1.0, info[0].metric.GetCounter().GetValue())
	assert.Equal(t, map[string]string{
		"code": "200", "handler": "test", "method": "get", "path": "/foo",
	}, info[0].labels)

	info = getMetricInfo(t, r, "foo_http_requests_duration_seconds")
	require.Len(t, info, 1)
	assert.Equal(t, uint64(1), info[0].metric.GetHistogram().GetSampleCount())
	assert.Equal(t, map[string]string{
		"handler": "test", "method": "get", "path": "/foo",
	}, info[0].labels)

	errs := s.Shutdown(30 * time.Second)
	require.Empty(t, errs)

	wg.Wait()
}

func TestServer_Run_BadPort(t *testing.T) {
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
		Application: httpserver.Application{
			Name: "test",
			Port: -1,
			Handlers: []httpserver.Handler{
				{
					Path: "/foo",
					Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						_, _ = w.Write([]byte("OK"))
					}),
				},
			},
			Metrics: httpserver.Metrics{
				RequestCounter:    requestCounter,
				DurationHistogram: durationHistogram,
			},
		},
		Prometheus: httpserver.Prometheus{
			Port: -1,
		},
	}

	errs := s.Run()
	require.Len(t, errs, 2)
}
