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

package watch

import (
	"context"
	"fmt"
	"io"
	"os"

	peerpb "github.com/cilium/cilium/api/v1/peer"
	"github.com/cilium/hubble/cmd/common/config"
	"github.com/cilium/hubble/cmd/common/conn"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newPeerCommand(vp *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:     "peers",
		Aliases: []string{"peer"},
		Short:   "Watch for Hubble peers updates",
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			hubbleConn, err := conn.New(ctx, vp.GetString(config.KeyServer), vp.GetDuration(config.KeyTimeout))
			if err != nil {
				return err
			}
			defer hubbleConn.Close()
			return runPeer(ctx, peerpb.NewPeerClient(hubbleConn))
		},
	}
}

func runPeer(ctx context.Context, client peerpb.PeerClient) error {
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
			processResponse(os.Stdout, resp)
		default:
			if status.Code(err) == codes.Canceled {
				return nil
			}
			return err
		}
	}
}

func processResponse(w io.Writer, resp *peerpb.ChangeNotification) {
	tlsServerName := ""
	if tls := resp.GetTls(); tls != nil {
		tlsServerName = fmt.Sprintf(" (TLS.ServerName: %s)", tls.GetServerName())
	}
	_, _ = fmt.Fprintf(w, "%-12s %s %s%s\n", resp.GetType(), resp.GetAddress(), resp.GetName(), tlsServerName)
}
