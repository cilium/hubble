// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

package observe

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	flowpb "github.com/cilium/cilium/api/v1/flow"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/proto"
)

type (
	flagName  = string
	flagDesc  = string
	shortName = string
)

type filterTracker struct {
	// the resulting filter will be `left OR right`
	left, right *flowpb.FlowFilter

	// which values were touched by the user. This is important because all of
	// the defaults need to be wiped the first time user touches a []string
	// value.
	changed []string
}

func (f filterTracker) String() string {
	ff := f.flowFilters()
	if bs, err := json.Marshal(ff); err == nil {
		return fmt.Sprintf("%v", string(bs))
	}
	return fmt.Sprintf("%v", ff)
}

func (f *filterTracker) add(name string) bool {
	for _, exists := range f.changed {
		if name == exists {
			return false
		}
	}

	// wipe the existing values if this is the first time usage of this
	// flag, otherwise defaults creep into the final set.
	f.changed = append(f.changed, name)

	return true
}

func (f *filterTracker) apply(update func(*flowpb.FlowFilter)) {
	f.applyLeft(update)
	f.applyRight(update)
}

func (f *filterTracker) applyLeft(update func(*flowpb.FlowFilter)) {
	if f.left == nil {
		f.left = &flowpb.FlowFilter{}
	}
	update(f.left)
}

func (f *filterTracker) applyRight(update func(*flowpb.FlowFilter)) {
	if f.right == nil {
		f.right = &flowpb.FlowFilter{}
	}
	update(f.right)
}

func (f *filterTracker) flowFilters() []*flowpb.FlowFilter {
	if f.left == nil && f.right == nil {
		return nil
	} else if proto.Equal(f.left, f.right) {
		return []*flowpb.FlowFilter{f.left}
	}

	filters := []*flowpb.FlowFilter{}
	if f.left != nil {
		filters = append(filters, f.left)
	}
	if f.right != nil {
		filters = append(filters, f.right)
	}
	return filters
}

// Implements pflag.Value
type flowFilter struct {
	whitelist *filterTracker
	blacklist *filterTracker

	// tracks if the next dispatched filter is going into blacklist or
	// whitelist. Blacklist is only triggered by `--not` and has to be set for
	// every blacklisted filter, i.e. `--not pod-ip 127.0.0.1 --not pod-ip
	// 2.2.2.2`.
	blacklisting bool

	conflicts [][]string // conflict config
}

func newFlowFilter() *flowFilter {
	return &flowFilter{
		conflicts: [][]string{
			{"from-fqdn", "from-ip", "from-namespace", "from-pod", "fqdn", "ip", "namespace", "pod"},
			{"to-fqdn", "to-ip", "to-namespace", "to-pod", "fqdn", "ip", "namespace", "pod"},
			{"label", "from-label"},
			{"label", "to-label"},
			{"service", "from-service"},
			{"service", "to-service"},
			{"verdict"},
			{"type"},
			{"http-status"},
			{"http-method"},
			{"http-path"},
			{"http-url"},
			{"protocol"},
			{"port", "to-port"},
			{"port", "from-port"},
			{"identity", "to-identity"},
			{"identity", "from-identity"},
			{"workload", "to-workload"},
			{"workload", "from-workload"},
			{"node-name"},
			{"tcp-flags"},
			{"uuid"},
			{"traffic-direction"},
		},
	}
}

func (of *flowFilter) hasChanged(list []string, name string) bool {
	for _, c := range list {
		if c == name {
			return true
		}
	}
	return false
}

func (of *flowFilter) checkConflict(t *filterTracker) error {
	// check for conflicts
	for _, group := range of.conflicts {
		for _, flag := range group {
			if of.hasChanged(t.changed, flag) {
				for _, conflict := range group {
					if flag != conflict && of.hasChanged(t.changed, conflict) {
						return fmt.Errorf(
							"filters --%s and --%s cannot be combined",
							flag, conflict,
						)
					}
				}
			}
		}
	}
	return nil
}

func parseTCPFlags(val string) (*flowpb.TCPFlags, error) {
	flags := &flowpb.TCPFlags{}
	s := strings.Split(val, ",")
	for _, f := range s {
		switch strings.ToUpper(f) {
		case "SYN":
			flags.SYN = true
		case "FIN":
			flags.FIN = true
		case "RST":
			flags.RST = true
		case "PSH":
			flags.PSH = true
		case "ACK":
			flags.ACK = true
		case "URG":
			flags.URG = true
		case "ECE":
			flags.ECE = true
		case "CWR":
			flags.CWR = true
		case "NS":
			flags.NS = true
		default:
			return nil, fmt.Errorf("unknown tcp flag: %s", f)
		}
	}
	return flags, nil
}

func ipVersion(v string) flowpb.IPVersion {
	switch strings.ToLower(v) {
	case "4", "v4", "ipv4", "ip4":
		return flowpb.IPVersion_IPv4
	case "6", "v6", "ipv6", "ip6":
		return flowpb.IPVersion_IPv6
	}
	return flowpb.IPVersion_IP_NOT_USED
}

func (of *flowFilter) Set(name, val string, track bool) error {
	// --not simply toggles the destination of the next filter into blacklist
	if name == "not" {
		if of.blacklisting {
			return errors.New("consecutive --not statements")
		}
		of.blacklisting = true
		return nil
	}

	if of.blacklisting {
		// --not only applies to a single filter so we turn off blacklisting
		of.blacklisting = false

		// lazy init blacklist
		if of.blacklist == nil {
			of.blacklist = &filterTracker{
				changed: []string{},
			}
		}
		return of.set(of.blacklist, name, val, track)
	}

	// lazy init whitelist
	if of.whitelist == nil {
		of.whitelist = &filterTracker{
			changed: []string{},
		}
	}

	return of.set(of.whitelist, name, val, track)
}

// agentEventSubtypes are the valid agent event sub-types. This map is
// necessary because the sub-type strings in monitorAPI.AgentNotifications
// contain upper-case characters and spaces which are inconvenient to pass as
// CLI filter arguments.
var agentEventSubtypes = map[string]monitorAPI.AgentNotification{
	"unspecified":                 monitorAPI.AgentNotifyUnspec,
	"message":                     monitorAPI.AgentNotifyGeneric,
	"agent-started":               monitorAPI.AgentNotifyStart,
	"policy-updated":              monitorAPI.AgentNotifyPolicyUpdated,
	"policy-deleted":              monitorAPI.AgentNotifyPolicyDeleted,
	"endpoint-regenerate-success": monitorAPI.AgentNotifyEndpointRegenerateSuccess,
	"endpoint-regenerate-failure": monitorAPI.AgentNotifyEndpointRegenerateFail,
	"endpoint-created":            monitorAPI.AgentNotifyEndpointCreated,
	"endpoint-deleted":            monitorAPI.AgentNotifyEndpointDeleted,
	"ipcache-upserted":            monitorAPI.AgentNotifyIPCacheUpserted,
	"ipcache-deleted":             monitorAPI.AgentNotifyIPCacheDeleted,
	"service-upserted":            monitorAPI.AgentNotifyServiceUpserted,
	"service-deleted":             monitorAPI.AgentNotifyServiceDeleted,
}

func (of *flowFilter) set(f *filterTracker, name, val string, track bool) error {
	// track the change if this is non-default user operation
	wipe := false
	if track {
		wipe = f.add(name)

		if err := of.checkConflict(f); err != nil {
			return err
		}
	}

	switch name {
	// flow identifier filter
	case "uuid":
		f.apply(func(f *flowpb.FlowFilter) {
			f.Uuid = append(f.Uuid, val)
		})
	// fqdn filters
	case "fqdn":
		f.applyLeft(func(f *flowpb.FlowFilter) {
			f.SourceFqdn = append(f.SourceFqdn, val)
		})
		f.applyRight(func(f *flowpb.FlowFilter) {
			f.DestinationFqdn = append(f.DestinationFqdn, val)
		})
	case "from-fqdn":
		f.apply(func(f *flowpb.FlowFilter) {
			f.SourceFqdn = append(f.SourceFqdn, val)
		})
	case "to-fqdn":
		f.apply(func(f *flowpb.FlowFilter) {
			f.DestinationFqdn = append(f.DestinationFqdn, val)
		})

	// pod filters
	case "pod":
		f.applyLeft(func(f *flowpb.FlowFilter) {
			f.SourcePod = append(f.SourcePod, val)
		})
		f.applyRight(func(f *flowpb.FlowFilter) {
			f.DestinationPod = append(f.DestinationPod, val)
		})
	case "from-pod":
		f.apply(func(f *flowpb.FlowFilter) {
			f.SourcePod = append(f.SourcePod, val)
		})
	case "to-pod":
		f.apply(func(f *flowpb.FlowFilter) {
			f.DestinationPod = append(f.DestinationPod, val)
		})
	// ip filters
	case "ip":
		f.applyLeft(func(f *flowpb.FlowFilter) {
			f.SourceIp = append(f.SourceIp, val)
		})
		f.applyRight(func(f *flowpb.FlowFilter) {
			f.DestinationIp = append(f.DestinationIp, val)
		})
	case "from-ip":
		f.apply(func(f *flowpb.FlowFilter) {
			f.SourceIp = append(f.SourceIp, val)
		})
	case "to-ip":
		f.apply(func(f *flowpb.FlowFilter) {
			f.DestinationIp = append(f.DestinationIp, val)
		})
	// ip version filters
	case "ipv4":
		f.apply(func(f *flowpb.FlowFilter) {
			f.IpVersion = append(f.IpVersion, flowpb.IPVersion_IPv4)
		})
	case "ipv6":
		f.apply(func(f *flowpb.FlowFilter) {
			f.IpVersion = append(f.IpVersion, flowpb.IPVersion_IPv6)
		})
	case "ip-version":
		f.apply(func(f *flowpb.FlowFilter) {
			f.IpVersion = append(f.IpVersion, ipVersion(val))
		})
	// label filters
	case "label":
		f.applyLeft(func(f *flowpb.FlowFilter) {
			f.SourceLabel = append(f.SourceLabel, val)
		})
		f.applyRight(func(f *flowpb.FlowFilter) {
			f.DestinationLabel = append(f.DestinationLabel, val)
		})
	case "from-label":
		f.apply(func(f *flowpb.FlowFilter) {
			f.SourceLabel = append(f.SourceLabel, val)
		})
	case "to-label":
		f.apply(func(f *flowpb.FlowFilter) {
			f.DestinationLabel = append(f.DestinationLabel, val)
		})

	// namespace filters (translated to pod filters)
	case "namespace":
		f.applyLeft(func(f *flowpb.FlowFilter) {
			f.SourcePod = append(f.SourcePod, val+"/")
		})
		f.applyRight(func(f *flowpb.FlowFilter) {
			f.DestinationPod = append(f.DestinationPod, val+"/")
		})
	case "from-namespace":
		f.apply(func(f *flowpb.FlowFilter) {
			f.SourcePod = append(f.SourcePod, val+"/")
		})
	case "to-namespace":
		f.apply(func(f *flowpb.FlowFilter) {
			f.DestinationPod = append(f.DestinationPod, val+"/")
		})
	// service filters
	case "service":
		f.applyLeft(func(f *flowpb.FlowFilter) {
			f.SourceService = append(f.SourceService, val)
		})
		f.applyRight(func(f *flowpb.FlowFilter) {
			f.DestinationService = append(f.DestinationService, val)
		})
	case "from-service":
		f.apply(func(f *flowpb.FlowFilter) {
			f.SourceService = append(f.SourceService, val)
		})
	case "to-service":
		f.apply(func(f *flowpb.FlowFilter) {
			f.DestinationService = append(f.DestinationService, val)
		})

	// port filters
	case "port":
		f.applyLeft(func(f *flowpb.FlowFilter) {
			f.SourcePort = append(f.SourcePort, val)
		})
		f.applyRight(func(f *flowpb.FlowFilter) {
			f.DestinationPort = append(f.DestinationPort, val)
		})
	case "from-port":
		f.apply(func(f *flowpb.FlowFilter) {
			f.SourcePort = append(f.SourcePort, val)
		})
	case "to-port":
		f.apply(func(f *flowpb.FlowFilter) {
			f.DestinationPort = append(f.DestinationPort, val)
		})

	case "trace-id":
		f.apply(func(f *flowpb.FlowFilter) {
			f.TraceId = append(f.TraceId, val)
		})

	case "verdict":
		if wipe {
			f.apply(func(f *flowpb.FlowFilter) {
				f.Verdict = nil
			})
		}

		vv, ok := flowpb.Verdict_value[val]
		if !ok {
			return fmt.Errorf("invalid --verdict value: %v", val)
		}
		f.apply(func(f *flowpb.FlowFilter) {
			f.Verdict = append(f.Verdict, flowpb.Verdict(vv))
		})

	case "http-status":
		f.apply(func(f *flowpb.FlowFilter) {
			f.HttpStatusCode = append(f.HttpStatusCode, val)
		})

	case "http-method":
		f.apply(func(f *flowpb.FlowFilter) {
			f.HttpMethod = append(f.HttpMethod, val)
		})

	case "http-path":
		f.apply(func(f *flowpb.FlowFilter) {
			f.HttpPath = append(f.HttpPath, val)
		})
	case "http-url":
		f.apply(func(f *flowpb.FlowFilter) {
			f.HttpUrl = append(f.HttpUrl, val)
		})

	case "type":
		if wipe {
			f.apply(func(f *flowpb.FlowFilter) {
				f.EventType = nil
			})
		}

		typeFilter := &flowpb.EventTypeFilter{}

		s := strings.SplitN(val, ":", 2)
		t, ok := monitorAPI.MessageTypeNames[s[0]]
		if ok {
			typeFilter.Type = int32(t)
		} else {
			t, err := strconv.ParseUint(s[0], 10, 32)
			if err != nil {
				return fmt.Errorf("unable to parse type '%s', not a known type name and unable to parse as numeric value: %s", s[0], err)
			}
			typeFilter.Type = int32(t)
		}

		if len(s) > 1 {
			switch t {
			case monitorAPI.MessageTypeTrace:
				for k, v := range monitorAPI.TraceObservationPoints {
					if s[1] == v {
						typeFilter.MatchSubType = true
						typeFilter.SubType = int32(k)
						break
					}
				}
			case monitorAPI.MessageTypeAgent:
				// See agentEventSubtypes godoc for why we're
				// not using monitorAPI.AgentNotifications here.
				if st, ok := agentEventSubtypes[s[1]]; ok {
					typeFilter.MatchSubType = true
					typeFilter.SubType = int32(st)
				}
			}
			if !typeFilter.MatchSubType {
				t, err := strconv.ParseUint(s[1], 10, 32)
				if err != nil {
					return fmt.Errorf("unable to parse event sub-type '%s', not a known sub-type name and unable to parse as numeric value: %s", s[1], err)
				}
				typeFilter.MatchSubType = true
				typeFilter.SubType = int32(t)
			}
		}
		f.apply(func(f *flowpb.FlowFilter) {
			f.EventType = append(f.EventType, typeFilter)
		})
	case "protocol":
		f.apply(func(f *flowpb.FlowFilter) {
			f.Protocol = append(f.Protocol, val)
		})

	// workload filters
	case "workload":
		workload := parseWorkload(val)
		f.applyLeft(func(f *flowpb.FlowFilter) {
			f.SourceWorkload = append(f.SourceWorkload, workload)
		})
		f.applyRight(func(f *flowpb.FlowFilter) {
			f.DestinationWorkload = append(f.DestinationWorkload, workload)
		})
	case "from-workload":
		workload := parseWorkload(val)
		f.apply(func(f *flowpb.FlowFilter) {
			f.SourceWorkload = append(f.SourceWorkload, workload)
		})
	case "to-workload":
		workload := parseWorkload(val)
		f.apply(func(f *flowpb.FlowFilter) {
			f.DestinationWorkload = append(f.DestinationWorkload, workload)
		})

	// identity filters
	case "identity":
		identity, err := parseIdentity(val)
		if err != nil {
			return fmt.Errorf("invalid security identity, expected one of %v or a numeric value", reservedIdentitiesNames())
		}
		f.applyLeft(func(f *flowpb.FlowFilter) {
			f.SourceIdentity = append(f.SourceIdentity, identity.Uint32())
		})
		f.applyRight(func(f *flowpb.FlowFilter) {
			f.DestinationIdentity = append(f.DestinationIdentity, identity.Uint32())
		})
	case "from-identity":
		identity, err := parseIdentity(val)
		if err != nil {
			return fmt.Errorf("invalid security identity, expected one of %v or a numeric value", reservedIdentitiesNames())
		}
		f.apply(func(f *flowpb.FlowFilter) {
			f.SourceIdentity = append(f.SourceIdentity, identity.Uint32())
		})
	case "to-identity":
		identity, err := parseIdentity(val)
		if err != nil {
			return fmt.Errorf("invalid security identity, expected one of %v or a numeric value", reservedIdentitiesNames())
		}
		f.apply(func(f *flowpb.FlowFilter) {
			f.DestinationIdentity = append(f.DestinationIdentity, identity.Uint32())
		})

	// node name filters
	case "node-name":
		f.apply(func(f *flowpb.FlowFilter) {
			f.NodeName = append(f.NodeName, val)
		})

	// TCP Flags filter
	case "tcp-flags":
		flags, err := parseTCPFlags(val)
		if err != nil {
			return err
		}
		f.apply(func(f *flowpb.FlowFilter) {
			f.TcpFlags = append(f.TcpFlags, flags)
		})

	// traffic direction filter
	case "traffic-direction":
		switch td := strings.ToLower(val); td {
		case "ingress":
			f.apply(func(f *flowpb.FlowFilter) {
				f.TrafficDirection = append(f.TrafficDirection, flowpb.TrafficDirection_INGRESS)
			})
		case "egress":
			f.apply(func(f *flowpb.FlowFilter) {
				f.TrafficDirection = append(f.TrafficDirection, flowpb.TrafficDirection_EGRESS)
			})
		default:
			return fmt.Errorf("%s: invalid traffic direction, expected ingress or egress", td)
		}
	}

	return nil
}

func (of flowFilter) Type() string {
	return "filter"
}

// Small dispatcher on top of a filter that allows all the filter arguments to
// flow through the same object. By default, Cobra doesn't call `Set()` with the
// name of the argument, only it's value.
type filterDispatch struct {
	*flowFilter

	name string
	def  []string
}

func (d filterDispatch) Set(s string) error {
	return d.flowFilter.Set(d.name, s, true)
}

// for some reason String() is used for default value in pflag/cobra
func (d filterDispatch) String() string {
	if len(d.def) == 0 {
		return ""
	}

	var b bytes.Buffer
	first := true
	b.WriteString("[")
	for _, def := range d.def {
		if first {
			first = false
		} else {
			// if not first, write a comma
			b.WriteString(",")
		}
		b.WriteString(def)
	}
	b.WriteString("]")

	return b.String()
}

func filterVar(
	name string,
	of *flowFilter,
	desc string,
) (pflag.Value, flagName, flagDesc) {
	return &filterDispatch{
		name:       name,
		flowFilter: of,
	}, name, desc
}

func filterVarP(
	name string,
	short string,
	of *flowFilter,
	def []string,
	desc string,
) (pflag.Value, flagName, shortName, flagDesc) {
	d := &filterDispatch{
		name:       name,
		def:        def,
		flowFilter: of,
	}
	for _, val := range def {
		d.flowFilter.Set(name, val, false /* do not track */)

	}
	return d, name, short, desc
}
