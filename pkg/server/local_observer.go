// Copyright 2020 Authors of Cilium
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

	"github.com/cilium/cilium/pkg/math"
	pb "github.com/cilium/hubble/api/v1/flow"
	"github.com/cilium/hubble/api/v1/observer"
	"github.com/cilium/hubble/pkg/container"
	"github.com/cilium/hubble/pkg/parser"
	"go.uber.org/zap"
)

// LocalObserverServer is an implementation of the server.Observer interface
// that's meant to be run embedded inside the Cilium process. It ignores all
// the state change events since the state is available locally.
type LocalObserverServer struct {
	// ring buffer that contains the references of all flows
	ring *container.Ring

	// events is the channel used by the writer(s) to send the flow data
	// into the observer server.
	events chan *pb.Payload

	// stopped is mostly used in unit tests to signalize when the events
	// channel is empty, once it's closed.
	stopped chan struct{}

	log *zap.Logger

	// channel to receive events from observer server.
	eventschan chan *observer.GetFlowsResponse

	// payloadParser decodes pb.Payload into pb.Flow
	payloadParser *parser.Parser
}

// NewLocalServer returns a new local observer server.
func NewLocalServer(
	payloadParser *parser.Parser,
	maxFlows int,
	logger *zap.Logger,
) *LocalObserverServer {
	return &LocalObserverServer{
		log:  logger,
		ring: container.NewRing(maxFlows),
		// have a channel with 1% of the max flows that we can receive
		events:        make(chan *pb.Payload, uint64(math.IntMin(maxFlows/100, 100))),
		stopped:       make(chan struct{}),
		eventschan:    make(chan *observer.GetFlowsResponse, 100),
		payloadParser: payloadParser,
	}
}

// Start starts the server to handle the events sent to the events channel as
// well as handle events to the EpAdd and EpDel channels.
func (s *LocalObserverServer) Start() {
	processEvents(s)
}

// GetEventsChannel returns the event channel to receive pb.Payload events.
func (s *LocalObserverServer) GetEventsChannel() chan *pb.Payload {
	return s.events
}

// GetRingBuffer implements Observer.GetRingBuffer.
func (s *LocalObserverServer) GetRingBuffer() *container.Ring {
	return s.ring
}

// GetLogger implements Observer.GetLogger.
func (s *LocalObserverServer) GetLogger() *zap.Logger {
	return s.log
}

// GetStopped implements Observer.GetStopped.
func (s *LocalObserverServer) GetStopped() chan struct{} {
	return s.stopped
}

// GetPayloadParser implements Observer.GetPayloadParser.
func (s *LocalObserverServer) GetPayloadParser() *parser.Parser {
	return s.payloadParser
}

// ServerStatus should have a comment, apparently. It returns the server status.
func (s *LocalObserverServer) ServerStatus(
	ctx context.Context, req *observer.ServerStatusRequest,
) (*observer.ServerStatusResponse, error) {
	return getServerStatusFromObserver(s)
}

// GetFlows implements the proto method for client requests.
func (s *LocalObserverServer) GetFlows(
	req *observer.GetFlowsRequest,
	server observer.Observer_GetFlowsServer,
) (err error) {
	return getFlows(req, server, s)
}

// HandleMonitorSocket is a noop for local server since it doesn't connect to the monitor socket.
func (s *LocalObserverServer) HandleMonitorSocket(nodeName string) error {
	return nil
}
