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

package getters

import (
	"net"

	"github.com/cilium/cilium/api/v1/models"
	v1 "github.com/cilium/hubble/pkg/api/v1"
)

// DNSGetter ...
type DNSGetter interface {
	// GetNamesOf fetches FQDNs of a given IP from the perspective of
	// the endpoint with ID sourceEpID
	GetNamesOf(sourceEpID uint64, ip net.IP) (names []string)
}

// EndpointGetter ...
type EndpointGetter interface {
	// GetEndpoint looks up endpoint by IP address.
	GetEndpoint(ip net.IP) (endpoint *v1.Endpoint, ok bool)
}

// IdentityGetter ...
type IdentityGetter interface {
	// GetIdentity fetches a full identity object given a numeric security id.
	GetIdentity(id uint64) (*models.Identity, error)
}

// K8sGetter fetches per-IP K8s metadata
type K8sGetter interface {
	// GetPodNameOf fetches K8s pod and namespace information.
	GetPodNameOf(ip net.IP) (ns, pod string, ok bool)
}
