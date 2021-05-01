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

package time

import (
	"fmt"
	"strings"
	"time"
)

const (
	// RFC3339Milli is a time format layout for use in time.Format and
	// time.Parse. It follows the RFC3339 format with millisecond precision.
	RFC3339Milli = "2006-01-02T15:04:05.999Z07:00"
	// RFC3339Micro is a time format layout for use in time.Format and
	// time.Parse. It follows the RFC3339 format with microsecond precision.
	RFC3339Micro = "2006-01-02T15:04:05.999999Z07:00"
)

var (
	// Now is a hijackable function for time.Now() that makes unit testing a lot
	// easier for stuff that relies on relative time.
	Now = time.Now
)

// layouts is a set of supported time format layouts. Format that only apply to
// local times should not be added to this list.
var layouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	RFC3339Milli,
	RFC3339Micro,
	time.RFC1123Z,
}

// FromString takes as input a string in either RFC3339 or time.Duration
// format in the past and converts it to a time.Time.
func FromString(input string) (time.Time, error) {
	// try as relative duration first
	d, err := time.ParseDuration(input)
	if err == nil {
		return Now().Add(-d), nil
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, input)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf(
		"failed to convert %s to time", input,
	)
}

// FormatNames are the valid time format names accepted by this package.
var FormatNames = []string{
	"StampMilli",
	"RFC3339",
	"RFC3339Milli",
	"RFC3339Micro",
	"RFC3339Nano",
	"RFC1123Z",
}

// FormatNameToLayout returns the time format layout for the time format name.
func FormatNameToLayout(name string) string {
	switch strings.ToLower(name) {
	case "rfc3339":
		return time.RFC3339
	case "rfc3339milli":
		return RFC3339Milli
	case "rfc3339micro":
		return RFC3339Micro
	case "rfc3339nano":
		return time.RFC3339Nano
	case "rfc1123z":
		return time.RFC1123Z
	case "stampmilli":
		fallthrough
	default:
		return time.StampMilli
	}
}
