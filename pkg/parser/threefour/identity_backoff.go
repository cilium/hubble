// Copyright 2020 Authors of Hubble
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

package threefour

import (
	"sync"
	"time"

	"github.com/cilium/cilium/pkg/backoff"
	hubtime "github.com/cilium/hubble/pkg/time"
)

const (
	defaultBurst      = 5
	defaultMinBackoff = 100 * time.Millisecond
	defaultMaxBackoff = 1 * time.Minute
)

func newBackoff() *identityBackoff {
	return &identityBackoff{
		burst: defaultBurst,
		exp: &backoff.Exponential{
			Min:    defaultMinBackoff,
			Max:    defaultMaxBackoff,
			Jitter: true,
		},
		bm: make(map[uint64]*tracker),
	}
}

// Per-identity request back-off to the Cilium API.
//
// Because Cilium IP state is synced only periodically, there is a window of
// time between ipcache sync where flows will be coming in with what hubble
// would consider to be invalid ips, or ips that map to old identities. During
// those times it's important not to flood the cilium API.
//
// In general, all this code is a temporary work-around for stand-alone hubble
// server running on Cilium <=1.7.X. Cilium 1.8+ doesn't have synchronization
// issues like this as the hubble server is embedded within cilium process.
type identityBackoff struct {
	sync.RWMutex

	burst int // do not back-off until burst+1 failed requests
	exp   *backoff.Exponential
	bm    map[uint64]*tracker
}

type tracker struct {
	attempt     int
	lastAttempt time.Time // last time it was tried
}

func (ib *identityBackoff) allowed(i uint64) bool {
	ib.Lock()
	defer ib.Unlock()

	tr, ok := ib.bm[i]
	if !ok {
		// no backoff is configured for this identity, carry on
		return true
	}

	if tr.attempt < ib.burst {
		// failed attempts are still under burst, allow the request
		return true
	}

	// burst has been exhausted for this id, only allow if time passed since
	// last attempt is allowed by the backoff.
	nextAllowed := tr.lastAttempt.Add(ib.exp.Duration(tr.attempt - ib.burst))
	if hubtime.Now().After(nextAllowed) {
		return true
	}

	return false
}

func (ib *identityBackoff) failed(i uint64) {
	ib.Lock()
	defer ib.Unlock()

	tr, ok := ib.bm[i]
	if !ok {
		tr = &tracker{}
	}

	tr.attempt++
	tr.lastAttempt = hubtime.Now()

	ib.bm[i] = tr
}

func (ib *identityBackoff) clear(i uint64) {
	ib.Lock()
	defer ib.Unlock()

	if _, ok := ib.bm[i]; ok {
		delete(ib.bm, i)
	}
}
