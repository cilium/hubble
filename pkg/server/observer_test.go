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

package server

import (
	"context"
	"net"
	"testing"

	"github.com/cilium/cilium/pkg/monitor"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"

	pb "github.com/cilium/hubble/api/v1/observer"
	v1 "github.com/cilium/hubble/pkg/api/v1"
	"github.com/cilium/hubble/pkg/fqdncache"
	"github.com/cilium/hubble/pkg/ipcache"
	"github.com/cilium/hubble/pkg/parser"
	"github.com/cilium/hubble/pkg/testutils"

	"github.com/gogo/protobuf/types"
	"github.com/google/gopacket/layers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

var allTypes = []*pb.EventTypeFilter{
	{Type: 1},
	{Type: 2},
	{Type: 3},
	{Type: 4},
	{Type: 129},
	{Type: 130},
}

type FakeGRPCServerStream struct {
	OnSetHeader  func(metadata.MD) error
	OnSendHeader func(metadata.MD) error
	OnSetTrailer func(m metadata.MD)
	OnContext    func() context.Context
	OnSendMsg    func(m interface{}) error
	OnRecvMsg    func(m interface{}) error
}

type FakeGetFlowsServer struct {
	OnSend func(response *pb.GetFlowsResponse) error
	*FakeGRPCServerStream
}

func (s *FakeGetFlowsServer) Send(response *pb.GetFlowsResponse) error {
	if s.OnSend != nil {
		// TODO: completely convert this into using pb.Flow
		return s.OnSend(response)
	}
	panic("OnSend not set")
}

func (s *FakeGRPCServerStream) SetHeader(m metadata.MD) error {
	if s.OnSetHeader != nil {
		return s.OnSetHeader(m)
	}
	panic("OnSetHeader not set")
}

func (s *FakeGRPCServerStream) SendHeader(m metadata.MD) error {
	if s.OnSendHeader != nil {
		return s.OnSendHeader(m)
	}
	panic("OnSendHeader not set")
}

func (s *FakeGRPCServerStream) SetTrailer(m metadata.MD) {
	if s.OnSetTrailer != nil {
		s.OnSetTrailer(m)
	}
	panic("OnSetTrailer not set")
}

func (s *FakeGRPCServerStream) Context() context.Context {
	if s.OnContext != nil {
		return s.OnContext()
	}
	panic("OnContext not set")
}

func (s *FakeGRPCServerStream) SendMsg(m interface{}) error {
	if s.OnSendMsg != nil {
		return s.OnSendMsg(m)
	}
	panic("OnSendMsg not set")
}

func (s *FakeGRPCServerStream) RecvMsg(m interface{}) error {
	if s.OnRecvMsg != nil {
		return s.OnRecvMsg(m)
	}
	panic("OnRecvMsg not set")
}

func TestObserverServer_GetLastNFlows(t *testing.T) {
	es := v1.NewEndpoints()
	ipc := ipcache.New()
	fqdnc := fqdncache.New()

	pp, err := parser.New(es, fakeDummyCiliumClient, fqdnc, ipc)
	assert.NoError(t, err)

	s := NewServer(fakeDummyCiliumClient, es, ipc, fqdnc, pp, 0xff)
	if s.ring.Cap() != 0x100 {
		t.Errorf("s.ring.Len() got = %#v, want %#v", s.ring.Cap(), 0x100)
	}
	go s.Start()

	m := s.GetEventsChannel()
	want := make([]*pb.Payload, 10, 10)
	for i := uint64(0); i < s.ring.Cap(); i++ {
		pl := &pb.Payload{
			Time: &types.Timestamp{Seconds: int64(i)},
			Type: pb.EventType_EventSample,
			Data: []byte{
				0: 1,       // event type
				1: byte(i), // other data
			},
		}
		m <- pl

		if i >= s.ring.Cap()-10-1 && i < s.ring.Cap()-1 {
			// store the last 10 flows that we have written.
			want[i-(s.ring.Cap()-10-1)] = pl
		}
	}
	// Make sure all flows were consumed by the server
	close(m)
	<-s.stopped

	// We could use s.ring.LastWrite() but the Server uses LastWriteParallel
	// so we should use LastWriteParallel in testing as well
	if lastWrite := s.ring.LastWriteParallel(); lastWrite != 0xfe {
		t.Errorf("LastWriteParallel() returns = %v, want %v", lastWrite, 0xfe)
	}

	req := &pb.GetFlowsRequest{
		Number:       10,
		Whitelist:    []*pb.FlowFilter{{EventType: allTypes}},
		SkipDecoding: true,
	}
	got := make([]*pb.GetFlowsResponse, 10, 10)
	i := 0
	fakeServer := &FakeGetFlowsServer{
		OnSend: func(response *pb.GetFlowsResponse) error {
			got[i] = response
			i++
			return nil
		},
		FakeGRPCServerStream: &FakeGRPCServerStream{
			OnContext: func() context.Context {
				return context.Background()
			},
		},
	}
	err = s.GetFlows(req, fakeServer)
	if err != nil {
		t.Errorf("GetLastNFlows error = %v, wantErr %v", err, nil)
	}

	if len(got) != len(want) {
		t.Errorf("Length of 'got' is not the same as 'wanted'")
	}
	for i := 0; i < 10; i++ {
		assert.Equal(t, want[i], got[i].GetFlow().Payload)
	}
}

func TestObserverServer_GetLastNFlows_With_Follow(t *testing.T) {
	es := v1.NewEndpoints()
	ipc := ipcache.New()
	fqdnc := fqdncache.New()

	pp, err := parser.New(es, fakeDummyCiliumClient, fqdnc, ipc)
	assert.NoError(t, err)

	s := NewServer(fakeDummyCiliumClient, es, ipc, fqdnc, pp, 0xff)
	if s.ring.Cap() != 0x100 {
		t.Errorf("s.ring.Len() got = %#v, want %#v", s.ring.Cap(), 0x100)
	}
	go s.Start()

	m := s.GetEventsChannel()
	want := make([]*pb.Payload, 12, 12)
	for i := uint64(0); i < s.ring.Cap(); i++ {
		pl := &pb.Payload{
			Time: &types.Timestamp{Seconds: int64(i)},
			Type: pb.EventType_EventSample,
			Data: []byte{
				0: 1,       // event type
				1: byte(i), // other data
			},
		}
		m <- pl

		if i >= s.ring.Cap()-10-1 {
			// store the last 11 flows that we have written.
			want[i-(s.ring.Cap()-10-1)] = pl
		}
	}
	// Make sure all flows were consumed by the server
	close(m)
	<-s.stopped

	// We could use s.ring.LastWrite() but the Server uses LastWriteParallel
	// so we should use LastWriteParallel in testing as well
	if lastWrite := s.ring.LastWriteParallel(); lastWrite != 0xfe {
		t.Errorf("LastWriteParallel() returns = %v, want %v", lastWrite, 0xfe)
	}

	req := &pb.GetFlowsRequest{
		Number:       10,
		Whitelist:    []*pb.FlowFilter{{EventType: allTypes}},
		Follow:       true,
		SkipDecoding: true,
	}
	got := make([]*pb.GetFlowsResponse, 12, 12)
	i := 0
	receivedFirstBatch, receivedSecondBatch := make(chan struct{}), make(chan struct{})
	fakeServer := &FakeGetFlowsServer{
		OnSend: func(response *pb.GetFlowsResponse) error {
			got[i] = response
			i++
			if i == 10 {
				close(receivedFirstBatch)
			}
			if i == 12 {
				close(receivedSecondBatch)
			}
			return nil
		},
		FakeGRPCServerStream: &FakeGRPCServerStream{
			OnContext: func() context.Context {
				return context.Background()
			},
		},
	}
	go func() {
		err := s.GetFlows(req, fakeServer)
		if err != nil {
			t.Errorf("GetLastNFlows error = %v, wantErr %v", err, nil)
		}
	}()
	<-receivedFirstBatch

	for i := 0; i < 10; i++ {
		assert.Equal(t, want[i], got[i].GetFlow().Payload)
	}

	// hacky to restart the events consumer.
	s.events = make(chan *pb.Payload, 10)
	go s.Start()
	m = s.GetEventsChannel()

	for i := uint64(0); i < 2; i++ {
		pl := &pb.Payload{
			Time: &types.Timestamp{Seconds: int64(i + s.ring.Cap())},
			Type: pb.EventType_EventSample,
			Data: []byte{
				0: 2,                      // event type
				1: byte(i + s.ring.Cap()), // other data
			},
		}
		m <- pl
		if i < 1 {
			// store the second-last flow, the client will never be able to read
			// the last flow. Check s.ring.LastWriteParallel() docs for why.
			want[i+11] = pl
		}
	}

	<-receivedSecondBatch
	for i := 0; i < len(got); i++ {
		assert.Equal(t, want[i], got[i].GetFlow().Payload)
	}
}

func TestObserverServer_GetFlowsBetween(t *testing.T) {
	es := v1.NewEndpoints()
	ipc := ipcache.New()
	fqdnc := fqdncache.New()

	pp, err := parser.New(es, fakeDummyCiliumClient, fqdnc, ipc)
	assert.NoError(t, err)

	s := NewServer(fakeDummyCiliumClient, es, ipc, fqdnc, pp, 0xff)
	if s.ring.Cap() != 0x100 {
		t.Errorf("s.ring.Len() got = %#v, want %#v", s.ring.Cap(), 0x100)
	}
	go s.Start()

	m := s.GetEventsChannel()
	for i := uint64(0); i < s.ring.Cap(); i++ {
		m <- &pb.Payload{
			Time: &types.Timestamp{Seconds: int64(i)},
			Type: pb.EventType_EventSample,
			Data: []byte{
				0: 1,       // event type
				1: byte(i), // other data
			},
		}
	}
	// Make sure all flows were consumed by the server
	close(m)
	<-s.stopped

	// We could use s.ring.LastWrite() but the Server uses LastWriteParallel
	// so we should use LastWriteParallel in testing as well
	if lastWrite := s.ring.LastWriteParallel(); lastWrite != 0xfe {
		t.Errorf("LastWriteParallel() returns = %v, want %v", lastWrite, 0xfe)
	}

	req := &pb.GetFlowsRequest{
		Since:        &types.Timestamp{Seconds: 2, Nanos: 0},
		Until:        &types.Timestamp{Seconds: 7, Nanos: 0},
		Whitelist:    []*pb.FlowFilter{{EventType: allTypes}},
		SkipDecoding: true,
	}
	want := []*pb.Payload{
		{
			Time: &types.Timestamp{Seconds: int64(6)},
			Type: pb.EventType_EventSample,
			Data: []byte{
				0: 1,       // event type
				1: byte(6), // other data
			},
		},
		{
			Time: &types.Timestamp{Seconds: int64(5)},
			Type: pb.EventType_EventSample,
			Data: []byte{
				0: 1,       // event type
				1: byte(5), // other data
			},
		},
		{
			Time: &types.Timestamp{Seconds: int64(4)},
			Type: pb.EventType_EventSample,
			Data: []byte{
				0: 1,       // event type
				1: byte(4), // other data
			},
		},
		{
			Time: &types.Timestamp{Seconds: int64(3)},
			Type: pb.EventType_EventSample,
			Data: []byte{
				0: 1,       // event type
				1: byte(3), // other data
			},
		},
	}
	got := make([]*pb.GetFlowsResponse, 4, 4)
	i := 0
	fakeServer := &FakeGetFlowsServer{
		OnSend: func(response *pb.GetFlowsResponse) error {
			got[i] = response
			i++
			return nil
		},
		FakeGRPCServerStream: &FakeGRPCServerStream{
			OnContext: func() context.Context {
				return context.Background()
			},
		},
	}
	err = s.GetFlows(req, fakeServer)
	if err != nil {
		t.Errorf("GetFlowsBetween error = %v, wantErr %v", err, nil)
	}

	for i := 0; i < 4; i++ {
		assert.Equal(t, want[i], got[i].GetFlow().GetPayload())
	}
}

type FakeObserverGetFlowsServer struct {
	OnSend func(*pb.GetFlowsResponse) error
	*FakeGRPCServerStream
}

func (s *FakeObserverGetFlowsServer) Send(flow *pb.GetFlowsResponse) error {
	if s.OnSend != nil {
		return s.OnSend(flow)
	}
	panic("OnSend not set")
}

func TestObserverServer_GetFlows(t *testing.T) {
	numFlows := 10
	count := 0
	fakeServer := &FakeObserverGetFlowsServer{
		OnSend: func(_ *pb.GetFlowsResponse) error {
			count++
			return nil
		},
		FakeGRPCServerStream: &FakeGRPCServerStream{
			OnContext: func() context.Context {
				return context.Background()
			},
		},
	}
	es := v1.NewEndpoints()
	ipc := ipcache.New()
	fqdnc := fqdncache.New()

	pp, err := parser.New(es, fakeDummyCiliumClient, fqdnc, ipc)
	assert.NoError(t, err)

	s := NewServer(fakeDummyCiliumClient, es, ipc, fqdnc, pp, 30)
	go s.Start()
	m := s.GetEventsChannel()
	eth := layers.Ethernet{
		EthernetType: layers.EthernetTypeIPv4,
		SrcMAC:       net.HardwareAddr{1, 2, 3, 4, 5, 6},
		DstMAC:       net.HardwareAddr{1, 2, 3, 4, 5, 6},
	}
	ip := layers.IPv4{
		SrcIP: net.ParseIP("1.1.1.1"),
		DstIP: net.ParseIP("2.2.2.2"),
	}
	tcp := layers.TCP{}
	ch := s.GetEventsChannel()
	for i := 0; i < numFlows; i++ {
		data, err := testutils.CreateL3L4Payload(monitor.DropNotify{Type: monitorAPI.MessageTypeDrop}, &eth, &ip, &tcp)
		require.NoError(t, err)
		m <- &pb.Payload{Type: pb.EventType_EventSample, Data: data}
		// This payload will be ignored by GetFlows.
		data, err = testutils.CreateL3L4Payload(monitorAPI.AgentNotify{Type: monitorAPI.MessageTypeAgent}, &eth, &ip, &tcp)
		require.NoError(t, err)
		m <- &pb.Payload{Type: pb.EventType_EventSample, Data: data}
	}
	close(ch)
	<-s.stopped
	err = s.GetFlows(&pb.GetFlowsRequest{Number: 20}, fakeServer)
	assert.NoError(t, err)
	assert.Equal(t, numFlows, count)
}

func TestObserverServer_GetFlowsWithFilters(t *testing.T) {
	numFlows := 10
	count := 0
	fakeServer := &FakeObserverGetFlowsServer{
		OnSend: func(res *pb.GetFlowsResponse) error {
			count++
			assert.Equal(t, "1.1.1.1", res.GetFlow().GetIP().GetSource())
			assert.Equal(t, "2.2.2.2", res.GetFlow().GetIP().GetDestination())
			assert.Equal(t, uint8(monitorAPI.MessageTypeDrop), res.GetFlow().GetPayload().GetData()[0])
			return nil
		},
		FakeGRPCServerStream: &FakeGRPCServerStream{
			OnContext: func() context.Context {
				return context.Background()
			},
		},
	}

	es := v1.NewEndpoints()
	ipc := ipcache.New()
	fqdnc := fqdncache.New()

	pp, err := parser.New(es, fakeDummyCiliumClient, fqdnc, ipc)
	assert.NoError(t, err)

	s := NewServer(fakeDummyCiliumClient, es, ipc, fqdnc, pp, 30)
	go s.Start()
	m := s.GetEventsChannel()
	eth := layers.Ethernet{
		EthernetType: layers.EthernetTypeIPv4,
		SrcMAC:       net.HardwareAddr{1, 2, 3, 4, 5, 6},
		DstMAC:       net.HardwareAddr{1, 2, 3, 4, 5, 6},
	}
	ip := layers.IPv4{
		SrcIP: net.ParseIP("1.1.1.1"),
		DstIP: net.ParseIP("2.2.2.2"),
	}
	ipRev := layers.IPv4{
		SrcIP: net.ParseIP("2.2.2.2"),
		DstIP: net.ParseIP("1.1.1.1"),
	}
	tcp := layers.TCP{}
	udp := layers.UDP{}
	ch := s.GetEventsChannel()
	for i := 0; i < numFlows; i++ {
		// flow which is matched by the whitelist (to be included)
		data, err := testutils.CreateL3L4Payload(monitor.DropNotify{Type: monitorAPI.MessageTypeDrop}, &eth, &ip, &tcp)
		require.NoError(t, err)
		m <- &pb.Payload{Type: pb.EventType_EventSample, Data: data}
		// flow which is neither matched by the whitelist nor blacklist (to be ignored)
		data, err = testutils.CreateL3L4Payload(monitor.DropNotify{Type: monitorAPI.MessageTypeDrop}, &eth, &ipRev, &tcp)
		require.NoError(t, err)
		m <- &pb.Payload{Type: pb.EventType_EventSample, Data: data}
		// flows which is matched by both the white- and blacklist (to be excluded)
		data, err = testutils.CreateL3L4Payload(monitor.TraceNotifyV0{Type: monitorAPI.MessageTypeTrace}, &eth, &ip, &udp)
		require.NoError(t, err)
		m <- &pb.Payload{Type: pb.EventType_EventSample, Data: data}
	}
	close(ch)
	<-s.stopped
	err = s.GetFlows(&pb.GetFlowsRequest{
		Number: uint64(numFlows * 3),
		Whitelist: []*pb.FlowFilter{
			{SourceIp: []string{"1.1.1.1"}, EventType: allTypes},
		},
		Blacklist: []*pb.FlowFilter{
			{EventType: []*pb.EventTypeFilter{{Type: monitorAPI.MessageTypeTrace}}},
		},
	}, fakeServer)
	assert.NoError(t, err)
	assert.Equal(t, numFlows, count)
}

func TestObserverServer_GetFlowsOfANonLocalPod(t *testing.T) {
	numFlows := 5
	count := 0
	fakeServer := &FakeObserverGetFlowsServer{
		OnSend: func(_ *pb.GetFlowsResponse) error {
			count++
			return nil
		},
		FakeGRPCServerStream: &FakeGRPCServerStream{
			OnContext: func() context.Context {
				return context.Background()
			},
		},
	}
	fakeIPGetter := &testutils.FakeIPGetter{
		OnGetIPIdentity: func(ip net.IP) (identity ipcache.IPIdentity, ok bool) {
			if ip.Equal(net.ParseIP("1.1.1.1")) {
				return ipcache.IPIdentity{Namespace: "default", PodName: "foo-bar"}, true
			}
			return ipcache.IPIdentity{}, false
		},
	}

	es := v1.NewEndpoints()
	ipc := ipcache.New()
	fqdnc := fqdncache.New()

	pp, err := parser.New(es, fakeDummyCiliumClient, fqdnc, fakeIPGetter)
	assert.NoError(t, err)

	s := NewServer(fakeDummyCiliumClient, es, ipc, fqdnc, pp, 30)
	go s.Start()
	m := s.GetEventsChannel()
	eth := layers.Ethernet{
		EthernetType: layers.EthernetTypeIPv4,
		SrcMAC:       net.HardwareAddr{1, 2, 3, 4, 5, 6},
		DstMAC:       net.HardwareAddr{1, 2, 3, 4, 5, 6},
	}
	ip := layers.IPv4{
		SrcIP: net.ParseIP("1.1.1.1"),
		DstIP: net.ParseIP("2.2.2.2"),
	}
	tcp := layers.TCP{}
	for i := 0; i < numFlows; i++ {
		data, err := testutils.CreateL3L4Payload(monitor.DropNotify{Type: monitorAPI.MessageTypeDrop}, &eth, &ip, &tcp)
		require.NoError(t, err)
		m <- &pb.Payload{Type: pb.EventType_EventSample, Data: data}
		// This payload will be ignored by GetFlows.
		data, err = testutils.CreateL3L4Payload(monitorAPI.AgentNotify{Type: monitorAPI.MessageTypeAgent}, &eth, &ip, &tcp)
		require.NoError(t, err)
		m <- &pb.Payload{Type: pb.EventType_EventSample, Data: data}
	}
	close(m)
	<-s.stopped

	// pod exist so we will be able to get flows
	flowFilter := []*pb.FlowFilter{
		{
			SourcePod: []string{"default/foo-bar"},
			EventType: []*pb.EventTypeFilter{
				{
					Type: monitorAPI.MessageTypeDrop,
				},
			},
		},
	}
	err = s.GetFlows(&pb.GetFlowsRequest{Whitelist: flowFilter, Number: 20}, fakeServer)
	assert.NoError(t, err)
	assert.Equal(t, numFlows, count)
}
