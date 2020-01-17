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
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/defaults"
	"github.com/cilium/cilium/pkg/math"
	"github.com/cilium/cilium/pkg/monitor"
	"github.com/cilium/cilium/pkg/monitor/agent/listener"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	"github.com/cilium/cilium/pkg/monitor/payload"
	pb "github.com/cilium/hubble/api/v1/flow"
	"github.com/cilium/hubble/api/v1/observer"
	"github.com/cilium/hubble/pkg/api"
	v1 "github.com/cilium/hubble/pkg/api/v1"
	"github.com/cilium/hubble/pkg/container"
	"github.com/cilium/hubble/pkg/filters"
	"github.com/cilium/hubble/pkg/ipcache"
	"github.com/cilium/hubble/pkg/logger"
	"github.com/cilium/hubble/pkg/metrics"
	"github.com/cilium/hubble/pkg/parser"
	parserErrors "github.com/cilium/hubble/pkg/parser/errors"
	"github.com/cilium/hubble/pkg/servicecache"
	"github.com/gogo/protobuf/types"
	"go.uber.org/zap"
)

// Observer defines the interface for observer server.
type Observer interface {
	observer.ObserverServer
	Start()
	GetEventsChannel() chan *pb.Payload
	GetRingBuffer() *container.Ring
	GetLogger() *zap.Logger
	GetStopped() chan struct{}
	GetPayloadParser() *parser.Parser
	HandleMonitorSocket(nodeName string) error
}

type ciliumClient interface {
	EndpointList() ([]*models.Endpoint, error)
	GetEndpoint(id uint64) (*models.Endpoint, error)
	GetIdentity(id uint64) (*models.Identity, error)
	GetFqdnCache() ([]*models.DNSLookup, error)
	GetIPCache() ([]*models.IPListEntry, error)
	GetServiceCache() ([]*models.Service, error)
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

	// serviceCache is a cache that contains information about services.
	serviceCache *servicecache.ServiceCache

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
	serviceCache *servicecache.ServiceCache,
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
		serviceCache:   serviceCache,
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

	processEvents(s)
}

func processEvents(s Observer) {
	for pl := range s.GetEventsChannel() {
		flow, err := decodeFlow(s.GetPayloadParser(), pl)
		if err != nil {
			if !parserErrors.IsErrInvalidType(err) {
				s.GetLogger().Debug("failed to decode payload", zap.ByteString("data", pl.Data), zap.Error(err))
			}
			continue
		}

		metrics.ProcessFlow(flow)
		s.GetRingBuffer().Write(&v1.Event{
			Timestamp: pl.Time,
			Event:     flow,
		})
	}
	close(s.GetStopped())
}

// startMirroringIPCache will obtain an initial IPCache snapshot from Cilium
// and then start mirroring IPCache events based on IPCacheNotification sent
// through the ipCacheEvents channels. Only messages of type
// `AgentNotifyIPCacheUpserted` and `AgentNotifyIPCacheDeleted` should be sent
// through that channel. This function assumes that the caller is already
// connected to Cilium Monitor, i.e. no IPCacheNotification must be lost after
// calling this method.
func (s *ObserverServer) startMirroringIPCache(ipCacheEvents <-chan monitorAPI.AgentNotify) {
	go s.syncIPCache(ipCacheEvents)
}

// startMirroringServiceCache initially caches service information from Cilium
// and then starts to mirror service information based on events that are sent
// to the serviceEvents channel. Only messages of type
// `AgentNotifyServiceUpserted` and `AgentNotifyServiceDeleted` should be sent
// to this channel.  This function assumes that the caller is already connected
// to Cilium Monitor, i.e. no Service notification must be lost after calling
// this method.
func (s *ObserverServer) startMirroringServiceCache(serviceEvents <-chan monitorAPI.AgentNotify) {
	go s.syncServiceCache(serviceEvents)
}

// getLogRecordNotifyChannel returns the event channel to receive
// monitorAPI.LogRecordNotify events.
func (s *ObserverServer) getLogRecordNotifyChannel() chan<- monitor.LogRecordNotify {
	return s.logRecord
}

// GetEventsChannel returns the event channel to receive pb.Payload events.
func (s *ObserverServer) GetEventsChannel() chan *pb.Payload {
	return s.events
}

// getEndpointEventsChannel returns a channel that should be used to send
// AgentNotifyEndpoint* events when an endpoint is added, deleted or updated
// in Cilium.
func (s *ObserverServer) getEndpointEventsChannel() chan<- monitorAPI.AgentNotify {
	return s.endpointEvents
}

// GetRingBuffer implements Observer.GetRingBuffer.
func (s *ObserverServer) GetRingBuffer() *container.Ring {
	return s.ring
}

// GetLogger implements Observer.GetLogger.
func (s *ObserverServer) GetLogger() *zap.Logger {
	return s.log
}

// GetStopped implements Observer.GetStopped.
func (s *ObserverServer) GetStopped() chan struct{} {
	return s.stopped
}

// GetPayloadParser implements Observer.GetPayloadParser.
func (s *ObserverServer) GetPayloadParser() *parser.Parser {
	return s.payloadParser
}

func decodeFlow(payloadParser *parser.Parser, pl *pb.Payload) (*pb.Flow, error) {
	// TODO: Pool these instead of allocating new flows each time.
	f := &pb.Flow{}
	err := payloadParser.Decode(pl, f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// ServerStatus should have a comment, apparently. It returns the server status.
func (s *ObserverServer) ServerStatus(
	ctx context.Context, req *observer.ServerStatusRequest,
) (*observer.ServerStatusResponse, error) {
	return getServerStatusFromObserver(s)
}

func getServerStatusFromObserver(obs Observer) (*observer.ServerStatusResponse, error) {
	res := &observer.ServerStatusResponse{
		MaxFlows: obs.GetRingBuffer().Cap(),
		NumFlows: obs.GetRingBuffer().Len(),
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
	return getFlowsFromObserver(req, server, s)
}

func getFlowsFromObserver(
	req *observer.GetFlowsRequest,
	server observer.Observer_GetFlowsServer,
	obs Observer,
) (err error) {
	reply, err := getFlows(server.Context(), obs.GetLogger(), obs.GetRingBuffer(), req)
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

// HandleMonitorSocket connects to the monitor socket and consumes monitor events.
func (s *ObserverServer) HandleMonitorSocket(nodeName string) error {
	// On EOF, retry
	// On other errors, exit
	// always wait connTimeout when retrying
	for ; ; time.Sleep(api.ConnectionTimeout) {
		conn, version, err := openMonitorSock()
		if err != nil {
			s.log.Error("Cannot open monitor serverSocketPath", zap.Error(err))
			return err
		}

		err = s.consumeMonitorEvents(conn, version, nodeName)
		switch {
		case err == nil:
			// no-op

		case err == io.EOF, err == io.ErrUnexpectedEOF:
			s.log.Warn("connection closed", zap.Error(err))
			continue

		default:
			log.Fatal("decoding error", zap.Error(err))
		}
	}
}

// getMonitorParser constructs and returns an eventParserFunc. It is
// appropriate for the monitor API version passed in.
func getMonitorParser(conn net.Conn, version listener.Version, nodeName string) (parser eventParserFunc, err error) {
	switch version {
	case listener.Version1_2:
		var (
			pl  payload.Payload
			dec = gob.NewDecoder(conn)
		)
		// This implements the newer 1.2 API. Each listener maintains its own gob
		// session, and type information is only ever sent once.
		return func() (*pb.Payload, error) {
			if err := pl.DecodeBinary(dec); err != nil {
				return nil, err
			}
			b := make([]byte, len(pl.Data))
			copy(b, pl.Data)

			// TODO: Eventually, the monitor will add these timestaps to events.
			// For now, we add them in hubble server.
			grpcPl := &pb.Payload{
				Data:     b,
				CPU:      int32(pl.CPU),
				Lost:     pl.Lost,
				Type:     pb.EventType(pl.Type),
				Time:     types.TimestampNow(),
				HostName: nodeName,
			}
			return grpcPl, nil
		}, nil

	default:
		return nil, fmt.Errorf("unsupported version %s", version)
	}
}

// consumeMonitorEvents handles and prints events on a monitor connection. It
// calls getMonitorParsed to construct a monitor-version appropriate parser.
// It closes conn on return, and returns on error, including io.EOF
func (s *ObserverServer) consumeMonitorEvents(conn net.Conn, version listener.Version, nodeName string) error {
	defer conn.Close()
	ch := s.GetEventsChannel()
	endpointEvents := s.getEndpointEventsChannel()

	dnsAdd := s.getLogRecordNotifyChannel()

	ipCacheEvents := make(chan monitorAPI.AgentNotify, 100)
	s.startMirroringIPCache(ipCacheEvents)

	serviceEvents := make(chan monitorAPI.AgentNotify, 100)
	s.startMirroringServiceCache(serviceEvents)

	getParsedPayload, err := getMonitorParser(conn, version, nodeName)
	if err != nil {
		return err
	}

	for {
		pl, err := getParsedPayload()
		if err != nil {
			return err
		}

		ch <- pl
		// we don't expect to have many MessageTypeAgent so we
		// can "decode" this messages as they come.
		switch pl.Data[0] {
		case monitorAPI.MessageTypeAgent:
			buf := bytes.NewBuffer(pl.Data[1:])
			dec := gob.NewDecoder(buf)

			an := monitorAPI.AgentNotify{}
			if err := dec.Decode(&an); err != nil {
				fmt.Printf("Error while decoding agent notification message: %s\n", err)
				continue
			}
			switch an.Type {
			case monitorAPI.AgentNotifyEndpointCreated,
				monitorAPI.AgentNotifyEndpointRegenerateSuccess,
				monitorAPI.AgentNotifyEndpointDeleted:
				endpointEvents <- an
			case monitorAPI.AgentNotifyIPCacheUpserted,
				monitorAPI.AgentNotifyIPCacheDeleted:
				ipCacheEvents <- an
			case monitorAPI.AgentNotifyServiceUpserted,
				monitorAPI.AgentNotifyServiceDeleted:
				serviceEvents <- an
			}
		case monitorAPI.MessageTypeAccessLog:
			// TODO re-think the way this is being done. We are dissecting/
			//      TypeAccessLog messages here *and* when we are dumping
			//      them into JSON.
			buf := bytes.NewBuffer(pl.Data[1:])
			dec := gob.NewDecoder(buf)

			lr := monitor.LogRecordNotify{}

			if err := dec.Decode(&lr); err != nil {
				fmt.Printf("Error while decoding access log message type: %s\n", err)
				continue
			}
			if lr.DNS != nil {
				dnsAdd <- lr
			}
		}
	}
}

// eventParseFunc is a convenience function type used as a version-specific
// parser of monitor events
type eventParserFunc func() (*pb.Payload, error)

// openMonitorSock attempts to open a version specific monitor serverSocketPath It
// returns a connection, with a version, or an error.
func openMonitorSock() (conn net.Conn, version listener.Version, err error) {
	errors := make([]string, 0)

	// try the 1.2 serverSocketPath
	conn, err = net.Dial("unix", defaults.MonitorSockPath1_2)
	if err == nil {
		return conn, listener.Version1_2, nil
	}
	errors = append(errors, defaults.MonitorSockPath1_2+": "+err.Error())

	return nil, listener.VersionUnsupported, fmt.Errorf("cannot find or open a supported node-monitor serverSocketPath. %s", strings.Join(errors, ","))
}
