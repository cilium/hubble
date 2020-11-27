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

package node

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	observerpb "github.com/cilium/cilium/api/v1/observer"
	relaypb "github.com/cilium/cilium/api/v1/relay"
	"github.com/cilium/hubble/cmd/common/conn"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const notAvailable = "N/A"

var listOpts struct {
	output string
}

func newListCommand(vp *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Hubble nodes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			hubbleConn, err := conn.New(ctx, vp.GetString("server"), vp.GetDuration("timeout"))
			if err != nil {
				return err
			}
			defer hubbleConn.Close()
			return runList(ctx, cmd, hubbleConn)
		},
	}

	cmd.Flags().StringVarP(
		&listOpts.output, "output", "o", "table",
		`Specify the output format, one of:
 json:     JSON encoding
 table:    Tab-aligned columns
 wide:     Tab-aligned columns with additional information`)
	cmd.RegisterFlagCompletionFunc("output", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"json",
			"table",
			"wide",
		}, cobra.ShellCompDirectiveDefault
	})
	return cmd
}

func runList(ctx context.Context, cmd *cobra.Command, conn *grpc.ClientConn) error {
	req := &observerpb.GetNodesRequest{}
	res, err := observerpb.NewObserverClient(conn).GetNodes(ctx, req)
	if err != nil {
		return err
	}

	nodes := res.GetNodes()
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
	switch listOpts.output {
	case "json":
		bs, err := json.MarshalIndent(nodes, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(bs))
		return err
	case "table", "wide":
		tw := tabwriter.NewWriter(cmd.OutOrStdout(), 2, 0, 3, ' ', 0)
		fmt.Fprint(tw, "NAME\tSTATUS\tAGE\tFLOWS/S\tCURRENT/MAX-FLOWS")
		if listOpts.output == "wide" {
			fmt.Fprint(tw, "\tVERSION\tADDRESS\tTLS")
		}
		fmt.Fprintln(tw)

		for _, n := range nodes {
			age := notAvailable
			flowsPerSec := notAvailable
			if uptime := time.Duration(n.GetUptimeNs()).Round(time.Second); uptime > 0 {
				age = uptime.String()
				flowsPerSec = fmt.Sprintf("%.2f", float64(n.GetSeenFlows())/uptime.Seconds())
			}
			flowsRatio := notAvailable
			if maxFlows := n.GetMaxFlows(); maxFlows > 0 {
				flowsRatio = fmt.Sprintf("%d/%d (%6.2f%%)", n.GetNumFlows(), maxFlows, (float64(n.GetNumFlows())/float64(maxFlows))*100)
			}
			version := notAvailable
			if v := n.GetVersion(); v != "" {
				version = v
			}
			fmt.Fprint(tw, n.GetName(), "\t", strings.Title(nodeStateToString(n.GetState())), "\t", age, "\t", flowsPerSec, "\t", flowsRatio)
			if listOpts.output == "wide" {
				tls := notAvailable
				if t := n.GetTls(); t != nil {
					tls = "Disabled"
					if t.Enabled {
						tls = "Enabled"
					}
				}
				fmt.Fprint(tw, "\t", version, "\t", n.GetAddress(), "\t", tls)
			}
			fmt.Fprintln(tw)
		}
		return tw.Flush()
	default:
		return fmt.Errorf("unknown output format: %s", listOpts.output)
	}
}

func nodeStateToString(state relaypb.NodeState) string {
	switch state {
	case relaypb.NodeState_NODE_CONNECTED:
		return "connected"
	case relaypb.NodeState_NODE_UNAVAILABLE:
		return "unavailable"
	case relaypb.NodeState_NODE_GONE:
		return "gone"
	case relaypb.NodeState_NODE_ERROR:
		return "error"
	case relaypb.NodeState_UNKNOWN_NODE_STATE:
		fallthrough
	default:
		return "unknown"
	}
}
