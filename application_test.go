package httpserver_test

import (
	"fmt"
	"github.com/clambin/httpserver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestApplication_Run(t *testing.T) {
	requestCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name:        "http_requests_total",
		Help:        "Total number of http requests",
		ConstLabels: prometheus.Labels{"handler": "foo"},
	}, []string{"method", "path", "code"})
	durationHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "http_requests_duration_seconds",
		Help:        "Request duration in seconds",
		ConstLabels: prometheus.Labels{"handler": "foo"},
		Buckets:     httpserver.DefBuckets,
	}, []string{"method", "path"})
	r := prometheus.NewRegistry()
	r.MustRegister(requestCounter, durationHistogram)

	a := httpserver.Application{
		Name: "foo",
		Handlers: []httpserver.Handler{
			{
				Path: "/foo",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					_, _ = w.Write([]byte("foo"))
				}),
			},
			{
				Path: "/bar",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					_, _ = w.Write([]byte("bar"))
				}),
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
		err := a.Run()
		require.NoError(t, err)
		wg.Done()
	}()

	assert.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/foo", a.GetPort()))
		if err == nil {
			_ = resp.Body.Close()
		}
		return err == nil && resp.StatusCode == http.StatusOK
	}, time.Second, 10*time.Millisecond)

	assert.NotZero(t, a.GetPort())

	for _, ep := range []string{"bar"} {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/%s", a.GetPort(), ep))
		require.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		assert.Equal(t, ep, string(body))
	}

	info := getMetricInfo(t, r, "http_requests_total")
	require.Len(t, info, 2)
	assert.Equal(t, 1.0, info[0].metric.GetCounter().GetValue())
	assert.Equal(t, map[string]string{
		"code": "200", "handler": "foo", "method": "get", "path": "/bar",
	}, info[0].labels)
	assert.Equal(t, 1.0, info[1].metric.GetCounter().GetValue())
	assert.Equal(t, map[string]string{
		"code": "200", "handler": "foo", "method": "get", "path": "/foo",
	}, info[1].labels)

	info = getMetricInfo(t, r, "http_requests_duration_seconds")
	require.Len(t, info, 2)
	assert.Equal(t, uint64(1), info[0].metric.GetHistogram().GetSampleCount())
	assert.Equal(t, map[string]string{
		"handler": "foo", "method": "get", "path": "/bar",
	}, info[0].labels)
	assert.Equal(t, uint64(1), info[1].metric.GetHistogram().GetSampleCount())
	assert.Equal(t, map[string]string{
		"handler": "foo", "method": "get", "path": "/foo",
	}, info[1].labels)

	err := a.Shutdown(time.Minute)
	require.NoError(t, err)
	wg.Wait()
}

func TestApplication_Run_Default_Metrics(t *testing.T) {
	a := httpserver.Application{
		Name: "snafu",
		Handlers: []httpserver.Handler{
			{
				Path: "/foo",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					_, _ = w.Write([]byte("foo"))
				}),
			},
			{
				Path: "/bar",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					_, _ = w.Write([]byte("bar"))
				}),
			},
		},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := a.Run()
		require.NoError(t, err)
		wg.Done()
	}()

	assert.Eventually(t, func() bool {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/foo", a.GetPort()))
		if err == nil {
			_ = resp.Body.Close()
		}
		return err == nil && resp.StatusCode == http.StatusOK
	}, time.Second, 10*time.Millisecond)

	assert.NotZero(t, a.GetPort())

	for _, ep := range []string{"bar"} {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/%s", a.GetPort(), ep))
		require.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		assert.Equal(t, ep, string(body))
	}

	info := getMetricInfo(t, prometheus.DefaultGatherer, "http_requests_total")
	require.Len(t, info, 2)
	assert.Equal(t, 1.0, info[0].metric.GetCounter().GetValue())
	assert.Equal(t, map[string]string{
		"code": "200", "handler": "snafu", "method": "get", "path": "/bar",
	}, info[0].labels)
	assert.Equal(t, 1.0, info[1].metric.GetCounter().GetValue())
	assert.Equal(t, map[string]string{
		"code": "200", "handler": "snafu", "method": "get", "path": "/foo",
	}, info[1].labels)

	info = getMetricInfo(t, prometheus.DefaultGatherer, "http_requests_duration_seconds")
	require.Len(t, info, 2)
	assert.Equal(t, uint64(1), info[0].metric.GetHistogram().GetSampleCount())
	assert.Equal(t, map[string]string{
		"handler": "snafu", "method": "get", "path": "/bar",
	}, info[0].labels)
	assert.Equal(t, uint64(1), info[1].metric.GetHistogram().GetSampleCount())
	assert.Equal(t, map[string]string{
		"handler": "snafu", "method": "get", "path": "/foo",
	}, info[1].labels)

	err := a.Shutdown(time.Minute)
	require.NoError(t, err)
	wg.Wait()

}
