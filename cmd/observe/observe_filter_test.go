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
	"testing"

	pb "github.com/cilium/cilium/api/v1/flow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoBlacklist(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(f)

	require.NoError(t, cmd.Flags().Parse([]string{
		"--from-ip", "1.2.3.4",
	}))
	assert.Nil(t, f.blacklist, "blacklist should be nil")
}

// The default filter should be empty.
func TestDefaultFilter(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(f)
	require.NoError(t, cmd.Flags().Parse([]string{}))
	assert.Nil(t, f.whitelist)
	assert.Nil(t, f.blacklist)
}

func TestConflicts(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(f)

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
	cmd := newObserveCmd(f)

	err := cmd.Flags().Parse([]string{
		"--from-ip", "1.2.3.4",
		"--not",
	})
	require.NoError(t, err)

	err = handleArgs(f)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "trailing --not")
}

func TestFilterDispatch(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(f)

	require.NoError(t, cmd.Flags().Parse([]string{
		"--from-ip", "1.2.3.4",
		"--from-ip", "5.6.7.8",
		"--not",
		"--to-ip", "5.5.5.5",
		"--verdict", "DROPPED",
		"-t", "l7", // int:129 in cilium-land
	}))

	require.NoError(t, handleArgs(f))

	assert.Equal(t, []*pb.FlowFilter{
		{
			SourceIp:  []string{"1.2.3.4", "5.6.7.8"},
			Verdict:   []pb.Verdict{pb.Verdict_DROPPED},
			EventType: []*pb.EventTypeFilter{{Type: 129}},
		},
	}, f.whitelist.flowFilters(), "whitelist filter is incorrect")

	assert.Equal(t, []*pb.FlowFilter{
		{
			DestinationIp: []string{"5.5.5.5"},
		},
	}, f.blacklist.flowFilters(), "blacklist filter is incorrect")
}

func TestFilterLeftRight(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(f)

	require.NoError(t, cmd.Flags().Parse([]string{
		"--ip", "1.2.3.4",
		"--ip", "5.6.7.8",
		"--verdict", "DROPPED",
		"--not",
		"--pod", "deathstar",
		"--not",
		"--http-status", "200",
	}))

	require.NoError(t, handleArgs(f))

	assert.Equal(t, []*pb.FlowFilter{
		{
			SourceIp: []string{"1.2.3.4", "5.6.7.8"},
			Verdict:  []pb.Verdict{pb.Verdict_DROPPED},
		},
		{
			DestinationIp: []string{"1.2.3.4", "5.6.7.8"},
			Verdict:       []pb.Verdict{pb.Verdict_DROPPED},
		},
	}, f.whitelist.flowFilters(), "whitelist filter is incorrect")

	assert.Equal(t, []*pb.FlowFilter{
		{
			SourcePod:      []string{"deathstar"},
			HttpStatusCode: []string{"200"},
		},
		{
			DestinationPod: []string{"deathstar"},
			HttpStatusCode: []string{"200"},
		},
	}, f.blacklist.flowFilters(), "blacklist filter is incorrect")
}

func TestLabels(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(f)

	err := cmd.Flags().Parse([]string{
		"--label", "k1=v1,k2=v2",
		"-l", "k3",
	})
	require.NoError(t, err)
	assert.Equal(t, []*pb.FlowFilter{
		{SourceLabel: []string{"k1=v1,k2=v2", "k3"}},
		{DestinationLabel: []string{"k1=v1,k2=v2", "k3"}},
	}, f.whitelist.flowFilters())
	assert.Nil(t, f.blacklist)
}

func TestIdentity(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(f)
	require.NoError(t, cmd.Flags().Parse([]string{"--identity", "1", "--identity", "2"}))
	assert.Equal(t, []*pb.FlowFilter{
		{SourceIdentity: []uint32{1, 2}},
		{DestinationIdentity: []uint32{1, 2}},
	}, f.whitelist.flowFilters())
	assert.Nil(t, f.blacklist)
}

func TestFromIdentity(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(f)
	require.NoError(t, cmd.Flags().Parse([]string{"--from-identity", "1", "--from-identity", "2"}))
	assert.Equal(t, []*pb.FlowFilter{{SourceIdentity: []uint32{1, 2}}}, f.whitelist.flowFilters())
	assert.Nil(t, f.blacklist)
}

func TestToIdentity(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(f)
	require.NoError(t, cmd.Flags().Parse([]string{"--to-identity", "1", "--to-identity", "2"}))
	assert.Equal(t, []*pb.FlowFilter{{DestinationIdentity: []uint32{1, 2}}}, f.whitelist.flowFilters())
	assert.Nil(t, f.blacklist)
}

func TestInvalidIdentity(t *testing.T) {
	f := newObserveFilter()
	cmd := newObserveCmd(f)
	require.Error(t, cmd.Flags().Parse([]string{"--from-identity", "bad"}))
	require.Error(t, cmd.Flags().Parse([]string{"--to-identity", "bad"}))
	require.Error(t, cmd.Flags().Parse([]string{"--identity", "bad"}))
}
