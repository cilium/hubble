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

func TestEventTypes(t *testing.T) {
	// Make sure to keep event type slices in sync. Agent events, debug
	// events and recorder captures have separate subcommands and are not
	// supported in observe, thus the -3. See eventTypes godoc for details.
	require.Len(t, eventTypes, len(monitorAPI.MessageTypeNames)-3)
	for _, v := range eventTypes {
		require.Contains(t, monitorAPI.MessageTypeNames, v)
	}
	for k := range monitorAPI.MessageTypeNames {
		switch k {
		case monitorAPI.MessageTypeNameAgent,
			monitorAPI.MessageTypeNameDebug,
			monitorAPI.MessageTypeNameRecCapture:
			continue
		}
		require.Contains(t, eventTypes, k)
	}
}

func Test_getRequest(t *testing.T) {
	selectorOpts.since = ""
	selectorOpts.until = ""
	filter := newObserveFilter()
	req, err := getRequest(filter)
	assert.NoError(t, err)
	assert.Equal(t, &observer.GetFlowsRequest{Number: defaults.FlowPrintCount}, req)
	selectorOpts.since = "2021-03-23T00:00:00Z"
	selectorOpts.until = "2021-03-24T00:00:00Z"
	req, err = getRequest(filter)
	assert.NoError(t, err)
	since, err := time.Parse(time.RFC3339, selectorOpts.since)
	assert.NoError(t, err)
	until, err := time.Parse(time.RFC3339, selectorOpts.until)
	assert.NoError(t, err)
	assert.Equal(t, &observer.GetFlowsRequest{
		Number: defaults.FlowPrintCount,
		Since:  timestamppb.New(since),
		Until:  timestamppb.New(until),
	}, req)
}
