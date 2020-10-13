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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// New config command.
func New(vp *viper.Viper) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Modify or view hubble config",
		Long:  "Modify or view hubble config",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			// override root persistent pre-run to avoid flag/config checks
			// as we want to be able to modify/view the config even if it is
			// invalid
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	configCmd.AddCommand(
		newGetCommand(vp),
		newResetCommand(vp),
		newSetCommand(vp),
		newViewCommand(vp),
	)
	return configCmd
}

func isKey(vp *viper.Viper, key string) bool {
	for _, k := range vp.AllKeys() {
		if key == k {
			return true
		}
	}
	return false
}
