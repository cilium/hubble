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

	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"

	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	pb "github.com/cilium/hubble/api/v1/observer"
)

func TestPrinter_WriteProtoFlow(t *testing.T) {
	buf := bytes.Buffer{}
	f := pb.Flow{
		Time: &types.Timestamp{
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
			expected: `TIMESTAMP             SOURCE          DESTINATION              TYPE                 VERDICT   SUMMARY
Jan  1 00:20:34.567   1.1.1.1:31793   2.2.2.2:8080(http-alt)   Policy denied (L3)   DROPPED   TCP Flags: SYN`,
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
			expected: "Jan  1 00:20:34.567 " +
				"[k8s1]: 1.1.1.1:31793 -> 2.2.2.2:8080(http-alt) " +
				"Policy denied (L3) DROPPED (TCP Flags: SYN)\n",
		},
		{
			name: "json",
			options: []Option{
				JSON(),
				WithJSONEncoder(),
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
				`"event_type":{"sub_type":133},"Summary":"TCP Flags: SYN"}`,
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
DESTINATION: 2.2.2.2:8080(http-alt)
       TYPE: Policy denied (L3)
    VERDICT: DROPPED
    SUMMARY: TCP Flags: SYN`,
		},
	}
	for _, tt := range tests {
		buf.Reset()
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.options...)
			if err := p.WriteProtoFlow(tt.args.f); (err != nil) != tt.wantErr {
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
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name:  "nil flow",
			args:  args{},
			want:  "",
			want1: "",
		},
		{
			name: "nil ip",
			args: args{
				f: &pb.Flow{},
			},
			want:  "",
			want1: "",
		},
		{
			name: "valid ips",
			args: args{
				f: &pb.Flow{
					IP: &pb.IP{
						Source:      "1.1.1.1",
						Destination: "2.2.2.2",
					},
				},
			},
			want:  "1.1.1.1",
			want1: "2.2.2.2",
		},
		{
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
			want:  "srcns/srcpod",
			want1: "dstns/dstpod",
		},
		{
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
			want:  "1.1.1.1:55555",
			want1: "2.2.2.2:80(http)",
		},
		{
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
			want:  "1.1.1.1:55555",
			want1: "2.2.2.2:53(domain)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getHostNames(tt.args.f)
			if got != tt.want {
				t.Errorf("getHostNames() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getHostNames() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getTimestamp(t *testing.T) {
	type args struct {
		f *pb.Flow
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid",
			args: args{
				f: &pb.Flow{
					Time: &types.Timestamp{
						Seconds: 0,
						Nanos:   0,
					},
				},
			},
			want: "Jan  1 00:00:00.000",
		},
		{
			name: "invalid",
			args: args{},
			want: "N/A",
		},
		{
			name: "nil flow",
			args: args{},
			want: "N/A",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTimestamp(tt.args.f); got != tt.want {
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
			if got := getFlowType(tt.args.f); got != tt.want {
				t.Errorf("getFlowType() = %v, want %v", got, tt.want)
			}
		})
	}
}
