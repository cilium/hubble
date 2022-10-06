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
	"strings"

	flowpb "github.com/cilium/cilium/api/v1/flow"
)

// parseWorkload parse and returns workloads
func parseWorkload(s string) *flowpb.Workload {
	if s == "" {
		return &flowpb.Workload{}
	}
	var kind, name string
	elements := strings.SplitN(s, "/", 2)
	if len(elements) == 1 { // foo-deploy
		name = elements[0]
	} else { // Deployment/foo-deploy and Deployment/
		kind, name = elements[0], elements[1]
	}
	return &flowpb.Workload{Kind: kind, Name: name}
}
