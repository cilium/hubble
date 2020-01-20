// Copyright 2017-2019 Authors of Hubble
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

package serve

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof" // a comment justifying it
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cilium/cilium/pkg/defaults"
	"github.com/cilium/cilium/pkg/monitor"
	"github.com/cilium/cilium/pkg/monitor/agent/listener"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	"github.com/cilium/cilium/pkg/monitor/payload"
	pb "github.com/cilium/hubble/api/v1/flow"
	"github.com/cilium/hubble/api/v1/observer"
	"github.com/cilium/hubble/pkg/api"
	v1 "github.com/cilium/hubble/pkg/api/v1"
	"github.com/cilium/hubble/pkg/cilium/client"
	"github.com/cilium/hubble/pkg/format"
	"github.com/cilium/hubble/pkg/fqdncache"
	"github.com/cilium/hubble/pkg/ipcache"
	"github.com/cilium/hubble/pkg/metrics"
	metricsAPI "github.com/cilium/hubble/pkg/metrics/api"
	"github.com/cilium/hubble/pkg/parser"
	"github.com/cilium/hubble/pkg/server"
	"github.com/gogo/protobuf/types"
	"github.com/google/gops/agent"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// New ...
func New(log *zap.Logger) *cobra.Command {
	var numeric bool

	serverCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC server",
		Run: func(cmd *cobra.Command, args []string) {
			err := validateArgs(log)
			if err != nil {
				log.Fatal("failed to parse arguments", zap.Error(err))
			}

			if numeric {
				format.EnableIPTranslation = false
				format.EnablePortTranslation = false
			}

			if gopsVar {
				log.Debug("starting gops agent")
				if err := agent.Listen(agent.Options{}); err != nil {
					log.Fatal("failed to start gops agent", zap.Error(err))
				}
			}

			if pprofVar {
				log.Debug("starting http/pprof handler")
				// Even though gops agent might also be running running, http
				// pprof has no overhead unless called upon and can be very
				// useful.
				go func() {
					// ignore http/pprof error
					_ = http.ListenAndServe(":6060", nil)
				}()
			}

			err = Serve(log, listenClientUrls)
			if err != nil {
				log.Fatal("", zap.Error(err))
			}
		},
	}

	serverCmd.Flags().StringArrayVarP(&listenClientUrls, "listen-client-urls", "", []string{serverSocketPath}, "List of URLs to listen on for client traffic.")
	serverCmd.Flags().Uint32Var(&maxFlows, "max-flows", 131071, "Max number of flows to store in memory (gets rounded up to closest (2^n)-1")
	serverCmd.Flags().StringVar(&serveDurationVar, "duration", "", "Shut the server down after this duration")
	serverCmd.Flags().StringVar(&nodeName, "node-name", os.Getenv(envNodeName), "Node name where hubble is running (defaults to value set in env variable '"+envNodeName+"'")

	serverCmd.Flags().BoolVarP(&numeric, "numeric", "n", false, "Display all information in numeric form")
	serverCmd.Flags().BoolVar(&format.EnablePortTranslation, "port-translation", true, "Translate port numbers to names")
	serverCmd.Flags().BoolVar(&format.EnableIPTranslation, "ip-translation", true, "Translate IP addresses to logical names such as pod name, FQDN, ...")
	serverCmd.Flags().StringSliceVar(&enabledMetrics, "metric", []string{}, "Enable metrics reporting")
	serverCmd.Flags().StringVar(&metricsServer, "metrics-server", "", "Address to serve metrics on")

	serverCmd.Flags().BoolVar(&gopsVar, "gops", false, "Run gops agent")
	serverCmd.Flags().BoolVar(&pprofVar, "pprof", false, "Run http/pprof handler")
	serverCmd.Flags().Lookup("gops").Hidden = true
	serverCmd.Flags().Lookup("pprof").Hidden = true

	return serverCmd
}

// observerCmd represents the monitor command
var (
	maxFlows uint32

	serveDurationVar string
	serveDuration    time.Duration
	nodeName         string

	listenClientUrls []string

	// when the server started
	serverStart time.Time

	enabledMetrics []string
	metricsServer  string

	gopsVar, pprofVar bool
)

const (
	serverSocketPath = "unix:///var/run/hubble.sock"
	envNodeName      = "HUBBLE_NODE_NAME"
)

func enableMetrics(log *zap.Logger, m []string) {
	errChan, err := metrics.Init(metricsServer, metricsAPI.ParseMetricList(m))
	if err != nil {
		log.Fatal("Unable to setup metrics", zap.Error(err))
	}

	go func() {
		err := <-errChan
		if err != nil {
			log.Fatal("Unable to initialize metrics server", zap.Error(err))
		}
	}()

}

func validateArgs(log *zap.Logger) error {
	if serveDurationVar != "" {
		d, err := time.ParseDuration(serveDurationVar)
		if err != nil {
			log.Fatal(
				"failed to parse the provided --duration",
				zap.String("duration", serveDurationVar),
			)
		}
		serveDuration = d
	}

	log.Info(
		"Started server with args",
		zap.Uint32("max-flows", maxFlows),
		zap.Duration("duration", serveDuration),
	)

	if metricsServer != "" {
		enableMetrics(log, enabledMetrics)
	}

	return nil
}

func setupListeners(listenClientUrls []string) (listeners map[string]net.Listener, err error) {
	listeners = map[string]net.Listener{}
	defer func() {
		if err != nil {
			for _, list := range listeners {
				list.Close()
			}
		}
	}()

	for _, listenClientURL := range listenClientUrls {
		if listenClientURL == "" {
			continue
		}
		if !strings.HasPrefix(listenClientURL, "unix://") {
			var socket net.Listener
			socket, err = net.Listen("tcp", listenClientURL)
			if err != nil {
				return nil, err
			}
			listeners[listenClientURL] = socket
		} else {
			socketPath := strings.TrimPrefix(listenClientURL, "unix://")
			syscall.Unlink(socketPath)
			var socket net.Listener
			socket, err = net.Listen("unix", socketPath)
			if err != nil {
				return
			}

			if os.Getuid() == 0 {
				err = api.SetDefaultPermissions(socketPath)
				if err != nil {
					return nil, err
				}
			}
			listeners[listenClientURL] = socket
		}
	}
	return listeners, nil
}

// Serve starts the GRPC server on the provided socketPath. If the port is non-zero, it listens
// to the TCP port instead of the unix domain socket.
func Serve(log *zap.Logger, listenClientUrls []string) error {
	clientListeners, err := setupListeners(listenClientUrls)
	if err != nil {
		return err
	}

	ciliumClient, err := client.NewClient()
	if err != nil {
		return err
	}

	ipCache := ipcache.New()
	fqdnCache := fqdncache.New()
	endpoints := v1.NewEndpoints()
	podGetter := &server.LegacyPodGetter{
		PodGetter:      ipCache,
		EndpointGetter: endpoints,
	}

	payloadParser, err := parser.New(endpoints, ciliumClient, fqdnCache, podGetter)
	if err != nil {
		return err
	}

	s := server.NewServer(
		ciliumClient,
		endpoints,
		ipCache,
		fqdnCache,
		payloadParser,
		int(maxFlows),
	)

	serverStart = time.Now()
	go s.Start()

	if serveDuration != 0 {
		// Register a server shutdown
		go func() {
			<-time.After(serveDuration)
			log.Info(
				"Shutting down after the configured duration",
				zap.Duration("duration", serveDuration),
			)
			os.Exit(0)
		}()
	}

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus(v1.ObserverServiceName, healthpb.HealthCheckResponse_SERVING)

	clientGRPC := grpc.NewServer()

	observer.RegisterObserverServer(clientGRPC, s)
	healthpb.RegisterHealthServer(clientGRPC, healthSrv)

	for clientListURL, clientList := range clientListeners {
		go func(clientListURL string, clientList net.Listener) {
			log.Info("Starting gRPC server on client-listener", zap.String("client-listener", clientListURL))
			err = clientGRPC.Serve(clientList)
			if err != nil {
				log.Fatal("failed to close grpc server", zap.Error(err))
			}
		}(clientListURL, clientList)
	}

	setupSigHandler()
	fmt.Printf("Press Ctrl-C to quit\n")

	// On EOF, retry
	// On other errors, exit
	// always wait connTimeout when retrying
	for ; ; time.Sleep(api.ConnectionTimeout) {
		conn, version, err := openMonitorSock()
		if err != nil {
			log.Error("Cannot open monitor serverSocketPath", zap.Error(err))
			return err
		}

		err = consumeMonitorEvents(s, conn, version)
		switch {
		case err == nil:
		// no-op

		case err == io.EOF, err == io.ErrUnexpectedEOF:
			log.Warn("connection closed", zap.Error(err))
			continue

		default:
			log.Fatal("decoding error", zap.Error(err))
		}
	}
}

func setupSigHandler() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			fmt.Printf("\nReceived an interrupt, disconnecting from monitor...\n\n")
			os.Exit(0)
		}
	}()
}

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

// eventParseFunc is a convenience function type used as a version-specific
// parser of monitor events
type eventParserFunc func() (*pb.Payload, error)

// getMonitorParser constructs and returns an eventParserFunc. It is
// appropriate for the monitor API version passed in.
func getMonitorParser(conn net.Conn, version listener.Version) (parser eventParserFunc, err error) {
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
func consumeMonitorEvents(s *server.ObserverServer, conn net.Conn, version listener.Version) error {
	defer conn.Close()
	ch := s.GetEventsChannel()
	endpointEvents := s.GetEndpointEventsChannel()

	dnsAdd := s.GetLogRecordNotifyChannel()

	ipCacheEvents := make(chan monitorAPI.AgentNotify, 100)
	s.StartMirroringIPCache(ipCacheEvents)

	getParsedPayload, err := getMonitorParser(conn, version)
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
			case monitorAPI.AgentNotifyIPCacheUpserted:
				ipCacheEvents <- an
			case monitorAPI.AgentNotifyIPCacheDeleted:
				ipCacheEvents <- an
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
