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

package format

import (
	"testing"
	"time"

	"github.com/google/gopacket/layers"
	"github.com/stretchr/testify/assert"
)

func TestMaybeTime(t *testing.T) {
	assert.Equal(t, "N/A", MaybeTime(nil))

	mt := time.Date(2018, time.July, 07, 17, 30, 0, 123000000, time.UTC)
	assert.Equal(t, "Jul  7 17:30:00.123", MaybeTime(&mt))
}

func TestPorts(t *testing.T) {
	orig := EnablePortTranslation
	defer func() {
		EnablePortTranslation = orig
	}()

	EnablePortTranslation = true
	assert.Equal(t, "80(http)", UDPPort(layers.UDPPort(80)))
	assert.Equal(t, "443(https)", TCPPort(layers.TCPPort(443)))
	assert.Equal(t, "4240(cilium-health)", TCPPort(layers.TCPPort(4240)))
	EnablePortTranslation = false
	assert.Equal(t, "80", UDPPort(layers.UDPPort(80)))
	assert.Equal(t, "443", TCPPort(layers.TCPPort(443)))
}

func TestHostname(t *testing.T) {
	orig := EnableIPTranslation
	defer func() {
		EnableIPTranslation = orig
	}()

	EnableIPTranslation = true
	assert.Equal(t, "default/pod", Hostname("", "", "default", "pod", []string{}))
	assert.Equal(t, "a,b", Hostname("", "", "", "", []string{"a", "b"}))
	EnableIPTranslation = false
	assert.Equal(t, "1.1.1.1:80", Hostname("1.1.1.1", "80", "default", "pod", []string{}))
}
