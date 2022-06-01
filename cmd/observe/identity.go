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
	"sort"

	"github.com/cilium/cilium/pkg/identity"
)

// reservedIdentitiesNames returns a slice of all the reserved identity
// strings.
func reservedIdentitiesNames() []string {
	identities := identity.GetAllReservedIdentities()
	// NOTE: identity.GetAllReservedIdentities() returned values are sorted in
	// a random order due to be sourced from a map. We sort them here in
	// identity order to ensure consistency before converting them to strings.
	// Once https://github.com/cilium/cilium/pull/20048 is merged and vendored,
	// we can remove this sort.
	sort.Slice(identities, func(i, j int) bool {
		return identities[i].Uint32() < identities[j].Uint32()
	})

	names := make([]string, len(identities))
	for i, id := range identities {
		names[i] = id.String()
	}

	return names
}

// parseIdentity parse and return both numeric and reserved identities, or an
// error.
func parseIdentity(s string) (identity.NumericIdentity, error) {
	if id := identity.GetReservedID(s); id != identity.IdentityUnknown {
		return id, nil
	}
	return identity.ParseNumericIdentity(s)
}
