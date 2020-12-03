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
	"fmt"
	"strings"

	"github.com/cilium/hubble/cmd/common/config"
	"github.com/spf13/pflag"
)

const (
	header = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}

`
	footer = `{{- if .HasHelpSubCommands}}Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
)

// Usage returns a usage template string with the given sets of flags.
// Each flag set is separated in a new flags section with the flagset name as
// section title.
// When used with cobra commands, the resulting template may be passed as a
// parameter to command.SetUsageTemplate().
func Usage(flagSets ...*pflag.FlagSet) string {
	var b strings.Builder
	b.WriteString(header)
	for _, fs := range append(flagSets, config.GlobalFlags) {
		fmt.Fprintf(&b, "%s Flags:\n", strings.Title(fs.Name()))
		fmt.Fprintln(&b, fs.FlagUsages())
	}
	// treat the special --help flag separately
	b.WriteString("Get help:\n  -h, --help	Help for any command or subcommand")
	b.WriteString(footer)
	return b.String()
}
