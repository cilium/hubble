// Copyright 2019 Authors of Hubble
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flow

import (
	pb "github.com/cilium/hubble/api/v1/flow"
	"github.com/cilium/hubble/pkg/metrics/api"

	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	"github.com/prometheus/client_golang/prometheus"
)

type flowHandler struct {
	flows *prometheus.CounterVec
}

func (h *flowHandler) Init(registry *prometheus.Registry, options api.Options) error {
	h.flows = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: api.DefaultPrometheusNamespace,
		Name:      "flows_processed_total",
		Help:      "Total number of flows processed",
	}, []string{"type", "subtype", "verdict"})

	registry.MustRegister(h.flows)
	return nil
}

func (h *flowHandler) Status() string {
	return ""
}

func (h *flowHandler) ProcessFlow(flow *pb.Flow) {
	var typeName, subType string
	switch flow.EventType.Type {
	case monitorAPI.MessageTypeAgent:
		typeName = "Agent"
	case monitorAPI.MessageTypeAccessLog:
		typeName = "L7"
		if l7 := flow.GetL7(); l7 != nil {
			switch {
			case l7.GetDns() != nil:
				subType = "DNS"
			case l7.GetHttp() != nil:
				subType = "HTTP"
			case l7.GetKafka() != nil:
				subType = "Kafka"
			}
		}

	case monitorAPI.MessageTypeDrop:
		typeName = "Drop"
	case monitorAPI.MessageTypeDebug:
		typeName = "Debug"
	case monitorAPI.MessageTypeCapture:
		typeName = "Capture"
	case monitorAPI.MessageTypeTrace:
		typeName = "Trace"
		subType = monitorAPI.TraceObservationPoints[uint8(flow.EventType.SubType)]
	}

	h.flows.WithLabelValues(typeName, subType, flow.Verdict.String()).Inc()
}
