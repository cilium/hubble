// Copyright 2021 Authors of Hubble
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
	"time"

	"github.com/cilium/cilium/api/v1/observer"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	"github.com/cilium/hubble/pkg/defaults"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAgentEventSubTypeMap(t *testing.T) {
	// Make sure to keep agent event sub-types maps in sync. See
	// agentEventSubtypes godoc for details.
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

func Test_getAgentEventsRequest(t *testing.T) {
	selectorOpts.since = ""
	selectorOpts.until = ""
	req, err := getAgentEventsRequest()
	assert.NoError(t, err)
	assert.Equal(t, &observer.GetAgentEventsRequest{Number: defaults.EventsPrintCount}, req)
	selectorOpts.since = "2021-04-26T00:00:00Z"
	selectorOpts.until = "2021-04-26T00:01:00Z"
	req, err = getAgentEventsRequest()
	assert.NoError(t, err)
	since, err := time.Parse(time.RFC3339, selectorOpts.since)
	assert.NoError(t, err)
	until, err := time.Parse(time.RFC3339, selectorOpts.until)
	assert.NoError(t, err)
	assert.Equal(t, &observer.GetAgentEventsRequest{
		Number: defaults.EventsPrintCount,
		Since:  timestamppb.New(since),
		Until:  timestamppb.New(until),
	}, req)
}
