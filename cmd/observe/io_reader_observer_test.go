// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of Hubble

package observe

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/cilium/cilium/api/v1/flow"
	"github.com/cilium/cilium/api/v1/observer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_getFlowsBasic(t *testing.T) {
	flows := []*observer.GetFlowsResponse{{}, {}, {}}
	var flowStrings []string
	for _, f := range flows {
		b, err := f.MarshalJSON()
		assert.NoError(t, err)
		flowStrings = append(flowStrings, string(b))
	}
	server := newIOReaderObserver(strings.NewReader(strings.Join(flowStrings, "\n") + "\n"))
	req := observer.GetFlowsRequest{}
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
	flows := []*observer.GetFlowsResponse{
		{
			ResponseTypes: &observer.GetFlowsResponse_Flow{Flow: &flow.Flow{Verdict: flow.Verdict_FORWARDED}},
			Time:          &timestamppb.Timestamp{Seconds: 0},
		},
		{
			ResponseTypes: &observer.GetFlowsResponse_Flow{Flow: &flow.Flow{Verdict: flow.Verdict_DROPPED}},
			Time:          &timestamppb.Timestamp{Seconds: 100},
		},
		{
			ResponseTypes: &observer.GetFlowsResponse_Flow{Flow: &flow.Flow{Verdict: flow.Verdict_ERROR}},
			Time:          &timestamppb.Timestamp{Seconds: 200},
		},
	}
	var flowStrings []string
	for _, f := range flows {
		b, err := f.MarshalJSON()
		assert.NoError(t, err)
		flowStrings = append(flowStrings, string(b))
	}
	server := newIOReaderObserver(strings.NewReader(strings.Join(flowStrings, "\n") + "\n"))
	req := observer.GetFlowsRequest{
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
	flows := []*observer.GetFlowsResponse{
		{
			ResponseTypes: &observer.GetFlowsResponse_Flow{Flow: &flow.Flow{Verdict: flow.Verdict_FORWARDED}},
			Time:          &timestamppb.Timestamp{Seconds: 0},
		},
		{
			ResponseTypes: &observer.GetFlowsResponse_Flow{Flow: &flow.Flow{Verdict: flow.Verdict_DROPPED}},
			Time:          &timestamppb.Timestamp{Seconds: 100},
		},
		{
			ResponseTypes: &observer.GetFlowsResponse_Flow{Flow: &flow.Flow{Verdict: flow.Verdict_ERROR}},
			Time:          &timestamppb.Timestamp{Seconds: 200},
		},
	}
	var flowStrings []string
	for _, f := range flows {
		b, err := f.MarshalJSON()
		assert.NoError(t, err)
		flowStrings = append(flowStrings, string(b))
	}
	server := newIOReaderObserver(strings.NewReader(strings.Join(flowStrings, "\n") + "\n"))
	req := observer.GetFlowsRequest{
		Whitelist: []*flow.FlowFilter{
			{
				Verdict: []flow.Verdict{flow.Verdict_FORWARDED, flow.Verdict_ERROR},
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
