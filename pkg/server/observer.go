// Copyright 2019 Authors of Cilium
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
	"fmt"
	"net"
	"strings"
	"time"

	pb "github.com/cilium/hubble/api/v1/flow"
	"github.com/cilium/hubble/api/v1/observer"
	v1 "github.com/cilium/hubble/pkg/api/v1"
	"github.com/cilium/hubble/pkg/container"
	"github.com/cilium/hubble/pkg/filters"
	"github.com/cilium/hubble/pkg/ipcache"
	"github.com/cilium/hubble/pkg/logger"
	"github.com/cilium/hubble/pkg/metrics"
	"github.com/cilium/hubble/pkg/parser"
	parserErrors "github.com/cilium/hubble/pkg/parser/errors"

	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/math"
	"github.com/cilium/cilium/pkg/monitor"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	"github.com/gogo/protobuf/types"
	"go.uber.org/zap"
)

type ciliumClient interface {
	EndpointList() ([]*models.Endpoint, error)
	GetEndpoint(id uint64) (*models.Endpoint, error)
	GetIdentity(id uint64) (*models.Identity, error)
	GetFqdnCache() ([]*models.DNSLookup, error)
	GetIPCache() ([]*models.IPListEntry, error)
}

type endpointsHandler interface {
	SyncEndpoints([]*v1.Endpoint)
	UpdateEndpoint(*v1.Endpoint)
	MarkDeleted(*v1.Endpoint)
	FindEPs(epID uint64, ns, pod string) []v1.Endpoint
	GetEndpoint(ip net.IP) (endpoint *v1.Endpoint, ok bool)
	GarbageCollect()
}

type fqdnCache interface {
	InitializeFrom(entries []*models.DNSLookup)
	AddDNSLookup(epID uint64, lookupTime time.Time, domainName string, ips []net.IP, ttl uint32)
	GetNamesOf(epID uint64, ip net.IP) []string
}

// ObserverServer is a server that can store events in memory
type ObserverServer struct {
	// ring buffer that contains the references of all flows
	ring *container.Ring

	// events is the channel used by the writer(s) to send the flow data
	// into the observer server.
	events chan *pb.Payload

	// stopped is mostly used in unit tests to signalize when the events
	// channel is empty, once it's closed.
	stopped chan struct{}

	// ciliumClient will connect to Cilium to pool cilium endpoint information
	ciliumClient ciliumClient

	// endpoints contains a slice of all endpoints running the node where
	// hubble is running.
	endpoints endpointsHandler

	// fqdnCache contains the responses of all intercepted DNS lookups
	// performed by local endpoints
	fqdnCache fqdnCache

	// ipcache is a mirror of Cilium's IPCache
	ipcache *ipcache.IPCache

	// epAdd is a channel used to exchange endpoint events from Cilium
	endpointEvents chan monitorAPI.AgentNotify

	// logRecord is a channel used to exchange L7 DNS requests seens from the
	// monitor
	logRecord chan monitor.LogRecordNotify

	log *zap.Logger

	// channel to receive events from observer server.
	eventschan chan *observer.GetFlowsResponse

	// payloadParser decodes pb.Payload into pb.Flow
	payloadParser *parser.Parser
}

// NewServer returns a server that can store up to the given of maxFlows
// received.
func NewServer(
	ciliumClient ciliumClient,
	endpoints endpointsHandler,
	ipCache *ipcache.IPCache,
	fqdnCache fqdnCache,
	payloadParser *parser.Parser,
	maxFlows int,
) *ObserverServer {

	return &ObserverServer{
		log:  logger.GetLogger(),
		ring: container.NewRing(maxFlows),
		// have a channel with 1% of the max flows that we can receive
		events:         make(chan *pb.Payload, uint64(math.IntMin(maxFlows/100, 100))),
		stopped:        make(chan struct{}),
		ciliumClient:   ciliumClient,
		endpoints:      endpoints,
		ipcache:        ipCache,
		fqdnCache:      fqdnCache,
		endpointEvents: make(chan monitorAPI.AgentNotify, 100),
		logRecord:      make(chan monitor.LogRecordNotify, 100),
		eventschan:     make(chan *observer.GetFlowsResponse, 100),
		payloadParser:  payloadParser,
	}
}

// Start starts the server to handle the events sent to the events channel as
// well as handle events to the EpAdd and EpDel channels.
func (s *ObserverServer) Start() {
	go s.syncEndpoints()
	go s.syncFQDNCache()
	go s.consumeEndpointEvents()
	go s.consumeLogRecordNotifyChannel()

	for pl := range s.events {
		flow, err := s.decodeFlow(pl)
		if err != nil {
			if !parserErrors.IsErrInvalidType(err) {
				s.log.Debug("failed to decode payload", zap.ByteString("data", pl.Data), zap.Error(err))
			}
			continue
		}

		metrics.ProcessFlow(flow)
		s.ring.Write(&v1.Event{
			Timestamp: pl.Time,
			Event:     flow,
		})
	}
	close(s.stopped)
}

// StartMirroringIPCache will obtain an initial IPCache snapshot from Cilium
// and then start mirroring IPCache events based on IPCacheNotification sent
// through the ipCacheEvents channels. Only messages of type
// `AgentNotifyIPCacheUpserted` and `AgentNotifyIPCacheDeleted` should be sent
// through that channel. This function assumes that the caller is already
// connected to Cilium Monitor, i.e. no IPCacheNotification must be lost after
// calling this method.
func (s *ObserverServer) StartMirroringIPCache(ipCacheEvents <-chan monitorAPI.AgentNotify) {
	go s.syncIPCache(ipCacheEvents)
}

// GetLogRecordNotifyChannel returns the event channel to receive
// monitorAPI.LogRecordNotify events.
func (s *ObserverServer) GetLogRecordNotifyChannel() chan<- monitor.LogRecordNotify {
	return s.logRecord
}

// GetEventsChannel returns the event channel to receive pb.Payload events.
func (s *ObserverServer) GetEventsChannel() chan<- *pb.Payload {
	return s.events
}

// GetEndpointEventsChannel returns a channel that should be used to send
// AgentNotifyEndpoint* events when an endpoint is added, deleted or updated
// in Cilium.
func (s *ObserverServer) GetEndpointEventsChannel() chan<- monitorAPI.AgentNotify {
	return s.endpointEvents
}

func (s *ObserverServer) decodeFlow(pl *pb.Payload) (*pb.Flow, error) {
	// TODO: Pool these instead of allocating new flows each time.
	f := &pb.Flow{}
	err := s.payloadParser.Decode(pl, f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// ServerStatus should have a comment, apparently. It returns the server status.
func (s *ObserverServer) ServerStatus(
	ctx context.Context, req *observer.ServerStatusRequest,
) (*observer.ServerStatusResponse, error) {
	res := &observer.ServerStatusResponse{
		MaxFlows: s.ring.Cap(),
		NumFlows: s.ring.Len(),
	}
	return res, nil
}

func logFilters(filters []*pb.FlowFilter) string {
	var s []string
	for _, f := range filters {
		s = append(s, f.String())
	}
	return "{" + strings.Join(s, ",") + "}"
}

// GetFlows implements the proto method for client requests.
func (s *ObserverServer) GetFlows(
	req *observer.GetFlowsRequest,
	server observer.Observer_GetFlowsServer,
) (err error) {
	reply, err := getFlows(server.Context(), s.log, s.ring, req)
	if err != nil {
		return err
	}
	for {
		select {
		case <-server.Context().Done():
			return nil
		case rep, ok := <-reply:
			if !ok {
				return nil
			}
			err := server.Send(&observer.GetFlowsResponse{
				ResponseTypes: &observer.GetFlowsResponse_Flow{
					Flow: rep,
				},
			})
			if err != nil {
				return err
			}
		}
	}
}

func getUntil(req *observer.GetFlowsRequest, defaultTime *types.Timestamp) (time.Time, error) {
	until := req.GetUntil()
	if until == nil {
		until = defaultTime
	}
	return types.TimestampFromProto(until)
}

func getBufferCh(ctx context.Context, ring *container.Ring, req *observer.GetFlowsRequest) (ch <-chan *v1.Event, stop context.CancelFunc, err error) {
	stop = func() {}

	// s.ring.ReadFrom reads the values up to the last written index, i.e.,
	// it will read all values from the given interval:
	// [ lastWrite, s.ring.write [
	lastWrite := ring.LastWriteParallel() + 1
	readIdx := lastWrite - req.Number

	switch {
	case req.Follow:
		ch = ring.ReadFrom(ctx.Done(), readIdx)
	case req.Number != 0:
		var ctx1 context.Context
		ctx1, stop = context.WithCancel(ctx)
		ch = ring.ReadFrom(ctx1.Done(), readIdx)
	default:
		beginning, err := types.TimestampFromProto(req.GetSince())
		if err != nil {
			return nil, nil, err
		}
		end, err := getUntil(req, types.TimestampNow())
		if err != nil {
			return nil, nil, err
		}
		timestampCh := make(chan *v1.Event, 1000)
		ch = timestampCh

		var ctx1 context.Context
		ctx1, stop = context.WithCancel(ctx)

		go func() {
			defer close(timestampCh)
			for lastWrite := ring.LastWriteParallel(); ; lastWrite-- {
				e, ok := ring.Read(lastWrite)
				// if the buffer was not full yet we can get nil payloads
				if e == nil || e.Event == nil || !ok {
					return
				}
				ts, err := types.TimestampFromProto(e.GetFlow().GetTime())
				if err != nil {
					return
				}
				if beginning.Before(ts) && end.After(ts) {
					select {
					case <-ctx1.Done():
						return
					case timestampCh <- e:
					}
				}
			}
		}()
	}
	return ch, stop, nil
}

// getFlow returns the flows either depending on the requests performed.
func getFlows(
	ctx context.Context,
	log *zap.Logger,
	ring *container.Ring,
	req *observer.GetFlowsRequest,
) (chan *pb.Flow, error) {
	start := time.Now()
	i := uint64(0)
	defer func() {
		size := ring.Cap()
		took := time.Now().Sub(start)
		log.Debug(
			"GetFlows finished",
			zap.Uint64("number_of_flows", i),
			zap.Uint64("buffer_size", size),
			zap.String("whitelist", logFilters(req.Whitelist)),
			zap.String("blacklist", logFilters(req.Blacklist)),
			zap.Duration("took", took),
		)
	}()

	whitelist, err := filters.BuildFilterList(req.Whitelist)
	if err != nil {
		return nil, err
	}
	blacklist, err := filters.BuildFilterList(req.Blacklist)
	if err != nil {
		return nil, err
	}

	log.Debug("filters", zap.String("req", fmt.Sprintf("%+v", req)))
	log.Debug("whitelist", zap.String("whitelist", fmt.Sprintf("%+v", whitelist)))
	log.Debug("blacklist", zap.String("blacklist", fmt.Sprintf("%+v", blacklist)))

	ch, stop, err := getBufferCh(ctx, ring, req)
	if err != nil {
		return nil, err
	}
	reply := make(chan *pb.Flow, 1)
	go func() {
		defer close(reply)
		defer stop()

		for e := range ch {
			if req.Number != 0 && !req.Follow {
				i++
				if i >= req.Number {
					// stop the channel buffer because we have reached
					// the number of requested flows.
					stop()
					if i > req.Number {
						// We will 'continue' since 'ch' might have flows and we
						// need to empty that channel.
						return
					}
				}
			}
			if e == nil {
				continue
			}
			flow, ok := e.Event.(*pb.Flow)
			if !ok || !filters.Apply(whitelist, blacklist, e) {
				continue
			}
			select {
			case reply <- flow:
				// We have sent all expected flows so we can return already
				if req.Number != 0 && i >= req.Number {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return reply, nil
}
