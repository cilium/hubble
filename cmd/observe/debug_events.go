// Copyright 2021 Authors of Hubble
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

package observe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"

	observerpb "github.com/cilium/cilium/api/v1/observer"
	"github.com/cilium/hubble/cmd/common/config"
	"github.com/cilium/hubble/cmd/common/conn"
	"github.com/cilium/hubble/cmd/common/template"
	"github.com/cilium/hubble/pkg/defaults"
	"github.com/cilium/hubble/pkg/logger"
	hubtime "github.com/cilium/hubble/pkg/time"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newDebugEventsCommand(vp *viper.Viper, flagSets ...*pflag.FlagSet) *cobra.Command {
	debugEventsCmd := &cobra.Command{
		Use:   "debug-events",
		Short: "Observe Cilium debug events",
		RunE: func(cmd *cobra.Command, _ []string) error {
			debug := vp.GetBool(config.KeyDebug)
			if err := handleEventsArgs(debug); err != nil {
				return err
			}
			req, err := getDebugEventsRequest()
			if err != nil {
				return err
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
			defer cancel()

			hubbleConn, err := conn.New(ctx, vp.GetString(config.KeyServer), vp.GetDuration(config.KeyTimeout))
			if err != nil {
				return err
			}
			defer hubbleConn.Close()
			client := observerpb.NewObserverClient(hubbleConn)
			logger.Logger.WithField("request", req).Debug("Sending GetDebugEvents request")
			if err := getDebugEvents(ctx, client, req); err != nil {
				msg := err.Error()
				// extract custom error message from failed grpc call
				if s, ok := status.FromError(err); ok && s.Code() == codes.Unknown {
					msg = s.Message()
				}
				return errors.New(msg)
			}
			return nil
		},
	}

	debugEventsCmd.SetUsageTemplate(template.Usage(flagSets...))

	return debugEventsCmd
}

func getDebugEventsRequest() (*observerpb.GetDebugEventsRequest, error) {
	// convert selectorOpts.since into a param for GetDebugEvents
	var since, until *timestamppb.Timestamp
	if selectorOpts.since != "" {
		st, err := hubtime.FromString(selectorOpts.since)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the since time: %v", err)
		}

		since = timestamppb.New(st)
		if err := since.CheckValid(); err != nil {
			return nil, fmt.Errorf("failed to convert `since` timestamp to proto: %v", err)
		}
		// Set the until field if both --since and --until options are specified and --follow
		// is not specified. If --since is specified but --until is not, the server sets the
		// --until option to the current timestamp.
		if selectorOpts.until != "" && !selectorOpts.follow {
			ut, err := hubtime.FromString(selectorOpts.until)
			if err != nil {
				return nil, fmt.Errorf("failed to parse the until time: %v", err)
			}
			until = timestamppb.New(ut)
			if err := until.CheckValid(); err != nil {
				return nil, fmt.Errorf("failed to convert `until` timestamp to proto: %v", err)
			}
		}
	}

	if since == nil && until == nil {
		switch {
		case selectorOpts.all:
			// all is an alias for last=uint64_max
			selectorOpts.last = ^uint64(0)
		case selectorOpts.last == 0:
			// no specific parameters were provided, just a vanilla `hubble events debug`
			selectorOpts.last = defaults.EventsPrintCount
		}
	}

	return &observerpb.GetDebugEventsRequest{
		Number: selectorOpts.last,
		Follow: selectorOpts.follow,
		Since:  since,
		Until:  until,
	}, nil
}

func getDebugEvents(ctx context.Context, client observerpb.ObserverClient, req *observerpb.GetDebugEventsRequest) error {
	b, err := client.GetDebugEvents(ctx, req)
	if err != nil {
		return err
	}

	defer printer.Close()

	for {
		resp, err := b.Recv()
		switch err {
		case io.EOF, context.Canceled:
			return nil
		case nil:
		default:
			if status.Code(err) == codes.Canceled {
				return nil
			}
			return err
		}

		if err = printer.WriteProtoDebugEvent(resp); err != nil {
			return err
		}
	}
}
