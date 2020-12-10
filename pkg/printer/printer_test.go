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
	"bytes"
	"strings"
	"testing"
	"time"

	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/cilium/cilium/api/v1/flow"
	observerpb "github.com/cilium/cilium/api/v1/observer"
)

func TestPrinter_WriteProtoFlow(t *testing.T) {
	buf := bytes.Buffer{}
	f := pb.Flow{
		Time: &timestamp.Timestamp{
			Seconds: 1234,
			Nanos:   567800000,
		},
		Type:     pb.FlowType_L3_L4,
		NodeName: "k8s1",
		Verdict:  pb.Verdict_DROPPED,
		IP: &pb.IP{
			Source:      "1.1.1.1",
			Destination: "2.2.2.2",
		},
		L4: &pb.Layer4{
			Protocol: &pb.Layer4_TCP{
				TCP: &pb.TCP{
					SourcePort:      31793,
					DestinationPort: 8080,
				},
			},
		},
		EventType: &pb.CiliumEventType{
			Type:    monitorAPI.MessageTypeDrop,
			SubType: 133,
		},
		Summary: "TCP Flags: SYN",
	}
	type args struct {
		f *pb.Flow
	}
	tests := []struct {
		name     string
		options  []Option
		args     args
		wantErr  bool
		expected string
	}{
		{
			name: "tabular",
			options: []Option{
				Writer(&buf),
			},
			args: args{
				f: &f,
			},
			wantErr: false,
			expected: `TIMESTAMP             SOURCE          DESTINATION    TYPE            VERDICT   SUMMARY
Jan  1 00:20:34.567   1.1.1.1:31793   2.2.2.2:8080   Policy denied   DROPPED   TCP Flags: SYN`,
		},
		{
			name: "tabular-with-node",
			options: []Option{
				WithNodeName(),
				Writer(&buf),
			},
			args: args{
				f: &f,
			},
			wantErr: false,
			expected: `TIMESTAMP             NODE   SOURCE          DESTINATION    TYPE            VERDICT   SUMMARY
Jan  1 00:20:34.567   k8s1   1.1.1.1:31793   2.2.2.2:8080   Policy denied   DROPPED   TCP Flags: SYN`,
		},
		{
			name: "compact",
			options: []Option{
				Compact(),
				Writer(&buf),
			},
			args: args{
				f: &f,
			},
			wantErr: false,
			expected: "Jan  1 00:20:34.567: " +
				"1.1.1.1:31793 -> 2.2.2.2:8080 " +
				"Policy denied DROPPED (TCP Flags: SYN)\n",
		},
		{
			name: "compact-with-node",
			options: []Option{
				Compact(),
				WithNodeName(),
				Writer(&buf),
			},
			args: args{
				f: &f,
			},
			wantErr: false,
			expected: "Jan  1 00:20:34.567 [k8s1]: " +
				"1.1.1.1:31793 -> 2.2.2.2:8080 " +
				"Policy denied DROPPED (TCP Flags: SYN)\n",
		},
		{
			name: "json",
			options: []Option{
				JSON(),
				Writer(&buf),
			},
			args: args{
				f: &f,
			},
			wantErr: false,
			expected: `{"time":"1970-01-01T00:20:34.567800Z",` +
				`"verdict":"DROPPED",` +
				`"IP":{"source":"1.1.1.1","destination":"2.2.2.2"},` +
				`"l4":{"TCP":{"source_port":31793,"destination_port":8080}},` +
				`"Type":"L3_L4","node_name":"k8s1",` +
				`"event_type":{"type":1,"sub_type":133},"Summary":"TCP Flags: SYN"}`,
		},
		{
			name: "jsonpb",
			options: []Option{
				JSONPB(),
				Writer(&buf),
			},
			args: args{
				f: &f,
			},
			wantErr: false,
			expected: `{"flow":{"time":"1970-01-01T00:20:34.567800Z",` +
				`"verdict":"DROPPED",` +
				`"IP":{"source":"1.1.1.1","destination":"2.2.2.2"},` +
				`"l4":{"TCP":{"source_port":31793,"destination_port":8080}},` +
				`"Type":"L3_L4","node_name":"k8s1",` +
				`"event_type":{"type":1,"sub_type":133},"Summary":"TCP Flags: SYN"}}`,
		},
		{
			name: "dict",
			options: []Option{
				Dict(),
				Writer(&buf),
			},
			args: args{
				f: &f,
			},
			wantErr: false,
			expected: `  TIMESTAMP: Jan  1 00:20:34.567
     SOURCE: 1.1.1.1:31793
DESTINATION: 2.2.2.2:8080
       TYPE: Policy denied
    VERDICT: DROPPED
    SUMMARY: TCP Flags: SYN`,
		},
		{
			name: "dict-with-node",
			options: []Option{
				Dict(),
				WithNodeName(),
				Writer(&buf),
			},
			args: args{
				f: &f,
			},
			wantErr: false,
			expected: `  TIMESTAMP: Jan  1 00:20:34.567
       NODE: k8s1
     SOURCE: 1.1.1.1:31793
DESTINATION: 2.2.2.2:8080
       TYPE: Policy denied
    VERDICT: DROPPED
    SUMMARY: TCP Flags: SYN`,
		},
	}
	for _, tt := range tests {
		buf.Reset()
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.options...)
			res := &observerpb.GetFlowsResponse{
				ResponseTypes: &observerpb.GetFlowsResponse_Flow{Flow: tt.args.f},
			}
			//writes a node status event into the error stream
			if err := p.WriteProtoFlow(res); (err != nil) != tt.wantErr {
				t.Errorf("WriteProtoFlow() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.NoError(t, p.Close())
			require.Equal(t, strings.TrimSpace(tt.expected), strings.TrimSpace(buf.String()))
		})
	}
}

func Test_getHostNames(t *testing.T) {
	type args struct {
		f *pb.Flow
	}
	type want struct {
		src, dst string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "nil flow",
			args: args{},
			want: want{},
		}, {
			name: "nil ip",
			args: args{
				f: &pb.Flow{},
			},
			want: want{},
		}, {
			name: "valid ips",
			args: args{
				f: &pb.Flow{
					IP: &pb.IP{
						Source:      "1.1.1.1",
						Destination: "2.2.2.2",
					},
				},
			},
			want: want{
				src: "1.1.1.1",
				dst: "2.2.2.2",
			},
		}, {
			name: "valid ips/endpoints",
			args: args{
				f: &pb.Flow{
					IP: &pb.IP{
						Source:      "1.1.1.1",
						Destination: "2.2.2.2",
					},
					Source: &pb.Endpoint{
						Namespace: "srcns",
						PodName:   "srcpod",
					},
					Destination: &pb.Endpoint{
						Namespace: "dstns",
						PodName:   "dstpod",
					},
				},
			},
			want: want{
				src: "srcns/srcpod",
				dst: "dstns/dstpod",
			},
		}, {
			name: "valid tcp",
			args: args{
				f: &pb.Flow{
					IP: &pb.IP{
						Source:      "1.1.1.1",
						Destination: "2.2.2.2",
					},
					L4: &pb.Layer4{
						Protocol: &pb.Layer4_TCP{
							TCP: &pb.TCP{
								SourcePort:      55555,
								DestinationPort: 80,
							},
						},
					},
				},
			},
			want: want{
				src: "1.1.1.1:55555",
				dst: "2.2.2.2:80",
			},
		}, {
			name: "valid udp",
			args: args{
				f: &pb.Flow{
					IP: &pb.IP{
						Source:      "1.1.1.1",
						Destination: "2.2.2.2",
					},
					L4: &pb.Layer4{
						Protocol: &pb.Layer4_UDP{
							UDP: &pb.UDP{
								SourcePort:      55555,
								DestinationPort: 53,
							},
						},
					},
				},
			},
			want: want{
				src: "1.1.1.1:55555",
				dst: "2.2.2.2:53",
			},
		}, {
			name: "valid tcp service",
			args: args{
				f: &pb.Flow{
					IP: &pb.IP{
						Source:      "1.1.1.1",
						Destination: "2.2.2.2",
					},
					L4: &pb.Layer4{
						Protocol: &pb.Layer4_TCP{
							TCP: &pb.TCP{
								SourcePort:      55555,
								DestinationPort: 80,
							},
						},
					},
					SourceService: &pb.Service{
						Name:      "xwing",
						Namespace: "default",
					},
					DestinationService: &pb.Service{
						Name:      "tiefighter",
						Namespace: "deathstar",
					},
				},
			},
			want: want{
				src: "default/xwing:55555",
				dst: "deathstar/tiefighter:80",
			},
		}, {
			name: "valid udp service",
			args: args{
				f: &pb.Flow{
					IP: &pb.IP{
						Source:      "1.1.1.1",
						Destination: "2.2.2.2",
					},
					L4: &pb.Layer4{
						Protocol: &pb.Layer4_UDP{
							UDP: &pb.UDP{
								SourcePort:      55555,
								DestinationPort: 53,
							},
						},
					},
					SourceService: &pb.Service{
						Name:      "xwing",
						Namespace: "default",
					},
					DestinationService: &pb.Service{
						Name:      "tiefighter",
						Namespace: "deathstar",
					},
				},
			},
			want: want{
				src: "default/xwing:55555",
				dst: "deathstar/tiefighter:53",
			},
		}, {
			name: "dns",
			args: args{
				f: &pb.Flow{
					IP: &pb.IP{
						Source:      "1.1.1.1",
						Destination: "2.2.2.2",
					},
					L4: &pb.Layer4{
						Protocol: &pb.Layer4_TCP{
							TCP: &pb.TCP{
								SourcePort:      54321,
								DestinationPort: 65432,
							},
						},
					},
					SourceNames:      []string{"a"},
					DestinationNames: []string{"b"},
				},
			},
			want: want{
				src: "a:54321",
				dst: "b:65432",
			},
		},
		{
			name: "ethernet",
			args: args{
				f: &pb.Flow{
					Ethernet: &pb.Ethernet{
						Source:      "00:01:02:03:04:05",
						Destination: "06:07:08:09:0a:0b",
					},
				},
			},
			want: want{
				src: "00:01:02:03:04:05",
				dst: "06:07:08:09:0a:0b",
			},
		},
	}
	p := New(WithIPTranslation())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSrc, gotDst := p.GetHostNames(tt.args.f)
			if gotSrc != tt.want.src {
				t.Errorf("GetHostNames() got = %v, want %v", gotSrc, tt.want.src)
			}
			if gotDst != tt.want.dst {
				t.Errorf("GetHostNames() got1 = %v, want %v", gotDst, tt.want.dst)
			}
		})
	}
}

func Test_fmtTimestamp(t *testing.T) {
	type args struct {
		t *timestamp.Timestamp
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid",
			args: args{
				t: &timestamp.Timestamp{
					Seconds: 0,
					Nanos:   0,
				},
			},
			want: "Jan  1 00:00:00.000",
		},
		{
			name: "invalid",
			args: args{
				t: &timestamp.Timestamp{
					Seconds: -1,
					Nanos:   -1,
				},
			},
			want: "N/A",
		},
		{
			name: "nil timestamp",
			args: args{},
			want: "N/A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmtTimestamp(tt.args.t); got != tt.want {
				t.Errorf("getTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getFlowType(t *testing.T) {
	type args struct {
		f *pb.Flow
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "L7",
			args: args{
				f: &pb.Flow{
					L7: &pb.Layer7{
						Type: pb.L7FlowType_REQUEST,
					},
					EventType: &pb.CiliumEventType{
						Type: monitorAPI.MessageTypeAccessLog,
					},
				},
			},
			want: "l7-request",
		},
		{
			name: "HTTP",
			args: args{
				f: &pb.Flow{
					L7: &pb.Layer7{
						Type:   pb.L7FlowType_RESPONSE,
						Record: &pb.Layer7_Http{},
					},
					EventType: &pb.CiliumEventType{
						Type: monitorAPI.MessageTypeAccessLog,
					},
				},
			},
			want: "http-response",
		},
		{
			name: "Kafka",
			args: args{
				f: &pb.Flow{
					L7: &pb.Layer7{
						Type:   pb.L7FlowType_REQUEST,
						Record: &pb.Layer7_Kafka{},
					},
					EventType: &pb.CiliumEventType{
						Type: monitorAPI.MessageTypeAccessLog,
					},
				},
			},
			want: "kafka-request",
		},
		{
			name: "DNS",
			args: args{
				f: &pb.Flow{
					L7: &pb.Layer7{
						Type:   pb.L7FlowType_REQUEST,
						Record: &pb.Layer7_Dns{},
					},
					EventType: &pb.CiliumEventType{
						Type: monitorAPI.MessageTypeAccessLog,
					},
				},
			},
			want: "dns-request",
		},
		{
			name: "L4",
			args: args{
				f: &pb.Flow{
					EventType: &pb.CiliumEventType{
						Type:    monitorAPI.MessageTypeTrace,
						SubType: monitorAPI.TraceToHost,
					},
				},
			},
			want: "to-host",
		},
		{
			name: "L4",
			args: args{
				f: &pb.Flow{
					Verdict: pb.Verdict_FORWARDED,
					EventType: &pb.CiliumEventType{
						Type: monitorAPI.MessageTypePolicyVerdict,
					},
					PolicyMatchType: monitorAPI.PolicyMatchL3L4,
				},
			},
			want: "L3-L4",
		},
		{
			name: "L4",
			args: args{
				f: &pb.Flow{
					Verdict: pb.Verdict_DROPPED,
					EventType: &pb.CiliumEventType{
						Type: monitorAPI.MessageTypePolicyVerdict,
					},
					DropReason: 153,
				},
			},
			want: "Error while correcting L3 checksum",
		},
		{
			name: "invalid",
			args: args{
				f: &pb.Flow{
					EventType: &pb.CiliumEventType{
						Type:    monitorAPI.MessageTypeTrace,
						SubType: 123, // invalid subtype
					},
				},
			},
			want: "123",
		},
		{
			name: "nil flow",
			args: args{},
			want: "UNKNOWN",
		},

		{
			name: "nil type",
			args: args{
				f: &pb.Flow{},
			},
			want: "UNKNOWN",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFlowType(tt.args.f); got != tt.want {
				t.Errorf("GetFlowType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaybeTime(t *testing.T) {
	assert.Equal(t, "N/A", MaybeTime(nil))

	mt := time.Date(2018, time.July, 07, 17, 30, 0, 123000000, time.UTC)
	assert.Equal(t, "Jul  7 17:30:00.123", MaybeTime(&mt))
}

func TestHostname(t *testing.T) {
	p := New(WithIPTranslation())
	assert.Equal(t, "default/pod", p.Hostname("", "", "default", "pod", "", []string{}))
	assert.Equal(t, "default/pod", p.Hostname("", "", "default", "pod", "service", []string{}))
	assert.Equal(t, "default/service", p.Hostname("", "", "default", "", "service", []string{}))
	assert.Equal(t, "a,b", p.Hostname("", "", "", "", "", []string{"a", "b"}))
	p = New()
	assert.Equal(t, "1.1.1.1:80", p.Hostname("1.1.1.1", "80", "default", "pod", "", []string{}))
	assert.Equal(t, "1.1.1.1:80", p.Hostname("1.1.1.1", "80", "default", "pod", "service", []string{}))
	assert.Equal(t, "1.1.1.1", p.Hostname("1.1.1.1", "0", "default", "pod", "", []string{}))
	assert.Equal(t, "1.1.1.1", p.Hostname("1.1.1.1", "0", "default", "pod", "service", []string{}))
}

func TestPrinter_AgentEventDetails(t *testing.T) {
	startTS, err := ptypes.TimestampProto(time.Now())
	assert.NoError(t, err)

	tests := []struct {
		name string
		ev   *pb.AgentEvent
		want string
	}{
		{
			name: "nil",
			want: "UNKNOWN",
		},
		{
			name: "empty",
			ev:   &pb.AgentEvent{},
			want: "UNKNOWN",
		},
		{
			name: "unknown without notification",
			ev: &pb.AgentEvent{
				Type: pb.AgentEventType_AGENT_EVENT_UNKNOWN,
			},
			want: "UNKNOWN",
		},
		{
			name: "agent start without notification",
			ev: &pb.AgentEvent{
				Type: pb.AgentEventType_AGENT_STARTED,
			},
			want: "UNKNOWN",
		},
		{
			name: "agent start with notification",
			ev: &pb.AgentEvent{
				Type: pb.AgentEventType_AGENT_STARTED,
				Notification: &pb.AgentEvent_AgentStart{
					AgentStart: &pb.TimeNotification{
						Time: startTS,
					},
				},
			},
			want: "start time: " + fmtTimestamp(startTS),
		},
		{
			name: "policy update",
			ev: &pb.AgentEvent{
				Type: pb.AgentEventType_POLICY_UPDATED,
				Notification: &pb.AgentEvent_PolicyUpdate{
					PolicyUpdate: &pb.PolicyUpdateNotification{
						Labels:    []string{"foo=bar", "baz=foo"},
						Revision:  1,
						RuleCount: 2,
					},
				},
			},
			want: "labels: [foo=bar,baz=foo], revision: 1, count: 2",
		},
		{
			name: "policy delete",
			ev: &pb.AgentEvent{
				Type: pb.AgentEventType_POLICY_DELETED,
				Notification: &pb.AgentEvent_PolicyUpdate{
					PolicyUpdate: &pb.PolicyUpdateNotification{
						Revision:  42,
						RuleCount: 1,
					},
				},
			},
			want: "labels: [], revision: 42, count: 1",
		},
		{
			name: "endpoint regenerate success",
			ev: &pb.AgentEvent{
				Type: pb.AgentEventType_ENDPOINT_REGENERATE_SUCCESS,
				Notification: &pb.AgentEvent_EndpointRegenerate{
					EndpointRegenerate: &pb.EndpointRegenNotification{
						Id:     42,
						Labels: []string{"baz=bar", "some=label"},
					},
				},
			},
			want: "id: 42, labels: [baz=bar,some=label]",
		},
		{
			name: "endpoint regenerate failure",
			ev: &pb.AgentEvent{
				Type: pb.AgentEventType_ENDPOINT_REGENERATE_FAILURE,
				Notification: &pb.AgentEvent_EndpointRegenerate{
					EndpointRegenerate: &pb.EndpointRegenNotification{
						Id:     42,
						Labels: []string{"baz=bar", "some=label"},
						Error:  "some error",
					},
				},
			},
			want: "id: 42, labels: [baz=bar,some=label], error: some error",
		},
		{
			name: "endpoint created",
			ev: &pb.AgentEvent{
				Type: pb.AgentEventType_ENDPOINT_CREATED,
				Notification: &pb.AgentEvent_EndpointUpdate{
					EndpointUpdate: &pb.EndpointUpdateNotification{
						Id:        1027,
						PodName:   "cilium-xyz",
						Namespace: "kube-system",
					},
				},
			},
			want: "id: 1027, pod name: cilium-xyz, namespace: kube-system",
		},
		{
			name: "ipcache upsert",
			ev: &pb.AgentEvent{
				Type: pb.AgentEventType_IPCACHE_UPSERTED,
				Notification: &pb.AgentEvent_IpcacheUpdate{
					IpcacheUpdate: &pb.IPCacheNotification{
						Cidr:     "10.1.2.3/32",
						Identity: 42,
						OldIdentity: &wrappers.UInt32Value{
							Value: 23,
						},
						HostIp:     "192.168.3.9",
						EncryptKey: 3,
					},
				},
			},
			want: "cidr: 10.1.2.3/32, identity: 42, old identity: 23, host ip: 192.168.3.9, encrypt key: 3",
		},
		{
			name: "ipcache delete",
			ev: &pb.AgentEvent{
				Type: pb.AgentEventType_IPCACHE_DELETED,
				Notification: &pb.AgentEvent_IpcacheUpdate{
					IpcacheUpdate: &pb.IPCacheNotification{
						Cidr:      "10.0.1.2/32",
						Identity:  42,
						OldHostIp: "192.168.1.23",
					},
				},
			},
			want: "cidr: 10.0.1.2/32, identity: 42, old host ip: 192.168.1.23, encrypt key: 0",
		},
		{
			name: "service upsert",
			ev: &pb.AgentEvent{
				Type: pb.AgentEventType_SERVICE_UPSERTED,
				Notification: &pb.AgentEvent_ServiceUpsert{
					ServiceUpsert: &pb.ServiceUpsertNotification{
						Id: 42,
						FrontendAddress: &pb.ServiceUpsertNotificationAddr{
							Ip:   "10.0.0.42",
							Port: 8008,
						},
						BackendAddresses: []*pb.ServiceUpsertNotificationAddr{
							{
								Ip:   "192.168.1.23",
								Port: 80,
							},
							{
								Ip:   "2001:db8:85a3:::8a2e:370:1337",
								Port: 8080,
							},
						},
						Type:          "foobar",
						TrafficPolicy: "pol1",
						Name:          "foo",
						Namespace:     "bar",
					},
				},
			},
			want: "id: 42, frontend: 10.0.0.42:8008, backends: [192.168.1.23:80,[2001:db8:85a3:::8a2e:370:1337]:8080], type: foobar, traffic policy: pol1, name: foo, namespace: bar",
		},
		{
			name: "service delete",
			ev: &pb.AgentEvent{
				Type: pb.AgentEventType_SERVICE_DELETED,
				Notification: &pb.AgentEvent_ServiceDelete{
					ServiceDelete: &pb.ServiceDeleteNotification{
						Id: 42,
					},
				},
			},
			want: "id: 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAgentEventDetails(tt.ev); got != tt.want {
				t.Errorf("getAgentEventDetails()\ngot:  %v,\nwant: %v", got, tt.want)
			}
		})
	}

}
