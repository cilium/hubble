// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of Hubble

package observe

import (
	"testing"
	"time"

	"github.com/cilium/cilium/api/v1/observer"
	"github.com/cilium/hubble/pkg/defaults"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_getDebugEventsRequest(t *testing.T) {
	selectorOpts.since = ""
	selectorOpts.until = ""
	req, err := getDebugEventsRequest()
	assert.NoError(t, err)
	assert.Equal(t, &observer.GetDebugEventsRequest{Number: defaults.EventsPrintCount}, req)
	selectorOpts.since = "2021-04-26T01:00:00Z"
	selectorOpts.until = "2021-04-26T01:01:00Z"
	req, err = getDebugEventsRequest()
	assert.NoError(t, err)
	since, err := time.Parse(time.RFC3339, selectorOpts.since)
	assert.NoError(t, err)
	until, err := time.Parse(time.RFC3339, selectorOpts.until)
	assert.NoError(t, err)
	assert.Equal(t, &observer.GetDebugEventsRequest{
		Number: defaults.EventsPrintCount,
		Since:  timestamppb.New(since),
		Until:  timestamppb.New(until),
	}, req)
}

func Test_getDebugEventsRequestWithoutSince(t *testing.T) {
	selectorOpts.since = ""
	selectorOpts.until = ""
	req, err := getDebugEventsRequest()
	assert.NoError(t, err)
	assert.Equal(t, &observer.GetDebugEventsRequest{Number: defaults.EventsPrintCount}, req)
	selectorOpts.until = "2021-04-26T01:01:00Z"
	req, err = getDebugEventsRequest()
	assert.NoError(t, err)
	until, err := time.Parse(time.RFC3339, selectorOpts.until)
	assert.NoError(t, err)
	assert.Equal(t, &observer.GetDebugEventsRequest{
		Number: defaults.EventsPrintCount,
		Until:  timestamppb.New(until),
	}, req)
}
