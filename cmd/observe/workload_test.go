// Copyright 2022 Authors of Hubble
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

	flowpb "github.com/cilium/cilium/api/v1/flow"
	"github.com/stretchr/testify/assert"
)

func TestParseWorkload(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *flowpb.Workload
	}{
		{
			name:     "empty",
			expected: &flowpb.Workload{},
		},
		{
			name:     "kind and name",
			input:    "Deployment/foo-deploy",
			expected: &flowpb.Workload{Kind: "Deployment", Name: "foo-deploy"},
		},
		{
			name:     "kind only",
			input:    "Deployment/",
			expected: &flowpb.Workload{Kind: "Deployment"},
		},
		{
			name:     "name only", // no trailing /
			input:    "foo-deploy",
			expected: &flowpb.Workload{Name: "foo-deploy"},
		},
		{
			name:  "multiple slashes",
			input: "Deployment/foo/bar/",
			// this isn't a valid resource name, but we don't validate that extensively
			expected: &flowpb.Workload{Kind: "Deployment", Name: "foo/bar/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseWorkload(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
