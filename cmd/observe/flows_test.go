// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

package observe

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	flowpb "github.com/cilium/cilium/api/v1/flow"
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
	// supported in observe, thus the 3. See flowEventTypes godoc for details.
	require.Len(t, flowEventTypes, len(monitorAPI.MessageTypeNames)-3)
	for _, v := range flowEventTypes {
		require.Contains(t, monitorAPI.MessageTypeNames, v)
	}
	for k := range monitorAPI.MessageTypeNames {
		switch k {
		case monitorAPI.MessageTypeNameAgent,
			monitorAPI.MessageTypeNameDebug,
			monitorAPI.MessageTypeNameRecCapture:
			continue
		}
		require.Contains(t, flowEventTypes, k)
	}
}

func Test_getFlowsRequest(t *testing.T) {
	selectorOpts.since = ""
	selectorOpts.until = ""
	filter := newFlowFilter()
	req, err := getFlowsRequest(filter, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, &observer.GetFlowsRequest{Number: defaults.FlowPrintCount}, req)
	selectorOpts.since = "2021-03-23T00:00:00Z"
	selectorOpts.until = "2021-03-24T00:00:00Z"
	req, err = getFlowsRequest(filter, nil, nil)
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

func Test_getFlowsRequestWithoutSince(t *testing.T) {
	selectorOpts.since = ""
	selectorOpts.until = ""
	filter := newFlowFilter()
	req, err := getFlowsRequest(filter, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, &observer.GetFlowsRequest{Number: defaults.FlowPrintCount}, req)
	selectorOpts.until = "2021-03-24T00:00:00Z"
	req, err = getFlowsRequest(filter, nil, nil)
	assert.NoError(t, err)
	until, err := time.Parse(time.RFC3339, selectorOpts.until)
	assert.NoError(t, err)
	assert.Equal(t, &observer.GetFlowsRequest{
		Number: defaults.FlowPrintCount,
		Until:  timestamppb.New(until),
	}, req)
}

func Test_getFlowsRequestWithRawFilters(t *testing.T) {
	allowlist := []string{
		`{"source_label":["k8s:io.kubernetes.pod.namespace=kube-system","reserved:host"]}`,
		`{"destination_label":["k8s:io.kubernetes.pod.namespace=kube-system","reserved:host"]}`,
	}
	denylist := []string{
		`{"source_label":["k8s:k8s-app=kube-dns"]}`,
		`{"destination_label":["k8s:k8s-app=kube-dns"]}`,
	}
	req, err := getFlowsRequest(newFlowFilter(), allowlist, denylist)
	assert.NoError(t, err)

	// convert filters in the request back to JSON and check if it matches the original allowlist/denylist.
	var b strings.Builder
	err = json.NewEncoder(&b).Encode(req.Whitelist)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("[%s]\n", strings.Join(allowlist, ",")), b.String())
	b.Reset()
	err = json.NewEncoder(&b).Encode(req.Blacklist)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("[%s]\n", strings.Join(denylist, ",")), b.String())
}

func Test_getFlowsRequestWithInvalidRawFilters(t *testing.T) {
	filters := []string{
		`{"invalid":["filters"]}`,
	}
	_, err := getFlowsRequest(newFlowFilter(), filters, nil)
	assert.Contains(t, err.Error(), `invalid --allowlist flag: failed to decode '{"invalid":["filters"]}': `)
	_, err = getFlowsRequest(newFlowFilter(), nil, filters)
	assert.Contains(t, err.Error(), `invalid --denylist flag: failed to decode '{"invalid":["filters"]}': `)
}

func Test_getFlowFiltersYAML(t *testing.T) {
	req := observer.GetFlowsRequest{
		Whitelist: []*flowpb.FlowFilter{{SourceIp: []string{"1.2.3.4/16"}}},
		Blacklist: []*flowpb.FlowFilter{{SourcePort: []string{"80"}}},
	}
	out, err := getFlowFiltersYAML(&req)
	expected := `allowlist:
- '{"source_ip":["1.2.3.4/16"]}'
denylist:
- '{"source_port":["80"]}'
`
	assert.NoError(t, err)
	assert.Equal(t, expected, out)
}
