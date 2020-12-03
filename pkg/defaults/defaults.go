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
	"path/filepath"
	"time"
)

const (
	// DialTimeout is the default timeout for dialing the server.
	DialTimeout = 5 * time.Second

	// RequestTimeout is the default timeout for client requests.
	RequestTimeout = 12 * time.Second

	// FlowPrintCount is the default number of flows to print on the hubble
	// observe CLI.
	FlowPrintCount = 20

	// TargetTLSPrefix is a scheme that indicates that the target connection
	// requires TLS.
	TargetTLSPrefix = "tls://"

	// socketPathKey is the environment variable name to override the default
	// socket path for observe and status commands.
	socketPathKey = "HUBBLE_DEFAULT_SOCKET_PATH"

	// socketPath is the path of the socket on which to connect to the local
	// hubble observer. Use GetDefaultSocketPath to access it.
	socketPath = "unix:///var/run/cilium/hubble.sock"
)

var (
	// ConfigDir is the default directory path to store Hubble
	// configuration files. It may be unset.
	ConfigDir string
	// ConfigDirFallback is the directory path to store Hubble configuration
	// files if defaultConfigDir is unset. Note that it may also be unset.
	ConfigDirFallback string
	// ConfigFile is the path to an optional configuration file.
	// It may be unset.
	ConfigFile string
)

func init() {
	// honor user config dir
	if dir, err := os.UserConfigDir(); err == nil {
		ConfigDir = filepath.Join(dir, "hubble")
	}
	// fallback to home directory
	if dir, err := os.UserHomeDir(); err == nil {
		ConfigDirFallback = filepath.Join(dir, ".hubble")
	}

	switch {
	case ConfigDir != "":
		ConfigFile = filepath.Join(ConfigDir, "config.yaml")
	case ConfigDirFallback != "":
		ConfigFile = filepath.Join(ConfigDirFallback, "config.yaml")
	}
}

// GetSocketPath returns the default server for status and observe command.
func GetSocketPath() string {
	if path, ok := os.LookupEnv(socketPathKey); ok {
		return path
	}
	return socketPath
}
