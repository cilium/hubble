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

package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	pb "github.com/cilium/hubble/api/v1/observer"
	"github.com/cilium/hubble/pkg/format"

	"github.com/cilium/cilium/pkg/monitor/api"
	"github.com/francoispqt/gojay"
	"github.com/gogo/protobuf/types"
	"github.com/google/gopacket/layers"
)

// Encoder for flows.
type Encoder interface {
	Encode(v interface{}) error
}

// Printer for flows.
type Printer struct {
	opts        Options
	line        int
	tw          *tabwriter.Writer
	jsonEncoder Encoder
}

// New Printer.
func New(fopts ...Option) *Printer {
	// default options
	opts := Options{
		output: TabOutput,
		w:      os.Stdout,
		werr:   os.Stderr,
	}

	// apply optional parameters
	for _, fopt := range fopts {
		fopt(&opts)
	}

	p := &Printer{
		opts: opts,
	}

	switch opts.output {
	case TabOutput:
		// initialize tabwriter since it's going to be needed
		p.tw = tabwriter.NewWriter(opts.w, 2, 0, 3, ' ', 0)
	case JSONOutput:
		if opts.withJSONEncoder {
			p.jsonEncoder = json.NewEncoder(p.opts.w)
		} else {
			p.jsonEncoder = gojay.NewEncoder(p.opts.w)
		}
	}

	return p
}

const (
	tab     = "\t"
	newline = "\n"
)

// Close any outstanding operations going on in the printer.
func (p *Printer) Close() error {
	if p.tw != nil {
		return p.tw.Flush()
	}

	return nil
}

// WriteErr returns the given msg into the err writer defined in the printer.
func (p *Printer) WriteErr(msg string) error {
	_, err := fmt.Fprint(p.opts.werr, fmt.Sprintf("%s\n", msg))
	return err
}

func getPorts(f *pb.Flow) (string, string) {
	if f.L4 == nil {
		return "", ""
	}
	switch f.L4.Protocol.(type) {
	case *pb.Layer4_TCP:
		return format.TCPPort(layers.TCPPort(f.L4.GetTCP().SourcePort)), format.TCPPort(layers.TCPPort(f.L4.GetTCP().DestinationPort))
	case *pb.Layer4_UDP:
		return format.UDPPort(layers.UDPPort(f.L4.GetUDP().SourcePort)), format.UDPPort(layers.UDPPort(f.L4.GetUDP().DestinationPort))
	default:
		return "", ""
	}
}

func getHostNames(f *pb.Flow) (string, string) {
	var srcNamespace, dstNamespace, srcPodName, dstPodName string
	if f == nil || f.IP == nil {
		return "", ""
	}
	if f.Source != nil {
		srcNamespace = f.Source.Namespace
		srcPodName = f.Source.PodName
	}
	if f.Destination != nil {
		dstNamespace = f.Destination.Namespace
		dstPodName = f.Destination.PodName
	}
	srcPort, dstPort := getPorts(f)
	src := format.Hostname(f.IP.Source, srcPort, srcNamespace, srcPodName, f.SourceNames)
	dst := format.Hostname(f.IP.Destination, dstPort, dstNamespace, dstPodName, f.DestinationNames)
	return src, dst
}

func getTimestamp(f *pb.Flow) string {
	if f == nil {
		return "N/A"
	}
	ts, err := types.TimestampFromProto(f.Time)
	if err != nil {
		return "N/A"
	}
	return format.MaybeTime(&ts)
}

func getFlowType(f *pb.Flow) string {
	if f == nil || f.EventType == nil {
		return "UNKNOWN"
	}
	if f.L7 != nil {
		l7Protocol := "l7"
		l7Type := strings.ToLower(f.GetL7().Type.String())
		switch f.L7.GetRecord().(type) {
		case *pb.Layer7_Http:
			l7Protocol = "http"
		case *pb.Layer7_Dns:
			l7Protocol = "dns"
		case *pb.Layer7_Kafka:
			l7Protocol = "kafka"
		}
		return l7Protocol + "-" + l7Type
	}
	if f.Verdict == pb.Verdict_DROPPED {
		return api.DropReason(uint8(f.EventType.SubType))
	}
	return api.TraceObservationPoint(uint8(f.EventType.SubType))
}

// WriteProtoFlow writes pb.Flow into the output writer.
func (p *Printer) WriteProtoFlow(f *pb.Flow) error {
	switch p.opts.output {
	case TabOutput:
		if p.line == 0 {
			_, err := fmt.Fprint(p.tw,
				"TIMESTAMP", tab,
				"SOURCE", tab,
				"DESTINATION", tab,
				"TYPE", tab,
				"VERDICT", tab,
				"SUMMARY", newline,
			)
			if err != nil {
				return err
			}
		}
		src, dst := getHostNames(f)
		_, err := fmt.Fprint(p.tw,
			getTimestamp(f), tab,
			src, tab,
			dst, tab,
			getFlowType(f), tab,
			f.Verdict.String(), tab,
			f.Summary, newline,
		)
		if err != nil {
			return fmt.Errorf("failed to write out packet: %v", err)
		}
	case DictOutput:
		if p.line != 0 {
			// TODO: line length?
			_, err := fmt.Fprintln(p.opts.w, "------------")
			if err != nil {
				return err
			}
		}
		src, dst := getHostNames(f)
		// this is a little crude, but will do for now. should probably find the
		// longest header and auto-format the keys
		_, err := fmt.Fprint(p.opts.w,
			"  TIMESTAMP: ", getTimestamp(f), newline,
			"     SOURCE: ", src, newline,
			"DESTINATION: ", dst, newline,
			"       TYPE: ", getFlowType(f), newline,
			"    VERDICT: ", f.Verdict.String(), newline,
			"    SUMMARY: ", f.Summary, newline,
		)
		if err != nil {
			return fmt.Errorf("failed to write out packet: %v", err)
		}
	case CompactOutput:
		src, dst := getHostNames(f)
		_, err := fmt.Fprintf(p.opts.w,
			"%s [%s]: %s -> %s %s %s (%s)\n",
			getTimestamp(f),
			f.NodeName,
			src,
			dst,
			getFlowType(f),
			f.Verdict.String(),
			f.Summary,
		)
		if err != nil {
			return fmt.Errorf("failed to write out packet: %v", err)
		}
	case JSONOutput:
		f.Payload = nil
		return p.jsonEncoder.Encode(f)
	}
	p.line++
	return nil
}
