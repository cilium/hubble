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
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cilium/cilium/api/v1/observer"
	v1 "github.com/cilium/cilium/pkg/hubble/api/v1"
	"github.com/cilium/hubble/cmd/common/config"
	"github.com/cilium/hubble/cmd/common/conn"
	"github.com/cilium/hubble/cmd/common/template"
	"github.com/cilium/hubble/pkg/defaults"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// New status command.
func New(vp *viper.Viper) *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Display status of Hubble server",
		Long: `Display shows the status of the Hubble server. This is intended as a basic
connectivity health check.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			hubbleConn, err := conn.New(ctx, vp.GetString(config.KeyServer), vp.GetDuration(config.KeyTimeout))
			if err != nil {
				return err
			}
			defer hubbleConn.Close()
			return runStatus(hubbleConn)
		},
	}

	// add config.ServerFlags to the help template as these flags are used by
	// this command
	statusCmd.SetUsageTemplate(template.Usage(config.ServerFlags))

	return statusCmd
}

func runStatus(conn *grpc.ClientConn) error {
	// get the standard GRPC health check to see if the server is up
	healthy, status, err := getHC(conn)
	if err != nil {
		return fmt.Errorf("failed getting status: %v", err)
	}
	fmt.Printf("Healthcheck (via %s): %s\n", conn.Target(), status)
	if !healthy {
		return errors.New("not healthy")
	}

	// if the server is up, lets try to get hubble specific status
	ss, err := getStatus(conn)
	if err != nil {
		return fmt.Errorf("failed to get hubble server status: %v", err)
	}
	flowsRatio := ""
	if ss.MaxFlows > 0 {
		flowsRatio = fmt.Sprintf(" (%.2f%%)", (float64(ss.NumFlows)/float64(ss.MaxFlows))*100)
	}
	fmt.Printf("Current/Max Flows: %v/%v%s\n", ss.NumFlows, ss.MaxFlows, flowsRatio)

	flowsPerSec := "N/A"
	if uptime := time.Duration(ss.UptimeNs).Seconds(); uptime > 0 {
		flowsPerSec = fmt.Sprintf("%.2f", float64(ss.SeenFlows)/uptime)
	}
	fmt.Printf("Flows/s: %s\n", flowsPerSec)

	numConnected := ss.GetNumConnectedNodes()
	numUnavailable := ss.GetNumUnavailableNodes()
	if numConnected != nil {
		total := ""
		if numUnavailable != nil {
			total = fmt.Sprintf("/%d", numUnavailable.Value+numConnected.Value)
		}
		fmt.Printf("Connected Nodes: %d%s\n", numConnected.Value, total)
	}
	if numUnavailable != nil && numUnavailable.Value > 0 {
		if unavailable := ss.GetUnavailableNodes(); unavailable != nil {
			sort.Strings(unavailable) // it's nicer when displaying unavailable nodes list
			if numUnavailable.Value > uint32(len(unavailable)) {
				unavailable = append(unavailable, fmt.Sprintf("and %d more...", numUnavailable.Value-uint32(len(unavailable))))
			}
			fmt.Printf("Unavailable Nodes: %d\n  - %s\n",
				numUnavailable.Value,
				strings.Join(unavailable, "\n  - "),
			)
		} else {
			fmt.Printf("Unavailable Nodes: %d\n", numUnavailable.Value)
		}
	}
	return nil
}

func getHC(conn *grpc.ClientConn) (healthy bool, status string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaults.RequestTimeout)
	defer cancel()

	req := &healthpb.HealthCheckRequest{Service: v1.ObserverServiceName}
	resp, err := healthpb.NewHealthClient(conn).Check(ctx, req)
	if err != nil {
		return false, "", err
	}
	if st := resp.GetStatus(); st != healthpb.HealthCheckResponse_SERVING {
		return false, fmt.Sprintf("Unavailable: %s", st), nil
	}
	return true, "Ok", nil
}

func getStatus(conn *grpc.ClientConn) (*observer.ServerStatusResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaults.RequestTimeout)
	defer cancel()

	req := &observer.ServerStatusRequest{}
	res, err := observer.NewObserverClient(conn).ServerStatus(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
