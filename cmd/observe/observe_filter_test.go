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
	"strconv"
	"testing"

	pb "github.com/cilium/cilium/api/v1/flow"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoBlacklist(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{
		"--from-ip", "1.2.3.4",
	}))
	assert.Nil(t, f.blacklist, "blacklist should be nil")
}

// The default filter should be empty.
func TestDefaultFilter(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{}))
	assert.Nil(t, f.whitelist)
	assert.Nil(t, f.blacklist)
}

func TestConflicts(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	err := cmd.Flags().Parse([]string{
		"--from-ip", "1.2.3.4",
		"--from-fqdn", "doesnt.work",
	})
	require.Error(t, err)

	assert.Contains(t,
		err.Error(),
		"filters --from-fqdn and --from-ip cannot be combined",
	)
}

func TestTrailingNot(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	err := cmd.Flags().Parse([]string{
		"--from-ip", "1.2.3.4",
		"--not",
	})
	require.NoError(t, err)

	err = handleArgs(f, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "trailing --not")
}

func TestFilterDispatch(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{
		"--from-ip", "1.2.3.4",
		"--from-ip", "5.6.7.8",
		"--not",
		"--to-ip", "5.5.5.5",
		"--verdict", "DROPPED",
		"-t", "l7", // int:129 in cilium-land
	}))

	require.NoError(t, handleArgs(f, false))
	if diff := cmp.Diff(
		[]*pb.FlowFilter{
			{
				SourceIp: []string{"1.2.3.4", "5.6.7.8"},
				Verdict:  []pb.Verdict{pb.Verdict_DROPPED},
				EventType: []*pb.EventTypeFilter{
					{Type: monitorAPI.MessageTypeAccessLog},
				},
			},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(pb.FlowFilter{}),
		cmpopts.IgnoreUnexported(pb.EventTypeFilter{}),
	); diff != "" {
		t.Errorf("whitelist filter mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(
		[]*pb.FlowFilter{
			{
				DestinationIp: []string{"5.5.5.5"},
			},
		},
		f.blacklist.flowFilters(),
		cmpopts.IgnoreUnexported(pb.FlowFilter{}),
	); diff != "" {
		t.Errorf("blacklist filter mismatch (-want +got):\n%s", diff)
	}
}

func TestFilterLeftRight(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{
		"--ip", "1.2.3.4",
		"--ip", "5.6.7.8",
		"--verdict", "DROPPED",
		"--not", "--pod", "deathstar",
		"--not", "--http-status", "200",
		"--http-method", "get",
		"--http-path", "/page/\\d+",
		"--node-name", "k8s*",
	}))

	require.NoError(t, handleArgs(f, false))

	if diff := cmp.Diff(
		[]*pb.FlowFilter{
			{
				SourceIp:   []string{"1.2.3.4", "5.6.7.8"},
				Verdict:    []pb.Verdict{pb.Verdict_DROPPED},
				HttpMethod: []string{"get"},
				HttpPath:   []string{"/page/\\d+"},
				NodeName:   []string{"k8s*"},
			},
			{
				DestinationIp: []string{"1.2.3.4", "5.6.7.8"},
				Verdict:       []pb.Verdict{pb.Verdict_DROPPED},
				HttpMethod:    []string{"get"},
				HttpPath:      []string{"/page/\\d+"},
				NodeName:      []string{"k8s*"},
			},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(pb.FlowFilter{}),
	); diff != "" {
		t.Errorf("whitelist filter mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(
		[]*pb.FlowFilter{
			{
				SourcePod:      []string{"deathstar"},
				HttpStatusCode: []string{"200"},
			},
			{
				DestinationPod: []string{"deathstar"},
				HttpStatusCode: []string{"200"},
			},
		},
		f.blacklist.flowFilters(),
		cmpopts.IgnoreUnexported(pb.FlowFilter{}),
	); diff != "" {
		t.Errorf("blacklist filter mismatch (-want +got):\n%s", diff)
	}
}

func TestAgentEventSubTypeMap(t *testing.T) {
	// Make sure to keep agent event sub-types maps in sync. See agentEventSubtypes godoc for
	// details.
	require.Len(t, agentEventSubtypes, len(monitorAPI.AgentNotifications))
	for _, v := range agentEventSubtypes {
		require.Contains(t, monitorAPI.AgentNotifications, v)
	}
	agentEventSubtypesContainsValue := func(an monitorAPI.AgentNotification) bool {
		for _, v := range agentEventSubtypes {
			if v == an {
				return true
			}
		}
		return false
	}
	for k := range monitorAPI.AgentNotifications {
		require.True(t, agentEventSubtypesContainsValue(k))
	}
}

func TestFilterType(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	require.Error(t, cmd.Flags().Parse([]string{
		"-t", "some-invalid-type",
	}))

	require.Error(t, cmd.Flags().Parse([]string{
		"-t", "trace:some-invalid-sub-type",
	}))

	require.Error(t, cmd.Flags().Parse([]string{
		"-t", "agent:Policy updated",
	}))

	require.NoError(t, cmd.Flags().Parse([]string{
		"-t", "254",
		"-t", "255:127",
		"-t", "trace:to-endpoint",
		"-t", "trace:from-endpoint",
		"-t", strconv.Itoa(monitorAPI.MessageTypeTrace) + ":" + strconv.Itoa(monitorAPI.TraceToHost),
		"-t", "agent",
		"-t", "agent:3",
		"-t", "agent:policy-updated",
		"-t", "agent:service-deleted",
	}))

	require.NoError(t, handleArgs(f, false))
	if diff := cmp.Diff(
		[]*pb.FlowFilter{
			{
				EventType: []*pb.EventTypeFilter{
					{
						Type: 254,
					},

					{
						Type:         255,
						MatchSubType: true,
						SubType:      127,
					},
					{
						Type:         monitorAPI.MessageTypeTrace,
						MatchSubType: true,
						SubType:      monitorAPI.TraceToLxc,
					},
					{
						Type:         monitorAPI.MessageTypeTrace,
						MatchSubType: true,
						SubType:      monitorAPI.TraceFromLxc,
					},
					{
						Type:         monitorAPI.MessageTypeTrace,
						MatchSubType: true,
						SubType:      monitorAPI.TraceToHost,
					},
					{
						Type: monitorAPI.MessageTypeAgent,
					},
					{
						Type:         monitorAPI.MessageTypeAgent,
						MatchSubType: true,
						SubType:      3,
					},
					{
						Type:         monitorAPI.MessageTypeAgent,
						MatchSubType: true,
						SubType:      int32(monitorAPI.AgentNotifyPolicyUpdated),
					},
					{
						Type:         monitorAPI.MessageTypeAgent,
						MatchSubType: true,
						SubType:      int32(monitorAPI.AgentNotifyServiceDeleted),
					},
				},
			},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(pb.FlowFilter{}),
		cmpopts.IgnoreUnexported(pb.EventTypeFilter{}),
	); diff != "" {
		t.Errorf("filter mismatch (-want +got):\n%s", diff)
	}
}

func TestLabels(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	err := cmd.Flags().Parse([]string{
		"--label", "k1=v1,k2=v2",
		"-l", "k3",
	})
	require.NoError(t, err)
	if diff := cmp.Diff(
		[]*pb.FlowFilter{
			{SourceLabel: []string{"k1=v1,k2=v2", "k3"}},
			{DestinationLabel: []string{"k1=v1,k2=v2", "k3"}},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(pb.FlowFilter{}),
	); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	assert.Nil(t, f.blacklist)
}

func TestIdentity(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{"--identity", "1", "--identity", "2"}))
	if diff := cmp.Diff(
		[]*pb.FlowFilter{
			{SourceIdentity: []uint32{1, 2}},
			{DestinationIdentity: []uint32{1, 2}},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(pb.FlowFilter{}),
	); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	assert.Nil(t, f.blacklist)
}

func TestFromIdentity(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{"--from-identity", "1", "--from-identity", "2"}))
	if diff := cmp.Diff(
		[]*pb.FlowFilter{
			{SourceIdentity: []uint32{1, 2}},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(pb.FlowFilter{}),
	); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	assert.Nil(t, f.blacklist)
}

func TestToIdentity(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{"--to-identity", "1", "--to-identity", "2"}))
	if diff := cmp.Diff(
		[]*pb.FlowFilter{
			{DestinationIdentity: []uint32{1, 2}},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(pb.FlowFilter{}),
	); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	assert.Nil(t, f.blacklist)
}

func TestInvalidIdentity(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	require.Error(t, cmd.Flags().Parse([]string{"--from-identity", "bad"}))
	require.Error(t, cmd.Flags().Parse([]string{"--to-identity", "bad"}))
	require.Error(t, cmd.Flags().Parse([]string{"--identity", "bad"}))
}

func TestTcpFlags(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(viper.New(), f)

	// valid TCP flags
	validflags := []string{"SYN", "syn", "FIN", "RST", "PSH", "ACK", "URG", "ECE", "CWR", "NS", "syn,ack"}
	for _, f := range validflags {
		require.NoError(t, cmd.Flags().Parse([]string{"--tcp-flags", f}))                               // single --tcp-flags
		require.NoError(t, cmd.Flags().Parse([]string{"--tcp-flags", f, "--tcp-flags", "syn"}))         // multiple --tcp-flags
		require.NoError(t, cmd.Flags().Parse([]string{"--tcp-flags", f, "--not", "--tcp-flags", "NS"})) // --not --tcp-flags
	}

	// invalid TCP flags
	invalidflags := []string{"unknown", "syn,unknown", "unknown,syn", "syn,", ",syn"}
	for _, f := range invalidflags {
		require.Error(t, cmd.Flags().Parse([]string{"--tcp-flags", f}))
	}
}
