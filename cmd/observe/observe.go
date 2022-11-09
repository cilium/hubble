// SPDX-License-Identifier: Apache-2.0
// Copyright 2019-2021 Authors of Hubble

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
