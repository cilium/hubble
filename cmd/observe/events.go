// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

package observe

import (
	"fmt"

	hubprinter "github.com/cilium/hubble/pkg/printer"
	hubtime "github.com/cilium/hubble/pkg/time"
)

func handleEventsArgs(debug bool) error {
	// initialize the printer with any options that were passed in
	var opts = []hubprinter.Option{
		hubprinter.WithTimeFormat(hubtime.FormatNameToLayout(formattingOpts.timeFormat)),
	}

	switch formattingOpts.output {
	case "compact":
		opts = append(opts, hubprinter.Compact())
	case "dict":
		opts = append(opts, hubprinter.Dict())
	case "json", "JSON":
		opts = append(opts, hubprinter.JSON())
	case "jsonpb":
		opts = append(opts, hubprinter.JSONPB())
	case "tab", "table":
		if selectorOpts.follow {
			return fmt.Errorf("table output format is not compatible with follow mode")
		}
		opts = append(opts, hubprinter.Tab())
	default:
		return fmt.Errorf("invalid output format: %s", formattingOpts.output)
	}

	if otherOpts.ignoreStderr {
		opts = append(opts, hubprinter.IgnoreStderr())
	}
	if debug {
		opts = append(opts, hubprinter.WithDebug())
	}
	if formattingOpts.nodeName {
		opts = append(opts, hubprinter.WithNodeName())
	}

	printer = hubprinter.New(opts...)
	return nil
}
