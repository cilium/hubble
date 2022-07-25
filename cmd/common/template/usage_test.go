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

package template

import (
	"strings"
	"testing"

	"github.com/cilium/hubble/pkg/defaults"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestUsage(t *testing.T) {
	cmd := &cobra.Command{
		Use:     "cmd",
		Aliases: []string{"and", "conquer"},
		Example: "I'm afraid this is not a good an example.",
		Short:   "Do foo with bar",
		Long:    "Do foo with bar and pay attention to baz and more.",
		Run: func(_ *cobra.Command, _ []string) {
			// noop
		},
	}
	flags := pflag.NewFlagSet("bar", pflag.ContinueOnError)
	flags.String("baz", "", "baz usage")
	cmd.Flags().AddFlagSet(flags)

	RegisterFlagSets(cmd, flags)
	cmd.SetUsageTemplate(Usage)

	subCmd := &cobra.Command{
		Use: "subcmd",
		Run: func(_ *cobra.Command, _ []string) {
			// noop
		},
	}
	cmd.AddCommand(subCmd)

	Initialize()

	var out strings.Builder
	cmd.SetOut(&out)
	cmd.Usage()

	var expect strings.Builder
	expect.WriteString(`Usage:
  cmd [flags]
  cmd [command]

Aliases:
  cmd, and, conquer

Examples:
I'm afraid this is not a good an example.

Available Commands:
  subcmd      

Bar Flags:
      --baz string   baz usage

Global Flags:
      --config string   Optional config file (default "`)

	expect.WriteString(defaults.ConfigFile)
	expect.WriteString(`")
  -D, --debug           Enable debug messages

Get help:
  -h, --help	Help for any command or subcommand

Use "cmd [command] --help" for more information about a command.
`)

	require.Equal(t, expect.String(), out.String())
}
