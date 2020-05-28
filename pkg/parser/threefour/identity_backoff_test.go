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
	"testing"
	"time"

	"github.com/cilium/cilium/pkg/backoff"
	hubtime "github.com/cilium/hubble/pkg/time"
	"github.com/stretchr/testify/assert"
)

func fail(bo *identityBackoff, id uint64, times int) {
	for i := 0; i < times; i++ {
		bo.failed(id)
	}
}

func TestBurst(t *testing.T) {
	bo := &identityBackoff{
		burst: defaultBurst,
		exp: &backoff.Exponential{
			Min:    defaultMinBackoff,
			Max:    defaultMaxBackoff,
			Jitter: true,
		},
		bm: make(map[uint64]*tracker),
	}

	id := uint64(42)
	assert.True(t, bo.allowed(id), "should be allowed by default")

	fail(bo, id, defaultBurst/2)
	assert.True(t, bo.allowed(id), "should still be allowed after half a burst")

	fail(bo, id, defaultBurst)
	assert.False(t, bo.allowed(id), "should be disallowed after burst exhaustion")

	// hijack the current time and then restore it after the test
	hubtime.Now = func() time.Time {
		// fast-forward time by max backoff
		return time.Now().Add(defaultMaxBackoff)
	}
	defer func() {
		hubtime.Now = time.Now
	}()

	assert.True(t, bo.allowed(id), "should be allowed after a passage of time")

	fail(bo, id, defaultBurst*10)
	assert.False(t, bo.allowed(id), "should be disallowed after a lot of failures")
	bo.clear(id)
	assert.True(t, bo.allowed(id), "should be allowed after clear()")
}
