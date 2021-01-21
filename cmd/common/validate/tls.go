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

package validate

import (
	"errors"
	"strings"

	"github.com/cilium/hubble/cmd/common/config"
	"github.com/cilium/hubble/pkg/defaults"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// ErrInvalidKeypair means that a TLS keypair is required but only one of the
	// key or certificate was provided.
	ErrInvalidKeypair = errors.New("certificate and private key are both required, but only one was provided")

	// ErrTLSRequired means that Transport Layer Security (TLS) is required but
	// not set.
	ErrTLSRequired = errors.New("transport layer security required")
)

func init() {
	FlagFuncs = append(FlagFuncs, validateMutualTLSFlags)
}

// validateMutualTLSFlags validates that TLS is set if a client keypair is
// provided.
func validateMutualTLSFlags(_ *cobra.Command, vp *viper.Viper) error {
	var needTLS bool
	switch {
	case vp.GetString(config.KeyTLSClientKeyFile) != "" && vp.GetString(config.KeyTLSClientCertFile) != "":
		needTLS = true
	case vp.GetString(config.KeyTLSClientKeyFile) != "":
		fallthrough
	case vp.GetString(config.KeyTLSClientCertFile) != "":
		return ErrInvalidKeypair
	}

	if needTLS && !(vp.GetBool(config.KeyTLS) || strings.HasPrefix(vp.GetString(config.KeyServer), defaults.TargetTLSPrefix)) {
		return ErrTLSRequired
	}
	return nil
}
