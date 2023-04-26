// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

package validate

import (
	"fmt"

	"github.com/cilium/hubble/cmd/common/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// ErrInvalidBasicAuthCredentials occurs when only one of basic-auth-user or basic-auth-password is configured.
	ErrInvalidBasicAuthCredentials = fmt.Errorf("must specify both %s and %s", config.KeyBasicAuthUsername, config.KeyBasicAuthPassword)
)

func init() {
	FlagFuncs = append(FlagFuncs, validateBasicAuth)
}

// validateBasicAuth validates that both username and password are set.
func validateBasicAuth(_ *cobra.Command, vp *viper.Viper) error {
	if vp.GetString(config.KeyBasicAuthUsername) != "" && vp.GetString(config.KeyBasicAuthPassword) == "" ||
		vp.GetString(config.KeyBasicAuthUsername) == "" && vp.GetString(config.KeyBasicAuthPassword) != "" {
		return ErrInvalidBasicAuthCredentials
	}
	return nil
}
