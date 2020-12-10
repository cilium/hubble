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
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	pb "github.com/cilium/cilium/api/v1/flow"
	observerpb "github.com/cilium/cilium/api/v1/observer"
	relaypb "github.com/cilium/cilium/api/v1/relay"
	v1 "github.com/cilium/cilium/pkg/hubble/api/v1"
	"github.com/cilium/cilium/pkg/monitor/api"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Printer for flows.
type Printer struct {
	opts        Options
	line        int
	tw          *tabwriter.Writer
	jsonEncoder *json.Encoder
}

type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) write(a ...interface{}) {
	if ew.err != nil {
		return
	}
	_, ew.err = fmt.Fprint(ew.w, a...)
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
	case JSONOutput, JSONPBOutput:
		p.jsonEncoder = json.NewEncoder(p.opts.w)
	}

	return p
}

const (
	tab     = "\t"
	newline = "\n"

	dictSeparator = "------------"

	nodeNamesCutOff = 50
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
	_, err := fmt.Fprintln(p.opts.werr, msg)
	return err
}

// GetPorts returns source and destination port of a flow.
func (p *Printer) GetPorts(f v1.Flow) (string, string) {
	l4 := f.GetL4()
	if l4 == nil {
		return "", ""
	}
	switch l4.Protocol.(type) {
	case *pb.Layer4_TCP:
		return strconv.Itoa(int(l4.GetTCP().SourcePort)), strconv.Itoa(int(l4.GetTCP().DestinationPort))
	case *pb.Layer4_UDP:
		return strconv.Itoa(int(l4.GetUDP().SourcePort)), strconv.Itoa(int(l4.GetUDP().DestinationPort))
	default:
		return "", ""
	}
}

// GetHostNames returns source and destination hostnames of a flow.
func (p *Printer) GetHostNames(f v1.Flow) (string, string) {
	var srcNamespace, dstNamespace, srcPodName, dstPodName, srcSvcName, dstSvcName string
	if f == nil {
		return "", ""
	}

	if f.GetIP() == nil {
		if eth := f.GetEthernet(); eth != nil {
			return eth.GetSource(), eth.GetDestination()
		}
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
	if svc := f.GetSourceService(); svc != nil {
		srcNamespace = svc.Namespace
		srcSvcName = svc.Name
	}
	if svc := f.GetDestinationService(); svc != nil {
		dstNamespace = svc.Namespace
		dstSvcName = svc.Name
	}
	srcPort, dstPort := p.GetPorts(f)
	src := p.Hostname(f.GetIP().Source, srcPort, srcNamespace, srcPodName, srcSvcName, f.GetSourceNames())
	dst := p.Hostname(f.GetIP().Destination, dstPort, dstNamespace, dstPodName, dstSvcName, f.GetDestinationNames())
	return src, dst
}

func fmtTimestamp(ts *timestamppb.Timestamp) string {
	if !ts.IsValid() {
		return "N/A"
	}
	// TODO: support more date formats through options to `hubble observe`
	return ts.AsTime().Format(time.StampMilli)
}

// GetFlowType returns the type of a flow as a string.
func GetFlowType(f v1.Flow) string {
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

	switch f.GetEventType().GetType() {
	case api.MessageTypeTrace:
		return api.TraceObservationPoint(uint8(f.GetEventType().GetSubType()))
	case api.MessageTypeDrop:
		return api.DropReason(uint8(f.GetEventType().GetSubType()))
	case api.MessageTypePolicyVerdict:
		switch f.GetVerdict() {
		case pb.Verdict_FORWARDED:
			return api.PolicyMatchType(f.GetPolicyMatchType()).String()
		case pb.Verdict_DROPPED:
			return api.DropReason(uint8(f.GetDropReason()))
		}
	}

	return "UNKNOWN"
}

// WriteProtoFlow writes v1.Flow into the output writer.
func (p *Printer) WriteProtoFlow(res *observerpb.GetFlowsResponse) error {
	f := res.GetFlow()

	switch p.opts.output {
	case TabOutput:
		ew := &errWriter{w: p.tw}
		src, dst := p.GetHostNames(f)

		if p.line == 0 {
			ew.write("TIMESTAMP", tab)
			if p.opts.nodeName {
				ew.write("NODE", tab)
			}
			ew.write(
				"SOURCE", tab,
				"DESTINATION", tab,
				"TYPE", tab,
				"VERDICT", tab,
				"SUMMARY", newline,
			)
		}
		ew.write(fmtTimestamp(f.GetTime()), tab)
		if p.opts.nodeName {
			ew.write(f.GetNodeName(), tab)
		}
		ew.write(
			src, tab,
			dst, tab,
			GetFlowType(f), tab,
			f.GetVerdict().String(), tab,
			f.GetSummary(), newline,
		)
		if ew.err != nil {
			return fmt.Errorf("failed to write out packet: %v", ew.err)
		}
	case DictOutput:
		ew := &errWriter{w: p.opts.w}
		src, dst := p.GetHostNames(f)

		if p.line != 0 {
			// TODO: line length?
			ew.write(dictSeparator)
		}

		// this is a little crude, but will do for now. should probably find the
		// longest header and auto-format the keys
		ew.write("  TIMESTAMP: ", fmtTimestamp(f.GetTime()), newline)
		if p.opts.nodeName {
			ew.write("       NODE: ", f.GetNodeName(), newline)
		}
		ew.write(
			"     SOURCE: ", src, newline,
			"DESTINATION: ", dst, newline,
			"       TYPE: ", GetFlowType(f), newline,
			"    VERDICT: ", f.GetVerdict().String(), newline,
			"    SUMMARY: ", f.GetSummary(), newline,
		)
		if ew.err != nil {
			return fmt.Errorf("failed to write out packet: %v", ew.err)
		}
	case CompactOutput:
		var node string
		src, dst := p.GetHostNames(f)

		if p.opts.nodeName {
			node = fmt.Sprintf(" [%s]", f.GetNodeName())
		}
		_, err := fmt.Fprintf(p.opts.w,
			"%s%s: %s -> %s %s %s (%s)\n",
			fmtTimestamp(f.GetTime()),
			node,
			src,
			dst,
			GetFlowType(f),
			f.GetVerdict().String(),
			f.GetSummary())
		if err != nil {
			return fmt.Errorf("failed to write out packet: %v", err)
		}
	case JSONOutput:
		return p.jsonEncoder.Encode(f)
	case JSONPBOutput:
		return p.jsonEncoder.Encode(res)
	}
	p.line++
	return nil
}

// joinWithCutOff performs a strings.Join, but will omit elements if the
// resulting string is longer than targetLen. The resulting string may be
// longer than targetLen, as it will always print at least one element.
// The number of omitted elements is appended to the resulting string as
// " (and N more)".
func joinWithCutOff(elems []string, sep string, targetLen int) string {
	strLen := 0
	end := len(elems)
	for i, elem := range elems {
		strLen += len(elem) + len(sep)
		if strLen > targetLen && i > 0 {
			end = i
			break
		}
	}

	joined := strings.Join(elems[:end], sep)
	omitted := len(elems) - end
	if omitted == 0 {
		return joined
	}

	return fmt.Sprintf("%s (and %d more)", joined, omitted)
}

// WriteProtoNodeStatusEvent writes a node status event into the error stream
func (p *Printer) WriteProtoNodeStatusEvent(r *observerpb.GetFlowsResponse) error {
	s := r.GetNodeStatus()
	if s == nil {
		return errors.New("not a node status event")
	}

	if !p.opts.enableDebug {
		switch s.GetStateChange() {
		case relaypb.NodeState_NODE_ERROR, relaypb.NodeState_NODE_UNAVAILABLE:
			break
		default:
			// skips informal messages in non-debug mode
			return nil
		}
	}

	switch p.opts.output {
	case JSONOutput, JSONPBOutput:
		return json.NewEncoder(p.opts.werr).Encode(r)
	case DictOutput:
		// this is a bit crude, but in case stdout and stderr are interleaved,
		// we want to make sure the separators are still printed to clearly
		// separate flows from node events.
		if p.line != 0 {
			_, err := fmt.Fprintln(p.opts.w, dictSeparator)
			if err != nil {
				return err
			}
		} else {
			p.line++
		}
		nodeNames := joinWithCutOff(s.NodeNames, ", ", nodeNamesCutOff)
		message := "N/A"
		if m := s.GetMessage(); len(m) != 0 {
			message = strconv.Quote(m)
		}
		_, err := fmt.Fprint(p.opts.werr,
			"  TIMESTAMP: ", fmtTimestamp(r.GetTime()), newline,
			"      STATE: ", s.StateChange.String(), newline,
			"      NODES: ", nodeNames, newline,
			"    MESSAGE: ", message, newline,
		)
		if err != nil {
			return fmt.Errorf("failed to write out node status: %v", err)
		}
	case TabOutput, CompactOutput:
		numNodes := len(s.NodeNames)
		nodeNames := joinWithCutOff(s.NodeNames, ", ", nodeNamesCutOff)
		prefix := fmt.Sprintf("%s [%s]", fmtTimestamp(r.GetTime()), r.GetNodeName())
		msg := fmt.Sprintf("%s: unknown node status event: %+v", prefix, s)
		switch s.StateChange {
		case relaypb.NodeState_NODE_CONNECTED:
			msg = fmt.Sprintf("%s: Receiving flows from %d nodes: %s", prefix, numNodes, nodeNames)
		case relaypb.NodeState_NODE_UNAVAILABLE:
			msg = fmt.Sprintf("%s: %d nodes are unavailable: %s", prefix, numNodes, nodeNames)
		case relaypb.NodeState_NODE_GONE:
			msg = fmt.Sprintf("%s: %d nodes removed from cluster: %s", prefix, numNodes, nodeNames)
		case relaypb.NodeState_NODE_ERROR:
			msg = fmt.Sprintf("%s: Error %q on %d nodes: %s", prefix, s.Message, numNodes, nodeNames)
		}

		return p.WriteErr(msg)
	}

	return nil
}

func formatServiceAddr(a *pb.ServiceUpsertNotificationAddr) string {
	return net.JoinHostPort(a.Ip, strconv.Itoa(int(a.Port)))
}

func getAgentEventDetails(e *pb.AgentEvent) string {
	switch e.GetType() {
	case pb.AgentEventType_AGENT_EVENT_UNKNOWN:
		if u := e.GetUnknown(); u != nil {
			return fmt.Sprintf("type: %s, notification: %s", u.Type, u.Notification)
		}
	case pb.AgentEventType_AGENT_STARTED:
		if a := e.GetAgentStart(); a != nil {
			return fmt.Sprintf("start time: %s", fmtTimestamp(a.Time))
		}
	case pb.AgentEventType_POLICY_UPDATED, pb.AgentEventType_POLICY_DELETED:
		if p := e.GetPolicyUpdate(); p != nil {
			return fmt.Sprintf("labels: [%s], revision: %d, count: %d",
				strings.Join(p.Labels, ","), p.Revision, p.RuleCount)
		}
	case pb.AgentEventType_ENDPOINT_REGENERATE_SUCCESS, pb.AgentEventType_ENDPOINT_REGENERATE_FAILURE:
		if r := e.GetEndpointRegenerate(); r != nil {
			var sb strings.Builder
			fmt.Fprintf(&sb, "id: %d, labels: [%s]", r.Id, strings.Join(r.Labels, ","))
			if re := r.Error; re != "" {
				fmt.Fprintf(&sb, ", error: %s", re)
			}
			return sb.String()
		}
	case pb.AgentEventType_ENDPOINT_CREATED, pb.AgentEventType_ENDPOINT_DELETED:
		if ep := e.GetEndpointUpdate(); ep != nil {
			var sb strings.Builder
			fmt.Fprintf(&sb, "id: %d", ep.Id)
			if n := ep.Namespace; n != "" {
				fmt.Fprintf(&sb, ", namespace: %s", n)
			}
			if n := ep.PodName; n != "" {
				fmt.Fprintf(&sb, ", pod name: %s", n)
			}
			return sb.String()
		}
	case pb.AgentEventType_IPCACHE_UPSERTED, pb.AgentEventType_IPCACHE_DELETED:
		if i := e.GetIpcacheUpdate(); i != nil {
			var sb strings.Builder
			fmt.Fprintf(&sb, "cidr: %s, identity: %d", i.Cidr, i.Identity)
			if i.OldIdentity != nil {
				fmt.Fprintf(&sb, ", old identity: %d", i.OldIdentity.Value)
			}
			if i.HostIp != "" {
				fmt.Fprintf(&sb, ", host ip: %s", i.HostIp)
			}
			if i.OldHostIp != "" {
				fmt.Fprintf(&sb, ", old host ip: %s", i.OldHostIp)
			}
			fmt.Fprintf(&sb, ", encrypt key: %d", i.EncryptKey)
			return sb.String()
		}
	case pb.AgentEventType_SERVICE_UPSERTED:
		if svc := e.GetServiceUpsert(); svc != nil {
			var sb strings.Builder
			fmt.Fprintf(&sb, "id: %d", svc.Id)
			if fe := svc.FrontendAddress; fe != nil {
				fmt.Fprintf(&sb, ", frontend: %s", formatServiceAddr(fe))
			}
			if bes := svc.BackendAddresses; len(bes) != 0 {
				backends := make([]string, 0, len(bes))
				for _, a := range bes {
					backends = append(backends, formatServiceAddr(a))
				}
				fmt.Fprintf(&sb, ", backends: [%s]", strings.Join(backends, ","))
			}
			if t := svc.Type; t != "" {
				fmt.Fprintf(&sb, ", type: %s", t)
			}
			if tp := svc.TrafficPolicy; tp != "" {
				fmt.Fprintf(&sb, ", traffic policy: %s", tp)
			}
			if ns := svc.Namespace; ns != "" {
				fmt.Fprintf(&sb, ", namespace: %s", ns)
			}
			if n := svc.Name; n != "" {
				fmt.Fprintf(&sb, ", name: %s", n)
			}
			return sb.String()
		}
	case pb.AgentEventType_SERVICE_DELETED:
		if s := e.GetServiceDelete(); s != nil {
			return fmt.Sprintf("id: %d", s.Id)
		}
	}
	return "UNKNOWN"
}

// WriteProtoAgentEvent writes v1.AgentEvent into the output writer.
func (p *Printer) WriteProtoAgentEvent(r *observerpb.GetFlowsResponse) error {
	e := r.GetAgentEvent()
	if e == nil {
		return errors.New("not an agent event")
	}

	switch p.opts.output {
	case JSONOutput:
		return p.jsonEncoder.Encode(e)
	case JSONPBOutput:
		return p.jsonEncoder.Encode(r)
	case DictOutput:
		ew := &errWriter{w: p.opts.w}

		if p.line != 0 {
			ew.write(dictSeparator)
		}

		ew.write("  TIMESTAMP: ", fmtTimestamp(r.GetTime()), newline)
		if p.opts.nodeName {
			ew.write("       NODE: ", r.GetNodeName(), newline)
		}
		ew.write(
			"       TYPE: ", e.GetType(), newline,
			"    DETAILS: ", getAgentEventDetails(e), newline,
		)
		if ew.err != nil {
			return fmt.Errorf("failed to write out agent event: %v", ew.err)
		}
	case TabOutput:
		ew := &errWriter{w: p.tw}
		if p.line == 0 {
			ew.write("TIMESTAMP", tab)
			if p.opts.nodeName {
				ew.write("NODE", tab)
			}
			ew.write(
				"TYPE", tab,
				"DETAILS", newline,
			)
		}
		ew.write(fmtTimestamp(r.GetTime()), tab)
		if p.opts.nodeName {
			ew.write(r.GetNodeName(), tab)
		}
		ew.write(
			e.GetType(), tab,
			getAgentEventDetails(e), newline,
		)
		if ew.err != nil {
			return fmt.Errorf("failed to write out agent event: %v", ew.err)
		}
	case CompactOutput:
		var node string

		if p.opts.nodeName {
			node = fmt.Sprintf(" [%s]", r.GetNodeName())
		}
		_, err := fmt.Fprintf(p.opts.w,
			"%s%s: %s\n",
			fmtTimestamp(r.GetTime()),
			node,
			e.GetType())
		if err != nil {
			return fmt.Errorf("failed to write out agent event: %v", err)
		}
	}
	p.line++
	return nil
}

// Hostname returns a "host:ip" formatted pair for the given ip and port. If
// port is empty, only the host is returned. The host part is either the pod or
// service name (if set), or a comma-separated list of domain names (if set),
// or just the ip address if EnableIPTranslation is false and/or there are no
// pod nor service name and domain names.
func (p *Printer) Hostname(ip, port string, ns, pod, svc string, names []string) (host string) {
	host = ip
	if p.opts.enableIPTranslation {
		if pod != "" {
			// path.Join omits the slash if ns is empty
			host = path.Join(ns, pod)
		} else if svc != "" {
			host = path.Join(ns, svc)
		} else if len(names) != 0 {
			host = strings.Join(names, ",")
		}
	}

	if port != "" && port != "0" {
		return net.JoinHostPort(host, port)
	}

	return host
}

// WriteGetFlowsResponse prints GetFlowsResponse according to the printer configuration.
func (p *Printer) WriteGetFlowsResponse(res *observerpb.GetFlowsResponse) error {
	switch r := res.GetResponseTypes().(type) {
	case *observerpb.GetFlowsResponse_Flow:
		return p.WriteProtoFlow(res)
	case *observerpb.GetFlowsResponse_NodeStatus:
		return p.WriteProtoNodeStatusEvent(res)
	case *observerpb.GetFlowsResponse_AgentEvent:
		return p.WriteProtoAgentEvent(res)
	case nil:
		return nil
	default:
		if p.opts.enableDebug {
			msg := fmt.Sprintf("unknown response type: %+v", r)
			return p.WriteErr(msg)
		}
		return nil
	}
}
