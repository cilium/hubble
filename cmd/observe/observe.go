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

package observe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"

	pb "github.com/cilium/cilium/api/v1/flow"
	"github.com/cilium/cilium/api/v1/observer"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	"github.com/cilium/hubble/cmd/common/config"
	"github.com/cilium/hubble/cmd/common/conn"
	"github.com/cilium/hubble/cmd/common/template"
	"github.com/cilium/hubble/pkg/defaults"
	hubprinter "github.com/cilium/hubble/pkg/printer"
	hubtime "github.com/cilium/hubble/pkg/time"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	selectorOpts struct {
		all          bool
		last         uint64
		since, until string
		follow       bool
	}

	formattingOpts struct {
		jsonOutput    bool
		compactOutput bool
		dictOutput    bool
		output        string

		enableIPTranslation bool
		nodeName            bool
		numeric             bool
	}

	otherOpts struct {
		ignoreStderr bool
	}

	printer *hubprinter.Printer
)

var verdicts = []string{
	pb.Verdict_FORWARDED.String(),
	pb.Verdict_DROPPED.String(),
	pb.Verdict_ERROR.String(),
}

func eventTypes() (l []string) {
	for t := range monitorAPI.MessageTypeNames {
		l = append(l, t)
	}
	return
}

// New observer command.
func New(vp *viper.Viper) *cobra.Command {
	return newObserveCmd(vp, newObserveFilter())
}

func newObserveCmd(vp *viper.Viper, ofilter *observeFilter) *cobra.Command {
	observeCmd := &cobra.Command{
		Use:   "observe",
		Short: "Observe flows of a Hubble server",
		Long: `Observe provides visibility into flow information on the network and
application level. Rich filtering enable observing specific flows related to
individual pods, services, TCP connections, DNS queries, HTTP requests and
more.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			debug := vp.GetBool(config.KeyDebug)
			if err := handleArgs(ofilter, debug); err != nil {
				return err
			}
			if debug {
				fmt.Fprintf(cmd.ErrOrStderr(), "Using filters:\n=> include: %s\n=> exclude: %s\n", ofilter.whitelist, ofilter.blacklist)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			hubbleConn, err := conn.New(ctx, vp.GetString(config.KeyServer), vp.GetDuration(config.KeyTimeout))
			if err != nil {
				return err
			}
			defer hubbleConn.Close()

			if err := runObserve(hubbleConn, ofilter); err != nil {
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

	// selector flags
	selectorFlags := pflag.NewFlagSet("selectors", pflag.ContinueOnError)
	selectorFlags.BoolVar(&selectorOpts.all, "all", false, "Get all flows stored in Hubble's buffer")
	selectorFlags.Uint64Var(&selectorOpts.last, "last", 0, fmt.Sprintf("Get last N flows stored in Hubble's buffer (default %d)", defaults.FlowPrintCount))
	selectorFlags.StringVar(&selectorOpts.since, "since", "", "Filter flows since a specific date (relative or RFC3339)")
	selectorFlags.StringVar(&selectorOpts.until, "until", "", "Filter flows until a specific date (relative or RFC3339)")
	selectorFlags.BoolVarP(&selectorOpts.follow, "follow", "f", false, "Follow flows output")
	observeCmd.Flags().AddFlagSet(selectorFlags)

	// filter flags
	filterFlags := pflag.NewFlagSet("filters", pflag.ContinueOnError)
	filterFlags.Var(filterVar(
		"not", ofilter,
		"Reverses the next filter to be blacklist i.e. --not --from-ip 2.2.2.2"))
	filterFlags.Var(filterVar(
		"node-name", ofilter,
		`Show all flows which match the given node names (e.g. "k8s*", "test-cluster/*.company.com")`))
	filterFlags.Var(filterVar(
		"protocol", ofilter,
		`Show only flows which match the given L4/L7 flow protocol (e.g. "udp", "http")`))
	filterFlags.Var(filterVar(
		"tcp-flags", ofilter,
		`Show only flows which match the given TCP flags (e.g. "syn", "ack", "fin")`))
	filterFlags.VarP(filterVarP(
		"type", "t", ofilter, []string{},
		fmt.Sprintf("Filter by event types TYPE[:SUBTYPE] (%v)", eventTypes())))
	filterFlags.Var(filterVar(
		"verdict", ofilter,
		fmt.Sprintf("Show only flows with this verdict [%s]", strings.Join(verdicts, ", ")),
	))

	filterFlags.Var(filterVar(
		"http-status", ofilter,
		`Show only flows which match this HTTP status code prefix (e.g. "404", "5+")`))
	filterFlags.Var(filterVar(
		"http-method", ofilter,
		`Show only flows which match this HTTP method (e.g. "get", "post")`))
	filterFlags.Var(filterVar(
		"http-path", ofilter,
		`Show only flows which match this HTTP path regular expressions (e.g. "/page/\\d+")`))

	filterFlags.Var(filterVar(
		"from-fqdn", ofilter,
		`Show all flows originating at the given fully qualified domain name (e.g. "*.cilium.io").`))
	filterFlags.Var(filterVar(
		"fqdn", ofilter,
		`Show all flows related to the given fully qualified domain name (e.g. "*.cilium.io").`))
	filterFlags.Var(filterVar(
		"to-fqdn", ofilter,
		`Show all flows terminating at the given fully qualified domain name (e.g. "*.cilium.io").`))

	filterFlags.Var(filterVar(
		"from-ip", ofilter,
		"Show all flows originating at the given IP address."))
	filterFlags.Var(filterVar(
		"ip", ofilter,
		"Show all flows related to the given IP address."))
	filterFlags.Var(filterVar(
		"to-ip", ofilter,
		"Show all flows terminating at the given IP address."))

	filterFlags.VarP(filterVarP(
		"ipv4", "4", ofilter, nil,
		`Show only IPv4 flows`))
	filterFlags.Lookup("ipv4").NoOptDefVal = "v4" // add default val so none is required to be provided
	filterFlags.VarP(filterVarP(
		"ipv6", "6", ofilter, nil,
		`Show only IPv6 flows`))
	filterFlags.Lookup("ipv6").NoOptDefVal = "v6" // add default val so none is required to be provided
	filterFlags.Var(filterVar(
		"ip-version", ofilter,
		`Show only IPv4, IPv6 flows or non IP flows (e.g. ARP packets) (ie: "none", "v4", "v6")`))

	filterFlags.Var(filterVar(
		"from-pod", ofilter,
		"Show all flows originating in the given pod name ([namespace/]<pod-name>). If namespace is not provided, 'default' is used"))
	filterFlags.Var(filterVar(
		"pod", ofilter,
		"Show all flows related to the given pod name ([namespace/]<pod-name>). If namespace is not provided, 'default' is used"))
	filterFlags.Var(filterVar(
		"to-pod", ofilter,
		"Show all flows terminating in the given pod name ([namespace/]<pod-name>). If namespace is not provided, 'default' is used"))

	filterFlags.Var(filterVar(
		"from-namespace", ofilter,
		"Show all flows originating in the given Kubernetes namespace."))
	filterFlags.VarP(filterVarP(
		"namespace", "n", ofilter, nil,
		"Show all flows related to the given Kubernetes namespace."))
	filterFlags.Var(filterVar(
		"to-namespace", ofilter,
		"Show all flows terminating in the given Kubernetes namespace."))

	filterFlags.Var(filterVar(
		"from-label", ofilter,
		`Show only flows originating in an endpoint with the given labels (e.g. "key1=value1", "reserved:world")`))
	filterFlags.VarP(filterVarP(
		"label", "l", ofilter, nil,
		`Show only flows related to an endpoint with the given labels (e.g. "key1=value1", "reserved:world")`))
	filterFlags.Var(filterVar(
		"to-label", ofilter,
		`Show only flows terminating in an endpoint with given labels (e.g. "key1=value1", "reserved:world")`))

	filterFlags.Var(filterVar(
		"from-service", ofilter,
		"Show all flows originating in the given service ([namespace/]<svc-name>). If namespace is not provided, 'default' is used"))
	filterFlags.Var(filterVar(
		"service", ofilter,
		"Show all flows related to the given service ([namespace/]<svc-name>). If namespace is not provided, 'default' is used"))
	filterFlags.Var(filterVar(
		"to-service", ofilter,
		"Show all flows terminating in the given service ([namespace/]<svc-name>). If namespace is not provided, 'default' is used"))

	filterFlags.Var(filterVar(
		"from-port", ofilter,
		"Show only flows with the given source port (e.g. 8080)"))
	filterFlags.Var(filterVar(
		"port", ofilter,
		"Show only flows with given port in either source or destination (e.g. 8080)"))
	filterFlags.Var(filterVar(
		"to-port", ofilter,
		"Show only flows with the given destination port (e.g. 8080)"))

	filterFlags.Var(filterVar(
		"from-identity", ofilter,
		"Show all flows originating at an endpoint with the given security identity"))
	filterFlags.Var(filterVar(
		"identity", ofilter,
		"Show all flows related to an endpoint with the given security identity"))
	filterFlags.Var(filterVar(
		"to-identity", ofilter,
		"Show all flows terminating at an endpoint with the given security identity"))
	observeCmd.Flags().AddFlagSet(filterFlags)

	formattingFlags := pflag.NewFlagSet("Formatting", pflag.ContinueOnError)
	formattingFlags.BoolVarP(
		&formattingOpts.jsonOutput, "json", "j", false, "Deprecated. Use '--output json' instead.",
	)
	formattingFlags.BoolVar(
		&formattingOpts.compactOutput, "compact", false, "Deprecated. Use '--output compact' instead.",
	)
	formattingFlags.BoolVar(
		&formattingOpts.dictOutput, "dict", false, "Deprecated. Use '--output dict' instead.",
	)
	formattingFlags.StringVarP(
		&formattingOpts.output, "output", "o", "",
		`Specify the output format, one of:
 compact:  Compact output
 dict:     Each flow is shown as KEY:VALUE pair
 json:     JSON encoding
 jsonpb:   Output each GetFlowResponse according to proto3's JSON mapping
 table:    Tab-aligned columns`)
	formattingFlags.BoolVar(
		&formattingOpts.numeric,
		"numeric",
		false,
		"Display all information in numeric form",
	)
	formattingFlags.BoolVar(
		&formattingOpts.enableIPTranslation,
		"ip-translation",
		true,
		"Translate IP addresses to logical names such as pod name, FQDN, ...",
	)
	formattingFlags.BoolVarP(&formattingOpts.nodeName, "print-node-name", "", false, "Print node name in output")
	observeCmd.Flags().AddFlagSet(formattingFlags)

	// other flags
	otherFlags := pflag.NewFlagSet("other", pflag.ContinueOnError)
	otherFlags.BoolVarP(
		&otherOpts.ignoreStderr, "silent-errors", "s", false, "Silently ignores errors and warnings")
	observeCmd.Flags().AddFlagSet(otherFlags)

	// advanced completion for flags
	observeCmd.RegisterFlagCompletionFunc("ip-version", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"none", "v4", "v6"}, cobra.ShellCompDirectiveDefault
	})
	observeCmd.RegisterFlagCompletionFunc("type", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return eventTypes(), cobra.ShellCompDirectiveDefault
	})
	observeCmd.RegisterFlagCompletionFunc("verdict", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return verdicts, cobra.ShellCompDirectiveDefault
	})
	observeCmd.RegisterFlagCompletionFunc("http-status", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		httpStatus := []string{
			"100", "101", "102", "103",
			"200", "201", "202", "203", "204", "205", "206", "207", "208",
			"226",
			"300", "301", "302", "303", "304", "305", "307", "308",
			"400", "401", "402", "403", "404", "405", "406", "407", "408", "409",
			"410", "411", "412", "413", "414", "415", "416", "417", "418",
			"421", "422", "423", "424", "425", "426", "428", "429",
			"431",
			"451",
			"500", "501", "502", "503", "504", "505", "506", "507", "508",
			"510", "511",
		}
		return httpStatus, cobra.ShellCompDirectiveDefault
	})
	observeCmd.RegisterFlagCompletionFunc("http-method", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{
			http.MethodConnect,
			http.MethodDelete,
			http.MethodGet,
			http.MethodHead,
			http.MethodOptions,
			http.MethodPatch,
			http.MethodPost,
			http.MethodPut,
			http.MethodTrace,
		}, cobra.ShellCompDirectiveDefault
	})
	observeCmd.RegisterFlagCompletionFunc("output", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"compact",
			"dict",
			"json",
			"jsonpb",
			"table",
		}, cobra.ShellCompDirectiveDefault
	})

	// default value for when the flag is on the command line without any options
	observeCmd.Flags().Lookup("not").NoOptDefVal = "true"

	observeCmd.SetUsageTemplate(template.Usage(selectorFlags, filterFlags, formattingFlags, config.ServerFlags, otherFlags))

	return observeCmd
}

func handleArgs(ofilter *observeFilter, debug bool) (err error) {
	if ofilter.blacklisting {
		return errors.New("trailing --not found in the arguments")
	}

	// initialize the printer with any options that were passed in
	var opts []hubprinter.Option

	if formattingOpts.output == "" { // support deprecated output flags if provided
		if formattingOpts.jsonOutput {
			formattingOpts.output = "json"
		} else if formattingOpts.dictOutput {
			formattingOpts.output = "dict"
		} else if formattingOpts.compactOutput {
			formattingOpts.output = "compact"
		}
	}

	switch formattingOpts.output {
	case "compact":
		opts = append(opts, hubprinter.Compact())
	case "dict":
		opts = append(opts, hubprinter.Dict())
	case "json", "JSON":
		opts = append(opts, hubprinter.JSON())
	case "jsonpb":
		opts = append(opts, hubprinter.JSONPB())
	case "tab", "table":
		if selectorOpts.follow {
			return fmt.Errorf("table output format is not compatible with follow mode")
		}
		opts = append(opts, hubprinter.Tab())
	case "":
		// no format specified, choose most appropriate format based on
		// user provided flags
		if selectorOpts.follow {
			opts = append(opts, hubprinter.Compact())
		} else {
			opts = append(opts, hubprinter.Tab())
		}
	default:
		return fmt.Errorf("invalid output format: %s", formattingOpts.output)
	}
	if otherOpts.ignoreStderr {
		opts = append(opts, hubprinter.IgnoreStderr())
	}
	if formattingOpts.numeric {
		formattingOpts.enableIPTranslation = false
	}
	if formattingOpts.enableIPTranslation {
		opts = append(opts, hubprinter.WithIPTranslation())
	}
	if debug {
		opts = append(opts, hubprinter.WithDebug())
	}
	if formattingOpts.nodeName {
		opts = append(opts, hubprinter.WithNodeName())
	}
	printer = hubprinter.New(opts...)
	return nil
}

func runObserve(conn *grpc.ClientConn, ofilter *observeFilter) error {
	// convert selectorOpts.since into a param for GetFlows
	var since, until *timestamppb.Timestamp
	if selectorOpts.since != "" {
		st, err := hubtime.FromString(selectorOpts.since)
		if err != nil {
			return fmt.Errorf("failed to parse the since time: %v", err)
		}

		since = timestamppb.New(st)
		if err := since.CheckValid(); err != nil {
			return fmt.Errorf("failed to convert `since` timestamp to proto: %v", err)
		}
		// Set the until field if both --since and --until options are specified and --follow
		// is not specified. If --since is specified but --until is not, the server sets the
		// --until option to the current timestamp.
		if selectorOpts.until != "" && !selectorOpts.follow {
			ut, err := hubtime.FromString(selectorOpts.until)
			if err != nil {
				return fmt.Errorf("failed to parse the until time: %v", err)
			}
			until = timestamppb.New(ut)
			if err := until.CheckValid(); err != nil {
				return fmt.Errorf("failed to convert `until` timestamp to proto: %v", err)
			}
		}
	}

	if since == nil && until == nil {
		switch {
		case selectorOpts.all:
			// all is an alias for last=uint64_max
			selectorOpts.last = ^uint64(0)
		case selectorOpts.last == 0:
			// no specific parameters were provided, just a vanilla `hubble observe`
			selectorOpts.last = defaults.FlowPrintCount
		}
	}

	var (
		wl []*pb.FlowFilter
		bl []*pb.FlowFilter
	)
	if ofilter.whitelist != nil {
		wl = ofilter.whitelist.flowFilters()
	}
	if ofilter.blacklist != nil {
		bl = ofilter.blacklist.flowFilters()
	}

	client := observer.NewObserverClient(conn)
	req := &observer.GetFlowsRequest{
		Number:    selectorOpts.last,
		Follow:    selectorOpts.follow,
		Whitelist: wl,
		Blacklist: bl,
		Since:     since,
		Until:     until,
	}

	return getFlows(client, req)
}

func getFlows(client observer.ObserverClient, req *observer.GetFlowsRequest) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b, err := client.GetFlows(ctx, req)
	if err != nil {
		return err
	}

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt)
		select {
		case <-sigs:
		case <-ctx.Done():
			signal.Stop(sigs)
		}
		cancel()
	}()

	defer printer.Close()

	for {
		getFlowResponse, err := b.Recv()
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

		if err = printer.WriteGetFlowsResponse(getFlowResponse); err != nil {
			return err
		}
	}
}
