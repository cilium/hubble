// Copyright 2020 Authors of Hubble
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

package server

import (
	"context"
	"testing"

	"github.com/cilium/cilium/pkg/monitor"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	pb "github.com/cilium/hubble/api/v1/flow"
	"github.com/cilium/hubble/api/v1/observer"
	"github.com/cilium/hubble/pkg/logger"
	"github.com/cilium/hubble/pkg/parser"
	"github.com/cilium/hubble/pkg/testutils"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalServer(t *testing.T) {
	pp, err := parser.New(
		&testutils.NoopEndpointGetter,
		&testutils.NoopIdentityGetter,
		&testutils.NoopDNSGetter,
		&testutils.NoopIPGetter,
		&testutils.NoopServiceGetter)
	require.NoError(t, err)
	s := NewLocalServer(pp, 10, 0, logger.GetLogger())
	assert.NotNil(t, s.GetStopped())
	assert.NotNil(t, s.GetPayloadParser())
	assert.NotNil(t, s.GetRingBuffer())
	assert.NotNil(t, s.GetLogger())
	assert.NotNil(t, s.GetEventsChannel())
}

func TestLocalObserverServer_ServerStatus(t *testing.T) {
	pp, err := parser.New(
		&testutils.NoopEndpointGetter,
		&testutils.NoopIdentityGetter,
		&testutils.NoopDNSGetter,
		&testutils.NoopIPGetter,
		&testutils.NoopServiceGetter)
	require.NoError(t, err)
	s := NewLocalServer(pp, 1, 0, logger.GetLogger())
	res, err := s.ServerStatus(context.Background(), &observer.ServerStatusRequest{})
	require.NoError(t, err)
	assert.Equal(t, &observer.ServerStatusResponse{NumFlows: 0, MaxFlows: 2}, res)
}

func TestLocalObserverServer_GetFlows(t *testing.T) {
	numFlows := 100
	queueSize := 0
	req := &observer.GetFlowsRequest{Number: uint64(10)}
	i := 0
	fakeServer := &FakeGetFlowsServer{
		OnSend: func(response *observer.GetFlowsResponse) error {
			i++
			return nil
		},
		FakeGRPCServerStream: &FakeGRPCServerStream{
			OnContext: func() context.Context {
				return context.Background()
			},
		},
	}
	pp, err := parser.New(
		&testutils.NoopEndpointGetter,
		&testutils.NoopIdentityGetter,
		&testutils.NoopDNSGetter,
		&testutils.NoopIPGetter,
		&testutils.NoopServiceGetter)
	require.NoError(t, err)
	s := NewLocalServer(pp, numFlows, queueSize, logger.GetLogger())
	go s.Start()

	m := s.GetEventsChannel()
	for i := 0; i < numFlows; i++ {
		tn := monitor.TraceNotifyV0{Type: byte(monitorAPI.MessageTypeTrace)}
		data := testutils.MustCreateL3L4Payload(tn)
		pl := &pb.Payload{
			Time: &types.Timestamp{Seconds: int64(i)},
			Type: pb.EventType_EventSample,
			Data: data,
		}
		m <- pl
	}
	close(s.GetEventsChannel())
	<-s.GetStopped()
	err = s.GetFlows(req, fakeServer)
	assert.NoError(t, err)
	assert.Equal(t, req.Number, uint64(i))
}
