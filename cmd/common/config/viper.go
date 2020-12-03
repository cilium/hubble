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

package config

import (
	"strings"

	"github.com/cilium/hubble/pkg/defaults"
	"github.com/spf13/viper"
)

// NewViper creates a new viper instance configured for Hubble.
func NewViper() *viper.Viper {
	vp := viper.New()

	// read config from a file
	vp.SetConfigName("config") // name of config file (without extension)
	vp.SetConfigType("yaml")   // useful if the given config file does not have the extension in the name
	vp.AddConfigPath(".")      // look for a config in the working directory first
	if defaults.ConfigDir != "" {
		vp.AddConfigPath(defaults.ConfigDir)
	}
	if defaults.ConfigDirFallback != "" {
		vp.AddConfigPath(defaults.ConfigDirFallback)
	}

	// read config from environment variables
	vp.SetEnvPrefix("hubble") // env var must start with HUBBLE_
	// replace - by _ for environment variable names
	// (eg: the env var for tls-server-name is TLS_SERVER_NAME)
	vp.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	vp.AutomaticEnv() // read in environment variables that match
	return vp
}
