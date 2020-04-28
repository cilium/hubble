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

package peer

import (
	"context"
	"fmt"
	"io"
	"time"

	peerpb "github.com/cilium/cilium/api/v1/peer"
	"github.com/cilium/hubble/pkg/defaults"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	serverURL     string
	serverTimeout time.Duration
)

// New creates a new hidden peer command.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "peers",
		Aliases: []string{"peer"},
		Short:   "Get information about Hubble peers",
		Long:    `Get information about Hubble peers.`,
		Hidden:  true, // this command is only useful for development/debugging purposes
	}
	cmd.PersistentFlags().StringVar(&serverURL,
		"server", defaults.GetDefaultSocketPath(),
		"URL to connect to server")
	cmd.PersistentFlags().DurationVar(&serverTimeout,
		"timeout", 5*time.Second,
		"How long to wait before giving up on server dialing")
	cmd.AddCommand(
		newWatchCommand(),
	)
	return cmd
}

func newConn(target string, timeout time.Duration) (*grpc.ClientConn, error) {
	dialCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := grpc.DialContext(dialCtx, target, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to dial grpc: %v", err)
	}
	return conn, nil
}

func newWatchCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "watch",
		Aliases: []string{"w"},
		Short:   "Watch for Hubble peers updates",
		Long:    `Watch for Hubble peers updates.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			conn, err := newConn(serverURL, serverTimeout)
			if err != nil {
				return err
			}
			defer conn.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			return runWatch(ctx, peerpb.NewPeerClient(conn))
		},
	}
}

func runWatch(ctx context.Context, client peerpb.PeerClient) error {
	b, err := client.Notify(ctx, &peerpb.NotifyRequest{})
	if err != nil {
		return err
	}
	for {
		resp, err := b.Recv()
		switch err {
		case io.EOF, context.Canceled:
			return nil
		case nil:
			fmt.Printf("%-12s %s %s\n", resp.GetType(), resp.GetName(), resp.GetAddress())
		default:
			if status.Code(err) == codes.Canceled {
				return nil
			}
			return err
		}
	}
}
