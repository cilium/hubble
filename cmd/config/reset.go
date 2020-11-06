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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newResetCommand(vp *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "reset [KEY]",
		Short: "Reset all or an individual value in the hubble config file",
		Long: "Reset all or an individual value in the hubble config file.\n" +
			"When KEY is provided, this command is equivalent to 'set KEY'.\n" +
			"If KEY is not provided, all values are reset to their default value.",
		ValidArgs: vp.AllKeys(),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch len(args) {
			case 1:
				return runSet(cmd, vp, args[0], "")
			case 0:
				return runReset(cmd, vp)
			default:
				return fmt.Errorf("invalid arguments: resset requires exactly 0 or 1 argument: got '%s'", strings.Join(args, " "))
			}
		},
	}
}

func runReset(cmd *cobra.Command, vp *viper.Viper) error {
	for _, key := range vp.AllKeys() {
		if err := runSet(cmd, vp, key, ""); err != nil {
			return err
		}
	}
	return nil
}
