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

package defaults

import (
	"os"
	"time"
)

const (
	// DefaultSocketPathKey is the environment variable name to override the
	// default socket path for observe and status commands.
	DefaultSocketPathKey = "HUBBLE_DEFAULT_SOCKET_PATH"

	// DefaultDialTimeout is the timeout for dialing the server
	DefaultDialTimeout = 5 * time.Second

	// DefaultRequestTimeout is the timeout for client requests
	DefaultRequestTimeout = 12 * time.Second

	// defaultSocketPath on which to connect to the local hubble observer. Use
	// GetDefaultSocketPath to access it.
	defaultSocketPath = "unix:///var/run/cilium/hubble.sock"

	// FlowPrintCount is the default number of flows to print on the hubble
	// observe CLI.
	FlowPrintCount = 20

	// TargetTLSPrefix is a scheme that indicates that the target connection
	// requires TLS.
	TargetTLSPrefix = "tls://"
)

// GetDefaultSocketPath returns the default server for status and observe command.
func GetDefaultSocketPath() string {
	if path, ok := os.LookupEnv(DefaultSocketPathKey); ok {
		return path
	}
	return defaultSocketPath
}
