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
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof" // a comment justifying it
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cilium/hubble/api/v1/observer"
	"github.com/cilium/hubble/pkg/api"
	v1 "github.com/cilium/hubble/pkg/api/v1"
	"github.com/cilium/hubble/pkg/cilium/client"
	"github.com/cilium/hubble/pkg/fqdncache"
	"github.com/cilium/hubble/pkg/ipcache"
	"github.com/cilium/hubble/pkg/logger"
	"github.com/cilium/hubble/pkg/metrics"
	metricsAPI "github.com/cilium/hubble/pkg/metrics/api"
	"github.com/cilium/hubble/pkg/parser"
	"github.com/cilium/hubble/pkg/server"
	"github.com/cilium/hubble/pkg/servicecache"
	"github.com/google/gops/agent"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// New ...
func New(log *zap.Logger) *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start gRPC server",
		Run: func(cmd *cobra.Command, args []string) {
			err := validateArgs(log)
			if err != nil {
				log.Fatal("failed to parse arguments", zap.Error(err))
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

			ciliumClient, err := client.NewClient()
			if err != nil {
				log.Fatal("failed to get Cilium client", zap.Error(err))
			}
			ipCache := ipcache.New()
			fqdnCache := fqdncache.New()
			serviceCache := servicecache.New()
			endpoints := v1.NewEndpoints()
			podGetter := &server.LegacyPodGetter{
				PodGetter:      ipCache,
				EndpointGetter: endpoints,
			}
			payloadParser, err := parser.New(endpoints, ciliumClient, fqdnCache, podGetter, serviceCache)
			if err != nil {
				log.Fatal("failed to get parser", zap.Error(err))
			}
			s := server.NewServer(
				ciliumClient,
				endpoints,
				ipCache,
				fqdnCache,
				serviceCache,
				payloadParser,
				int(maxFlows),
				logger.GetLogger(),
			)
			s.Start()
			err = Serve(log, listenClientUrls, s.GetGRPCServer())
			if err != nil {
				log.Fatal("", zap.Error(err))
			}
			fmt.Printf("Press Ctrl-C to quit\n")
			if err := s.HandleMonitorSocket(nodeName); err != nil {
				log.Fatal("HandleMonitorSocket failed", zap.Error(err))
			}
		},
	}

	serverCmd.Flags().StringArrayVarP(&listenClientUrls, "listen-client-urls", "", []string{serverSocketPath}, "List of URLs to listen on for client traffic.")
	serverCmd.Flags().Uint32Var(&maxFlows, "max-flows", 131071, "Max number of flows to store in memory (gets rounded up to closest (2^n)-1")
	serverCmd.Flags().StringVar(&serveDurationVar, "duration", "", "Shut the server down after this duration")
	serverCmd.Flags().StringVar(&nodeName, "node-name", os.Getenv(envNodeName), "Node name where hubble is running (defaults to value set in env variable '"+envNodeName+"'")

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
func Serve(log *zap.Logger, listenClientUrls []string, s server.GRPCServer) error {
	clientListeners, err := setupListeners(listenClientUrls)
	if err != nil {
		return err
	}

	serverStart = time.Now()

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
	return nil
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
