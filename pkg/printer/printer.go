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

	"github.com/cilium/cilium/pkg/monitor/api"
	pb "github.com/cilium/hubble/api/v1/flow"
	v1 "github.com/cilium/hubble/pkg/api/v1"
	"github.com/cilium/hubble/pkg/format"
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

func getPorts(f v1.Flow) (string, string) {
	l4 := f.GetL4()
	if l4 == nil {
		return "", ""
	}
	switch l4.Protocol.(type) {
	case *pb.Layer4_TCP:
		return format.TCPPort(layers.TCPPort(l4.GetTCP().SourcePort)), format.TCPPort(layers.TCPPort(l4.GetTCP().DestinationPort))
	case *pb.Layer4_UDP:
		return format.UDPPort(layers.UDPPort(l4.GetUDP().SourcePort)), format.UDPPort(layers.UDPPort(l4.GetUDP().DestinationPort))
	default:
		return "", ""
	}
}

func getHostNames(f v1.Flow) (string, string) {
	var srcNamespace, dstNamespace, srcPodName, dstPodName string
	if f == nil || f.GetIP() == nil {
		return "", ""
	}
	if src := f.GetSource(); src != nil {
		srcNamespace = src.Namespace
		srcPodName = src.PodName
	}
	if dst := f.GetDestination(); dst != nil {
		dstNamespace = dst.Namespace
		dstPodName = dst.PodName
	}
	srcPort, dstPort := getPorts(f)
	src := format.Hostname(f.GetIP().Source, srcPort, srcNamespace, srcPodName, f.GetSourceNames())
	dst := format.Hostname(f.GetIP().Destination, dstPort, dstNamespace, dstPodName, f.GetSourceNames())
	return src, dst
}

func getTimestamp(f v1.Flow) string {
	if f == nil {
		return "N/A"
	}
	ts, err := types.TimestampFromProto(f.GetTime())
	if err != nil {
		return "N/A"
	}
	return format.MaybeTime(&ts)
}

func getFlowType(f v1.Flow) string {
	if f == nil || f.GetEventType() == nil {
		return "UNKNOWN"
	}
	if l7 := f.GetL7(); l7 != nil {
		l7Protocol := "l7"
		l7Type := strings.ToLower(l7.Type.String())
		switch l7.GetRecord().(type) {
		case *pb.Layer7_Http:
			l7Protocol = "http"
		case *pb.Layer7_Dns:
			l7Protocol = "dns"
		case *pb.Layer7_Kafka:
			l7Protocol = "kafka"
		}
		return l7Protocol + "-" + l7Type
	}
	if f.GetVerdict() == pb.Verdict_DROPPED {
		return api.DropReason(uint8(f.GetEventType().SubType))
	}
	return api.TraceObservationPoint(uint8(f.GetEventType().SubType))
}

// WriteProtoFlow writes v1.Flow into the output writer.
func (p *Printer) WriteProtoFlow(f v1.Flow) error {
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
			f.GetVerdict().String(), tab,
			f.GetSummary(), newline,
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
			"    VERDICT: ", f.GetVerdict().String(), newline,
			"    SUMMARY: ", f.GetSummary(), newline,
		)
		if err != nil {
			return fmt.Errorf("failed to write out packet: %v", err)
		}
	case CompactOutput:
		src, dst := getHostNames(f)
		_, err := fmt.Fprintf(p.opts.w,
			"%s [%s]: %s -> %s %s %s (%s)\n",
			getTimestamp(f),
			f.GetNodeName(),
			src,
			dst,
			getFlowType(f),
			f.GetVerdict().String(),
			f.GetSummary(),
		)
		if err != nil {
			return fmt.Errorf("failed to write out packet: %v", err)
		}
	case JSONOutput:
		return p.jsonEncoder.Encode(f)
	}
	p.line++
	return nil
}
