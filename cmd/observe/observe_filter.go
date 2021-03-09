// Copyright 2019-2020 Authors of Hubble
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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	pb "github.com/cilium/cilium/api/v1/flow"
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
	left, right *pb.FlowFilter

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

func (f *filterTracker) apply(update func(*pb.FlowFilter)) {
	f.applyLeft(update)
	f.applyRight(update)
}

func (f *filterTracker) applyLeft(update func(*pb.FlowFilter)) {
	if f.left == nil {
		f.left = &pb.FlowFilter{}
	}
	update(f.left)
}

func (f *filterTracker) applyRight(update func(*pb.FlowFilter)) {
	if f.right == nil {
		f.right = &pb.FlowFilter{}
	}
	update(f.right)
}

func (f *filterTracker) flowFilters() []*pb.FlowFilter {
	if f.left == nil && f.right == nil {
		return nil
	} else if proto.Equal(f.left, f.right) {
		return []*pb.FlowFilter{f.left}
	}

	filters := []*pb.FlowFilter{}
	if f.left != nil {
		filters = append(filters, f.left)
	}
	if f.right != nil {
		filters = append(filters, f.right)
	}
	return filters
}

// Implements pflag.Value
type observeFilter struct {
	whitelist *filterTracker
	blacklist *filterTracker

	// tracks if the next dispatched filter is going into blacklist or
	// whitelist. Blacklist is only triggered by `--not` and has to be set for
	// every blacklisted filter, i.e. `--not pod-ip 127.0.0.1 --not pod-ip
	// 2.2.2.2`.
	blacklisting bool

	conflicts [][]string // conflict config
}

func newObserveFilter() *observeFilter {
	return &observeFilter{
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
			{"protocol"},
			{"port", "to-port"},
			{"port", "from-port"},
			{"identity", "to-identity"},
			{"identity", "from-identity"},
			{"node-name"},
			{"tcp-flags"},
		},
	}
}

func (of *observeFilter) hasChanged(list []string, name string) bool {
	for _, c := range list {
		if c == name {
			return true
		}
	}
	return false
}

func (of *observeFilter) checkConflict(t *filterTracker, name, val string) error {
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

func parseTCPFlags(val string) (*pb.TCPFlags, error) {
	flags := &pb.TCPFlags{}
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

func ipVersion(v string) pb.IPVersion {
	switch strings.ToLower(v) {
	case "4", "v4", "ipv4", "ip4":
		return pb.IPVersion_IPv4
	case "6", "v6", "ipv6", "ip6":
		return pb.IPVersion_IPv6
	}
	return pb.IPVersion_IP_NOT_USED
}

func (of *observeFilter) Set(name, val string, track bool) error {
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

// agentEventSubtypes are the valid agent event sub-types. This map is neccesary
// because the sub-type strings in monitorAPI.AgentNotifications contain
// upper-case characters and spaces which are inconvenient to pass as CLI filter
// arguments.
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

func (of *observeFilter) set(f *filterTracker, name, val string, track bool) error {
	// track the change if this is non-default user operation
	wipe := false
	if track {
		wipe = f.add(name)

		if err := of.checkConflict(f, name, val); err != nil {
			return err
		}
	}

	switch name {
	// fqdn filters
	case "fqdn":
		f.applyLeft(func(f *pb.FlowFilter) {
			f.SourceFqdn = append(f.SourceFqdn, val)
		})
		f.applyRight(func(f *pb.FlowFilter) {
			f.DestinationFqdn = append(f.DestinationFqdn, val)
		})
	case "from-fqdn":
		f.apply(func(f *pb.FlowFilter) {
			f.SourceFqdn = append(f.SourceFqdn, val)
		})
	case "to-fqdn":
		f.apply(func(f *pb.FlowFilter) {
			f.DestinationFqdn = append(f.DestinationFqdn, val)
		})

	// pod filters
	case "pod":
		f.applyLeft(func(f *pb.FlowFilter) {
			f.SourcePod = append(f.SourcePod, val)
		})
		f.applyRight(func(f *pb.FlowFilter) {
			f.DestinationPod = append(f.DestinationPod, val)
		})
	case "from-pod":
		f.apply(func(f *pb.FlowFilter) {
			f.SourcePod = append(f.SourcePod, val)
		})
	case "to-pod":
		f.apply(func(f *pb.FlowFilter) {
			f.DestinationPod = append(f.DestinationPod, val)
		})
	// ip filters
	case "ip":
		f.applyLeft(func(f *pb.FlowFilter) {
			f.SourceIp = append(f.SourceIp, val)
		})
		f.applyRight(func(f *pb.FlowFilter) {
			f.DestinationIp = append(f.DestinationIp, val)
		})
	case "from-ip":
		f.apply(func(f *pb.FlowFilter) {
			f.SourceIp = append(f.SourceIp, val)
		})
	case "to-ip":
		f.apply(func(f *pb.FlowFilter) {
			f.DestinationIp = append(f.DestinationIp, val)
		})
	// ip version filters
	case "ipv4":
		f.apply(func(f *pb.FlowFilter) {
			f.IpVersion = append(f.IpVersion, pb.IPVersion_IPv4)
		})
	case "ipv6":
		f.apply(func(f *pb.FlowFilter) {
			f.IpVersion = append(f.IpVersion, pb.IPVersion_IPv6)
		})
	case "ip-version":
		f.apply(func(f *pb.FlowFilter) {
			f.IpVersion = append(f.IpVersion, ipVersion(val))
		})
	// label filters
	case "label":
		f.applyLeft(func(f *pb.FlowFilter) {
			f.SourceLabel = append(f.SourceLabel, val)
		})
		f.applyRight(func(f *pb.FlowFilter) {
			f.DestinationLabel = append(f.DestinationLabel, val)
		})
	case "from-label":
		f.apply(func(f *pb.FlowFilter) {
			f.SourceLabel = append(f.SourceLabel, val)
		})
	case "to-label":
		f.apply(func(f *pb.FlowFilter) {
			f.DestinationLabel = append(f.DestinationLabel, val)
		})

	// namespace filters (translated to pod filters)
	case "namespace":
		f.applyLeft(func(f *pb.FlowFilter) {
			f.SourcePod = append(f.SourcePod, val+"/")
		})
		f.applyRight(func(f *pb.FlowFilter) {
			f.DestinationPod = append(f.DestinationPod, val+"/")
		})
	case "from-namespace":
		f.apply(func(f *pb.FlowFilter) {
			f.SourcePod = append(f.SourcePod, val+"/")
		})
	case "to-namespace":
		f.apply(func(f *pb.FlowFilter) {
			f.DestinationPod = append(f.DestinationPod, val+"/")
		})
	// service filters
	case "service":
		f.applyLeft(func(f *pb.FlowFilter) {
			f.SourceService = append(f.SourceService, val)
		})
		f.applyRight(func(f *pb.FlowFilter) {
			f.DestinationService = append(f.DestinationService, val)
		})
	case "from-service":
		f.apply(func(f *pb.FlowFilter) {
			f.SourceService = append(f.SourceService, val)
		})
	case "to-service":
		f.apply(func(f *pb.FlowFilter) {
			f.DestinationService = append(f.DestinationService, val)
		})

	// port filters
	case "port":
		f.applyLeft(func(f *pb.FlowFilter) {
			f.SourcePort = append(f.SourcePort, val)
		})
		f.applyRight(func(f *pb.FlowFilter) {
			f.DestinationPort = append(f.DestinationPort, val)
		})
	case "from-port":
		f.apply(func(f *pb.FlowFilter) {
			f.SourcePort = append(f.SourcePort, val)
		})
	case "to-port":
		f.apply(func(f *pb.FlowFilter) {
			f.DestinationPort = append(f.DestinationPort, val)
		})

	case "verdict":
		if wipe {
			f.apply(func(f *pb.FlowFilter) {
				f.Verdict = nil
			})
		}

		vv, ok := pb.Verdict_value[val]
		if !ok {
			return fmt.Errorf("invalid --verdict value: %v", val)
		}
		f.apply(func(f *pb.FlowFilter) {
			f.Verdict = append(f.Verdict, pb.Verdict(vv))
		})

	case "http-status":
		f.apply(func(f *pb.FlowFilter) {
			f.HttpStatusCode = append(f.HttpStatusCode, val)
		})

	case "http-method":
		f.apply(func(f *pb.FlowFilter) {
			f.HttpMethod = append(f.HttpMethod, val)
		})

	case "http-path":
		f.apply(func(f *pb.FlowFilter) {
			f.HttpPath = append(f.HttpPath, val)
		})

	case "type":
		if wipe {
			f.apply(func(f *pb.FlowFilter) {
				f.EventType = nil
			})
		}

		typeFilter := &pb.EventTypeFilter{}

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
		f.apply(func(f *pb.FlowFilter) {
			f.EventType = append(f.EventType, typeFilter)
		})
	case "protocol":
		f.apply(func(f *pb.FlowFilter) {
			f.Protocol = append(f.Protocol, val)
		})

	// identity filters
	case "identity":
		i, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid security identity: %s: %s", val, err)
		}
		identity := uint32(i)
		f.applyLeft(func(f *pb.FlowFilter) {
			f.SourceIdentity = append(f.SourceIdentity, identity)
		})
		f.applyRight(func(f *pb.FlowFilter) {
			f.DestinationIdentity = append(f.DestinationIdentity, identity)
		})
	case "from-identity":
		i, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid security identity: %s: %s", val, err)
		}
		identity := uint32(i)
		f.apply(func(f *pb.FlowFilter) {
			f.SourceIdentity = append(f.SourceIdentity, identity)
		})
	case "to-identity":
		i, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid security identity: %s: %s", val, err)
		}
		identity := uint32(i)
		f.apply(func(f *pb.FlowFilter) {
			f.DestinationIdentity = append(f.DestinationIdentity, identity)
		})

	// node name filters
	case "node-name":
		f.apply(func(f *pb.FlowFilter) {
			f.NodeName = append(f.NodeName, val)
		})

	// TCP Flags filter
	case "tcp-flags":
		flags, err := parseTCPFlags(val)
		if err != nil {
			return err
		}
		f.apply(func(f *pb.FlowFilter) {
			f.TcpFlags = append(f.TcpFlags, flags)
		})
	}

	return nil
}

func (of observeFilter) Type() string {
	return "filter"
}

// Small dispatcher on top of a filter that allows all the filter arguments to
// flow through the same object. By default, Cobra doesn't call `Set()` with the
// name of the argument, only it's value.
type filterDispatch struct {
	*observeFilter

	name string
	def  []string
}

func (d filterDispatch) Set(s string) error {
	return d.observeFilter.Set(d.name, s, true)
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
	of *observeFilter,
	desc string,
) (pflag.Value, flagName, flagDesc) {
	return &filterDispatch{
		name:          name,
		observeFilter: of,
	}, name, desc
}

func filterVarP(
	name string,
	short string,
	of *observeFilter,
	def []string,
	desc string,
) (pflag.Value, flagName, shortName, flagDesc) {
	d := &filterDispatch{
		name:          name,
		def:           def,
		observeFilter: of,
	}
	for _, val := range def {
		d.observeFilter.Set(name, val, false /* do not track */)

	}
	return d, name, short, desc
}
