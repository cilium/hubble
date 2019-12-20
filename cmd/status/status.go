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

package status

import (
	"context"
	"fmt"
	"os"

	"github.com/cilium/hubble/api/v1/observer"
	"github.com/cilium/hubble/pkg/api"
	v1 "github.com/cilium/hubble/pkg/api/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	serverURL string
)

// New status command.
func New() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Display status of hubble server",
		Long: `Displays the status of the hubble server. This is
		intended as a basic connectivity health check`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}

	statusCmd.Flags().StringVarP(&serverURL,
		"server", "", api.DefaultSocketPath, "URL to connect to server")
	viper.BindEnv("server", "HUBBLE_SOCK")

	return statusCmd
}

func runStatus() error {
	// get the standard GRPC health check to see if the server is up
	healthy, status, err := getHC(serverURL)
	if err != nil {
		fmt.Println("Failed getting status:", err)
		os.Exit(-1) // TODO: rewrite without os.Exit, return an error
	}
	fmt.Printf("Healthcheck (via %s): %s\n", serverURL, status)
	if !healthy {
		os.Exit(-1) // TODO: rewrite without os.Exit, return an error
	}

	// if the server is up, lets try to get hubble specific status
	ss, err := getStatus(serverURL)
	if err != nil {
		fmt.Println("Failed to get hubble server status:", err)
	}
	fmt.Println("Max Flows:", ss.MaxFlows)
	fmt.Printf(
		"Current Flows: %v (%.2f%%) \n",
		ss.NumFlows,
		(float64(ss.NumFlows)/float64(ss.MaxFlows))*100,
	)

	return nil
}

func getHC(s string) (bool, string, error) {
	healthy := false
	status := ""
	conn, err := grpc.Dial(s, grpc.WithInsecure())
	if err != nil {
		return healthy, status, err
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), api.ConnectionTimeout)
	defer cancel()

	req := &healthpb.HealthCheckRequest{Service: v1.ObserverServiceName}
	resp, err := healthpb.NewHealthClient(conn).Check(ctx, req)
	if err != nil {
		status = fmt.Sprintf("Error: %s", err)
	} else if st := resp.GetStatus(); st != healthpb.HealthCheckResponse_SERVING {
		status = fmt.Sprintf("Unavailable: %s", st)
	} else {
		status = "Ok"
		healthy = true
	}

	return healthy, status, err
}

func getStatus(s string) (*observer.ServerStatusResponse, error) {
	conn, err := grpc.Dial(s, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), api.ConnectionTimeout)
	defer cancel()

	req := &observer.ServerStatusRequest{}
	res, err := observer.NewObserverClient(conn).ServerStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
