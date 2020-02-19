package flow

import (
	"testing"

	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	pb "github.com/cilium/hubble/api/v1/flow"
	"github.com/cilium/hubble/pkg/metrics/api"
	"github.com/cilium/hubble/pkg/testutils"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlowHandler(t *testing.T) {
	registry := prometheus.NewRegistry()
	opts := api.Options{"sourceContext": "namespace", "destinationContext": "namespace"}

	h := &flowHandler{}

	t.Run("Init", func(t *testing.T) {
		require.NoError(t, h.Init(registry, opts))
	})

	t.Run("Status", func(t *testing.T) {
		require.Equal(t, "destination=namespace,source=namespace", h.Status())
	})

	t.Run("ProcessFlow", func(t *testing.T) {
		flow := &testutils.FakeFlow{
			EventType: &pb.CiliumEventType{Type: monitorAPI.MessageTypeAccessLog},
			L7: &pb.Layer7{
				Record: &pb.Layer7_Http{Http: &pb.HTTP{}},
			},
			Source:      &pb.Endpoint{Namespace: "foo"},
			Destination: &pb.Endpoint{Namespace: "bar"},
			Verdict:     pb.Verdict_FORWARDED,
		}
		h.ProcessFlow(flow)

		metricFamilies, err := registry.Gather()
		require.NoError(t, err)
		require.Len(t, metricFamilies, 1)

		assert.Equal(t, "hubble_flows_processed_total", *metricFamilies[0].Name)
		require.Len(t, metricFamilies[0].Metric, 1)
		metric := metricFamilies[0].Metric[0]

		assert.Equal(t, "destination", *metric.Label[0].Name)
		assert.Equal(t, "bar", *metric.Label[0].Value)

		assert.Equal(t, "source", *metric.Label[1].Name)
		assert.Equal(t, "foo", *metric.Label[1].Value)

		assert.Equal(t, "subtype", *metric.Label[2].Name)
		assert.Equal(t, "HTTP", *metric.Label[2].Value)

		assert.Equal(t, "type", *metric.Label[3].Name)
		assert.Equal(t, "L7", *metric.Label[3].Value)

		assert.Equal(t, "verdict", *metric.Label[4].Name)
		assert.Equal(t, "FORWARDED", *metric.Label[4].Value)
	})
}
