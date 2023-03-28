// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

package observe

import (
	"context"
	"io"
	"strings"
	"testing"

	flowpb "github.com/cilium/cilium/api/v1/flow"
	observerpb "github.com/cilium/cilium/api/v1/observer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_getFlowsBasic(t *testing.T) {
	flows := []*observerpb.GetFlowsResponse{{}, {}, {}}
	var flowStrings []string
	for _, f := range flows {
		b, err := f.MarshalJSON()
		assert.NoError(t, err)
		flowStrings = append(flowStrings, string(b))
	}
	server := NewIOReaderObserver(strings.NewReader(strings.Join(flowStrings, "\n") + "\n"))
	req := observerpb.GetFlowsRequest{}
	client, err := server.GetFlows(context.Background(), &req)
	assert.NoError(t, err)
	for i := 0; i < len(flows); i++ {
		_, err = client.Recv()
		assert.NoError(t, err)
	}
	_, err = client.Recv()
	assert.Equal(t, io.EOF, err)
}

func Test_getFlowsTimeRange(t *testing.T) {
	flows := []*observerpb.GetFlowsResponse{
		{
			ResponseTypes: &observerpb.GetFlowsResponse_Flow{Flow: &flowpb.Flow{Verdict: flowpb.Verdict_FORWARDED}},
			Time:          &timestamppb.Timestamp{Seconds: 0},
		},
		{
			ResponseTypes: &observerpb.GetFlowsResponse_Flow{Flow: &flowpb.Flow{Verdict: flowpb.Verdict_DROPPED}},
			Time:          &timestamppb.Timestamp{Seconds: 100},
		},
		{
			ResponseTypes: &observerpb.GetFlowsResponse_Flow{Flow: &flowpb.Flow{Verdict: flowpb.Verdict_ERROR}},
			Time:          &timestamppb.Timestamp{Seconds: 200},
		},
	}
	var flowStrings []string
	for _, f := range flows {
		b, err := f.MarshalJSON()
		assert.NoError(t, err)
		flowStrings = append(flowStrings, string(b))
	}
	server := NewIOReaderObserver(strings.NewReader(strings.Join(flowStrings, "\n") + "\n"))
	req := observerpb.GetFlowsRequest{
		Since: &timestamppb.Timestamp{Seconds: 50},
		Until: &timestamppb.Timestamp{Seconds: 150},
	}
	client, err := server.GetFlows(context.Background(), &req)
	assert.NoError(t, err)
	res, err := client.Recv()
	assert.NoError(t, err)
	assert.Equal(t, flows[1], res)
	_, err = client.Recv()
	assert.Equal(t, io.EOF, err)
}

func Test_getFlowsFilter(t *testing.T) {
	flows := []*observerpb.GetFlowsResponse{
		{
			ResponseTypes: &observerpb.GetFlowsResponse_Flow{Flow: &flowpb.Flow{Verdict: flowpb.Verdict_FORWARDED}},
			Time:          &timestamppb.Timestamp{Seconds: 0},
		},
		{
			ResponseTypes: &observerpb.GetFlowsResponse_Flow{Flow: &flowpb.Flow{Verdict: flowpb.Verdict_DROPPED}},
			Time:          &timestamppb.Timestamp{Seconds: 100},
		},
		{
			ResponseTypes: &observerpb.GetFlowsResponse_Flow{Flow: &flowpb.Flow{Verdict: flowpb.Verdict_ERROR}},
			Time:          &timestamppb.Timestamp{Seconds: 200},
		},
	}
	var flowStrings []string
	for _, f := range flows {
		b, err := f.MarshalJSON()
		assert.NoError(t, err)
		flowStrings = append(flowStrings, string(b))
	}
	server := NewIOReaderObserver(strings.NewReader(strings.Join(flowStrings, "\n") + "\n"))
	req := observerpb.GetFlowsRequest{
		Whitelist: []*flowpb.FlowFilter{
			{
				Verdict: []flowpb.Verdict{flowpb.Verdict_FORWARDED, flowpb.Verdict_ERROR},
			},
		},
	}
	client, err := server.GetFlows(context.Background(), &req)
	assert.NoError(t, err)
	res, err := client.Recv()
	assert.NoError(t, err)
	assert.Equal(t, flows[0], res)
	res, err = client.Recv()
	assert.NoError(t, err)
	assert.Equal(t, flows[2], res)
	_, err = client.Recv()
	assert.Equal(t, io.EOF, err)
}
