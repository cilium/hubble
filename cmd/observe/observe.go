// Copyright 2019-2021 Authors of Hubble
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
	"time"

	pb "github.com/cilium/cilium/api/v1/flow"
	"github.com/cilium/cilium/api/v1/observer"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	"github.com/cilium/hubble/cmd/common/config"
	"github.com/cilium/hubble/cmd/common/conn"
	"github.com/cilium/hubble/cmd/common/template"
	"github.com/cilium/hubble/pkg/defaults"
	"github.com/cilium/hubble/pkg/logger"
	hubprinter "github.com/cilium/hubble/pkg/printer"
	hubtime "github.com/cilium/hubble/pkg/time"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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

		timeFormat string

		enableIPTranslation bool
		nodeName            bool
		numeric             bool
		color               string
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

// eventTypes are the valid event types supported by observe. This corresponds
// to monitorAPI.MessageTypeNames, excluding MessageTypeNameAgent,
// MessageTypeNameDebug and MessageTypeNameRecCapture. These excluded message
// message types are not supported by observe but have separate sub-commands.
var eventTypes = []string{
	monitorAPI.MessageTypeNameDrop,
	monitorAPI.MessageTypeNameCapture,
	monitorAPI.MessageTypeNameTrace,
	monitorAPI.MessageTypeNameL7,
	monitorAPI.MessageTypeNamePolicyVerdict,
}

// New observer command.
func New(vp *viper.Viper) *cobra.Command {
	return newObserveCmd(vp, newObserveFilter())
}

func newObserveCmd(vp *viper.Viper, ofilter *observeFilter) *cobra.Command {
	observeCmd := &cobra.Command{
		Example: `* Piping flows to hubble observe

  Save output from 'hubble observe -o jsonpb' command to a file, and pipe it to
  the observe command later for offline processing. For example:

    hubble observe -o jsonpb --last 1000 > flows.json

  Then,

    cat flows.json | hubble observe

  Note that the observe command ignores --follow, --last, and server flags when it
  reads flows from stdin. The observe command processes and output flows in the same
  order they are read from stdin without sorting them by timestamp.`,
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
			req, err := getRequest(ofilter)
			if err != nil {
				return err
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
			defer cancel()

			var client observer.ObserverClient
			fi, err := os.Stdin.Stat()
			if err != nil {
				return err
			}
			if fi.Mode()&os.ModeNamedPipe != 0 {
				// read flows from stdin
				client = newIOReaderObserver(os.Stdin)
			} else {
				// read flows from a hubble server
				hubbleConn, err := conn.New(ctx, vp.GetString(config.KeyServer), vp.GetDuration(config.KeyTimeout))
				if err != nil {
					return err
				}
				defer hubbleConn.Close()
				client = observer.NewObserverClient(hubbleConn)
			}

			logger.Logger.WithField("request", req).Debug("Sending GetFlows request")
			if err := getFlows(ctx, client, req); err != nil {
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
	observeCmd.PersistentFlags().AddFlagSet(selectorFlags)

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
		fmt.Sprintf("Filter by event types TYPE[:SUBTYPE] (%v)", eventTypes)))
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

	// formatting flags only to `hubble observe`, but not sub-commands. Will be added to
	// generic formatting flags below.
	observeFormattingFlags := pflag.NewFlagSet("", pflag.ContinueOnError)
	observeFormattingFlags.BoolVarP(
		&formattingOpts.jsonOutput, "json", "j", false, "Deprecated. Use '--output json' instead.",
	)
	observeFormattingFlags.MarkDeprecated("json", "use '--output json' instead")
	observeFormattingFlags.BoolVar(
		&formattingOpts.compactOutput, "compact", false, "Deprecated. Use '--output compact' instead.",
	)
	observeFormattingFlags.MarkDeprecated("compact", "use '--output compact' instead")
	observeFormattingFlags.BoolVar(
		&formattingOpts.dictOutput, "dict", false, "Deprecated. Use '--output dict' instead.",
	)
	observeFormattingFlags.MarkDeprecated("dict", "use '--output dict' instead")
	observeFormattingFlags.BoolVar(
		&formattingOpts.numeric,
		"numeric",
		false,
		"Display all information in numeric form",
	)
	observeFormattingFlags.BoolVar(
		&formattingOpts.enableIPTranslation,
		"ip-translation",
		true,
		"Translate IP addresses to logical names such as pod name, FQDN, ...",
	)
	observeFormattingFlags.StringVar(
		&formattingOpts.color,
		"color", "auto",
		"Colorize the output when the output format is one of 'compact' or 'dict'. The value is one of 'auto' (default), 'always' or 'never'",
	)
	observeCmd.Flags().AddFlagSet(observeFormattingFlags)

	// generic formatting flags, available to `hubble observe`, including sub-commands.
	formattingFlags := pflag.NewFlagSet("Formatting", pflag.ContinueOnError)
	formattingFlags.StringVarP(
		&formattingOpts.output, "output", "o", "compact",
		`Specify the output format, one of:
 compact:  Compact output
 dict:     Each flow is shown as KEY:VALUE pair
 json:     JSON encoding
 jsonpb:   Output each GetFlowResponse according to proto3's JSON mapping
 table:    Tab-aligned columns
`)
	formattingFlags.BoolVarP(&formattingOpts.nodeName, "print-node-name", "", false, "Print node name in output")
	formattingFlags.StringVar(
		&formattingOpts.timeFormat, "time-format", "StampMilli",
		fmt.Sprintf(`Specify the time format for printing. This option does not apply to the json and jsonpb output type. One of:
  StampMilli:   %s
  RFC3339:      %s
  RFC3339Milli: %s
  RFC3339Micro: %s
  RFC3339Nano:  %s
  RFC1123Z:     %s
 `, time.StampMilli, time.RFC3339, hubtime.RFC3339Milli, hubtime.RFC3339Micro, time.RFC3339Nano, time.RFC1123Z),
	)
	observeCmd.PersistentFlags().AddFlagSet(formattingFlags)

	// other flags
	otherFlags := pflag.NewFlagSet("other", pflag.ContinueOnError)
	otherFlags.BoolVarP(
		&otherOpts.ignoreStderr, "silent-errors", "s", false, "Silently ignores errors and warnings")
	observeCmd.PersistentFlags().AddFlagSet(otherFlags)

	// advanced completion for flags
	observeCmd.RegisterFlagCompletionFunc("ip-version", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"none", "v4", "v6"}, cobra.ShellCompDirectiveDefault
	})
	observeCmd.RegisterFlagCompletionFunc("type", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return eventTypes, cobra.ShellCompDirectiveDefault
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
	observeCmd.RegisterFlagCompletionFunc("color", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"auto", "always", "never"}, cobra.ShellCompDirectiveDefault
	})
	observeCmd.RegisterFlagCompletionFunc("time-format", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return hubtime.FormatNames, cobra.ShellCompDirectiveDefault
	})

	// default value for when the flag is on the command line without any options
	observeCmd.Flags().Lookup("not").NoOptDefVal = "true"

	observeCmd.AddCommand(
		newAgentEventsCommand(vp, selectorFlags, formattingFlags, config.ServerFlags, otherFlags),
		newDebugEventsCommand(vp, selectorFlags, formattingFlags, config.ServerFlags, otherFlags),
	)

	formattingFlags.AddFlagSet(observeFormattingFlags)
	observeCmd.SetUsageTemplate(template.Usage(selectorFlags, filterFlags, formattingFlags, config.ServerFlags, otherFlags))

	return observeCmd
}

func handleArgs(ofilter *observeFilter, debug bool) (err error) {
	if ofilter.blacklisting {
		return errors.New("trailing --not found in the arguments")
	}

	// initialize the printer with any options that were passed in
	var opts = []hubprinter.Option{
		hubprinter.WithTimeFormat(hubtime.FormatNameToLayout(formattingOpts.timeFormat)),
		hubprinter.WithColor(formattingOpts.color),
	}

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

func getRequest(ofilter *observeFilter) (*observer.GetFlowsRequest, error) {
	// convert selectorOpts.since into a param for GetFlows
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

	req := &observer.GetFlowsRequest{
		Number:    selectorOpts.last,
		Follow:    selectorOpts.follow,
		Whitelist: wl,
		Blacklist: bl,
		Since:     since,
		Until:     until,
	}

	return req, nil
}

func getFlows(ctx context.Context, client observer.ObserverClient, req *observer.GetFlowsRequest) error {
	b, err := client.GetFlows(ctx, req)
	if err != nil {
		return err
	}
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
