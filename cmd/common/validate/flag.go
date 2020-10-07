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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// FlagFunc is a function that validates a flag or set of flags of cmd.
type FlagFunc func(cmd *cobra.Command, vp *viper.Viper) error

// FlagFuncs is a combination of multiple flag validation functions.
var FlagFuncs []FlagFunc

// Flags validates flags for the given command.
func Flags(cmd *cobra.Command, vp *viper.Viper) error {
	for _, fn := range FlagFuncs {
		if err := fn(cmd, vp); err != nil {
			return fmt.Errorf("invalid flag(s): %w", err)
		}
	}
	return nil
}
