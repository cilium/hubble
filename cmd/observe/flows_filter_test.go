// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

package observe

import (
	"os"
	"strconv"
	"testing"

	flowpb "github.com/cilium/cilium/api/v1/flow"
	"github.com/cilium/cilium/pkg/identity"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoBlacklist(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{
		"--from-ip", "1.2.3.4",
	}))
	assert.Nil(t, f.blacklist, "blacklist should be nil")
}

// The default filter should be empty.
func TestDefaultFilter(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{}))
	assert.Nil(t, f.whitelist)
	assert.Nil(t, f.blacklist)
}

func TestConflicts(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

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
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

	err := cmd.Flags().Parse([]string{
		"--from-ip", "1.2.3.4",
		"--not",
	})
	require.NoError(t, err)

	err = handleFlowArgs(os.Stdout, f, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "trailing --not")
}

func TestFilterDispatch(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{
		"--from-ip", "1.2.3.4",
		"--from-ip", "5.6.7.8",
		"--not",
		"--to-ip", "5.5.5.5",
		"--verdict", "DROPPED",
		"-t", "l7", // int:129 in cilium-land
	}))

	require.NoError(t, handleFlowArgs(os.Stdout, f, false))
	if diff := cmp.Diff(
		[]*flowpb.FlowFilter{
			{
				SourceIp: []string{"1.2.3.4", "5.6.7.8"},
				Verdict:  []flowpb.Verdict{flowpb.Verdict_DROPPED},
				EventType: []*flowpb.EventTypeFilter{
					{Type: monitorAPI.MessageTypeAccessLog},
				},
			},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
		cmpopts.IgnoreUnexported(flowpb.EventTypeFilter{}),
	); diff != "" {
		t.Errorf("whitelist filter mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(
		[]*flowpb.FlowFilter{
			{
				DestinationIp: []string{"5.5.5.5"},
			},
		},
		f.blacklist.flowFilters(),
		cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
	); diff != "" {
		t.Errorf("blacklist filter mismatch (-want +got):\n%s", diff)
	}
}

func TestFilterLeftRight(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

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

	require.NoError(t, handleFlowArgs(os.Stdout, f, false))

	if diff := cmp.Diff(
		[]*flowpb.FlowFilter{
			{
				SourceIp:   []string{"1.2.3.4", "5.6.7.8"},
				Verdict:    []flowpb.Verdict{flowpb.Verdict_DROPPED},
				HttpMethod: []string{"get"},
				HttpPath:   []string{"/page/\\d+"},
				NodeName:   []string{"k8s*"},
			},
			{
				DestinationIp: []string{"1.2.3.4", "5.6.7.8"},
				Verdict:       []flowpb.Verdict{flowpb.Verdict_DROPPED},
				HttpMethod:    []string{"get"},
				HttpPath:      []string{"/page/\\d+"},
				NodeName:      []string{"k8s*"},
			},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
	); diff != "" {
		t.Errorf("whitelist filter mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(
		[]*flowpb.FlowFilter{
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
		cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
	); diff != "" {
		t.Errorf("blacklist filter mismatch (-want +got):\n%s", diff)
	}
}

func TestFilterType(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

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

	require.NoError(t, handleFlowArgs(os.Stdout, f, false))
	if diff := cmp.Diff(
		[]*flowpb.FlowFilter{
			{
				EventType: []*flowpb.EventTypeFilter{
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
		cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
		cmpopts.IgnoreUnexported(flowpb.EventTypeFilter{}),
	); diff != "" {
		t.Errorf("filter mismatch (-want +got):\n%s", diff)
	}
}

func TestLabels(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

	err := cmd.Flags().Parse([]string{
		"--label", "k1=v1,k2=v2",
		"-l", "k3",
	})
	require.NoError(t, err)
	if diff := cmp.Diff(
		[]*flowpb.FlowFilter{
			{SourceLabel: []string{"k1=v1,k2=v2", "k3"}},
			{DestinationLabel: []string{"k1=v1,k2=v2", "k3"}},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
	); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	assert.Nil(t, f.blacklist)
}

func TestFromToWorkloadCombined(t *testing.T) {
	t.Run("single filter", func(t *testing.T) {
		f := newFlowFilter()
		cmd := newFlowsCmdWithFilter(viper.New(), f)

		require.NoError(t, cmd.Flags().Parse([]string{"--from-pod", "cilium", "--to-workload", "app"}))
		if diff := cmp.Diff(
			[]*flowpb.FlowFilter{
				{
					SourcePod:           []string{"cilium"},
					DestinationWorkload: []*flowpb.Workload{{Name: "app"}},
				},
			},
			f.whitelist.flowFilters(),
			cmpopts.IgnoreUnexported(flowpb.FlowFilter{}, flowpb.Workload{}),
		); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
		assert.Nil(t, f.blacklist)
	})

	t.Run("two filters", func(t *testing.T) {
		f := newFlowFilter()
		cmd := newFlowsCmdWithFilter(viper.New(), f)

		require.NoError(t, cmd.Flags().Parse([]string{"--pod", "cilium", "--to-workload", "app"}))
		if diff := cmp.Diff(
			[]*flowpb.FlowFilter{
				{
					SourcePod:           []string{"cilium"},
					DestinationWorkload: []*flowpb.Workload{{Name: "app"}},
				},
				{
					DestinationPod:      []string{"cilium"},
					DestinationWorkload: []*flowpb.Workload{{Name: "app"}},
				},
			},
			f.whitelist.flowFilters(),
			cmpopts.IgnoreUnexported(flowpb.FlowFilter{}, flowpb.Workload{}),
		); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
		assert.Nil(t, f.blacklist)
	})
}

func TestIdentity(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{"--identity", "1", "--identity", "2"}))
	if diff := cmp.Diff(
		[]*flowpb.FlowFilter{
			{SourceIdentity: []uint32{1, 2}},
			{DestinationIdentity: []uint32{1, 2}},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
	); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	assert.Nil(t, f.blacklist)

	// reserved identities
	for _, id := range identity.GetAllReservedIdentities() {
		t.Run(id.String(), func(t *testing.T) {
			f := newFlowFilter()
			cmd := newFlowsCmdWithFilter(viper.New(), f)
			require.NoError(t, cmd.Flags().Parse([]string{"--identity", id.String()}))
			if diff := cmp.Diff(
				[]*flowpb.FlowFilter{
					{SourceIdentity: []uint32{id.Uint32()}},
					{DestinationIdentity: []uint32{id.Uint32()}},
				},
				f.whitelist.flowFilters(),
				cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
			); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			assert.Nil(t, f.blacklist)
		})
	}
}

func TestFromIdentity(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{"--from-identity", "1", "--from-identity", "2"}))
	if diff := cmp.Diff(
		[]*flowpb.FlowFilter{
			{SourceIdentity: []uint32{1, 2}},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
	); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	assert.Nil(t, f.blacklist)

	// reserved identities
	for _, id := range identity.GetAllReservedIdentities() {
		t.Run(id.String(), func(t *testing.T) {
			f := newFlowFilter()
			cmd := newFlowsCmdWithFilter(viper.New(), f)
			require.NoError(t, cmd.Flags().Parse([]string{"--from-identity", id.String()}))
			if diff := cmp.Diff(
				[]*flowpb.FlowFilter{
					{SourceIdentity: []uint32{id.Uint32()}},
				},
				f.whitelist.flowFilters(),
				cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
			); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			assert.Nil(t, f.blacklist)
		})
	}
}

func TestToIdentity(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{"--to-identity", "1", "--to-identity", "2"}))
	if diff := cmp.Diff(
		[]*flowpb.FlowFilter{
			{DestinationIdentity: []uint32{1, 2}},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
	); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	assert.Nil(t, f.blacklist)

	// reserved identities
	for _, id := range identity.GetAllReservedIdentities() {
		t.Run(id.String(), func(t *testing.T) {
			f := newFlowFilter()
			cmd := newFlowsCmdWithFilter(viper.New(), f)
			require.NoError(t, cmd.Flags().Parse([]string{"--to-identity", id.String()}))
			if diff := cmp.Diff(
				[]*flowpb.FlowFilter{
					{DestinationIdentity: []uint32{id.Uint32()}},
				},
				f.whitelist.flowFilters(),
				cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
			); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			assert.Nil(t, f.blacklist)
		})
	}
}

func TestInvalidIdentity(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

	require.Error(t, cmd.Flags().Parse([]string{"--from-identity", "bad"}))
	require.Error(t, cmd.Flags().Parse([]string{"--to-identity", "bad"}))
	require.Error(t, cmd.Flags().Parse([]string{"--identity", "bad"}))
}

func TestTcpFlags(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

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

func TestUuid(t *testing.T) {
	f := newFlowFilter()
	cmd := newFlowsCmdWithFilter(viper.New(), f)

	require.NoError(t, cmd.Flags().Parse([]string{"--uuid", "b9fab269-04ae-495c-9d12-b6c36d41de0d"}))
	if diff := cmp.Diff(
		[]*flowpb.FlowFilter{
			{Uuid: []string{"b9fab269-04ae-495c-9d12-b6c36d41de0d"}},
		},
		f.whitelist.flowFilters(),
		cmpopts.IgnoreUnexported(flowpb.FlowFilter{}),
	); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	assert.Nil(t, f.blacklist)
}

func TestTrafficDirection(t *testing.T) {
	tt := []struct {
		name    string
		flags   []string
		filters []*flowpb.FlowFilter
		err     string
	}{
		{
			name:  "ingress",
			flags: []string{"--traffic-direction", "ingress"},
			filters: []*flowpb.FlowFilter{
				{TrafficDirection: []flowpb.TrafficDirection{flowpb.TrafficDirection_INGRESS}},
			},
		},
		{
			name:  "egress",
			flags: []string{"--traffic-direction", "egress"},
			filters: []*flowpb.FlowFilter{
				{TrafficDirection: []flowpb.TrafficDirection{flowpb.TrafficDirection_EGRESS}},
			},
		},
		{
			name:  "mixed case",
			flags: []string{"--traffic-direction", "INGRESS", "--traffic-direction", "EgrEss"},
			filters: []*flowpb.FlowFilter{
				{
					TrafficDirection: []flowpb.TrafficDirection{
						flowpb.TrafficDirection_INGRESS,
						flowpb.TrafficDirection_EGRESS,
					},
				},
			},
		},
		{
			name:  "invalid",
			flags: []string{"--traffic-direction", "to the moon"},
			err:   "to the moon: invalid traffic direction, expected ingress or egress",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := newFlowFilter()
			cmd := newFlowsCmdWithFilter(viper.New(), f)
			err := cmd.Flags().Parse(tc.flags)
			diff := cmp.Diff(tc.filters, f.whitelist.flowFilters(), cmpopts.IgnoreUnexported(flowpb.FlowFilter{}))
			if diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
			if tc.err != "" {
				assert.Errorf(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
			assert.Nil(t, f.blacklist)
		})
	}
}
