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

	"github.com/cilium/hubble/cmd/common/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var commandFlagSets = map[string][]*pflag.FlagSet{}

func init() {
	cobra.AddTemplateFunc("title", strings.Title)
	cobra.AddTemplateFunc("getFlagSets", getFlagSets)
}

// RegisterFlagSets registers flags to be included in a commands usage text (--help).
func RegisterFlagSets(cmdName string, flagsets ...*pflag.FlagSet) {
	commandFlagSets[cmdName] = append(commandFlagSets[cmdName], flagsets...)
}

func getFlagSets(cmdName string) []*pflag.FlagSet {
	flagsets, ok := commandFlagSets[cmdName]
	if !ok {
		return []*pflag.FlagSet{config.GlobalFlags}
	}
	return append(flagsets, config.GlobalFlags)
}

const (
	// Usage is the cobra UsageTemplate for Hubble CLI.
	Usage = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}

{{range getFlagSets .Name }}{{ title .Name}} Flags:
{{ .FlagUsages }}
{{end}}Get help:
  -h, --help	Help for any command or subcommand
{{- if .HasHelpSubCommands}}Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
)
