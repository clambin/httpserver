package httpserver_test

import (
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"testing"
)

type metricInfo struct {
	metric *io_prometheus_client.Metric
	labels map[string]string
}

func getMetricInfo(t *testing.T, g prometheus.Gatherer, name string) (output []metricInfo) {
	t.Helper()

	metrics, err := g.Gather()
	require.NoError(t, err)

	for _, metric := range metrics {
		if metric.GetName() != name {
			continue
		}
		for _, m := range metric.GetMetric() {
			info := metricInfo{
				metric: m,
				labels: make(map[string]string),
			}
			for _, l := range m.GetLabel() {
				info.labels[l.GetName()] = l.GetValue()
			}
			output = append(output, info)
		}
	}
	return output
}
