package testtools

import (
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func CheckRequests(t *testing.T, metric *io_prometheus_client.MetricFamily, handler, method, path, code string) {
	t.Helper()
	assert.Equal(t, io_prometheus_client.MetricType_COUNTER, metric.GetType())
	require.Len(t, metric.Metric, 1)
	assert.Equal(t, 1.0, metric.Metric[0].Counter.GetValue())
	require.Len(t, metric.Metric[0].Label, 4)
	for _, label := range metric.Metric[0].Label {
		switch label.GetName() {
		case "handler":
			assert.Equal(t, handler, label.GetValue())
		case "method":
			assert.Equal(t, method, label.GetValue())
		case "code":
			assert.Equal(t, code, label.GetValue())
		case "path":
			assert.Equal(t, path, label.GetValue())
		}
	}

}
func CheckDuration(t *testing.T, metric *io_prometheus_client.MetricFamily, handler, method, path string) {
	t.Helper()
	assert.Equal(t, io_prometheus_client.MetricType_HISTOGRAM, metric.GetType())
	require.Len(t, metric.Metric, 1)
	assert.Equal(t, uint64(1), metric.Metric[0].Histogram.GetSampleCount())
	require.Len(t, metric.Metric[0].Label, 3)
	for _, label := range metric.Metric[0].Label {
		switch label.GetName() {
		case "handler":
			assert.Equal(t, handler, label.GetValue())
		case "method":
			assert.Equal(t, method, label.GetValue())
		case "path":
			assert.Equal(t, path, label.GetValue())
		default:
			t.Fatalf("unexpected label for duration metric: %s", label)
		}
	}
}
