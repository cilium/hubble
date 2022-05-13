// Copyright 2019-2021 Authors of Hubble
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

package observe

import (
	hubprinter "github.com/cilium/hubble/pkg/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	selectorOpts struct {
		all          bool
		last         uint64
		since, until string
		follow       bool
		first        uint64
	}

	formattingOpts struct {
		output string

		timeFormat string

		enableIPTranslation bool
		nodeName            bool
		numeric             bool
		color               string
	}

	otherOpts struct {
		ignoreStderr    bool
		printRawFilters bool
	}

	printer *hubprinter.Printer
)

// New observer command.
func New(vp *viper.Viper) *cobra.Command {
	return newFlowsCmd(vp, newFlowFilter())
}
