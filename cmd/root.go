// Copyright 2017-2020 Authors of Hubble
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

package cmd

import (
	"os"

	"github.com/cilium/hubble/cmd/observe"
	"github.com/cilium/hubble/cmd/peer"
	"github.com/cilium/hubble/cmd/status"
	"github.com/cilium/hubble/cmd/version"
	"github.com/cilium/hubble/pkg"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:           "hubble",
	Short:         "CLI",
	Long:          `Hubble is a utility to observe and inspect recent Cilium routed traffic in a cluster.`,
	SilenceErrors: true, // this is being handled in main, no need to duplicate error messages
	SilenceUsage:  true,
	Version:       pkg.Version,
}

func addSubcommands() {
	RootCmd.AddCommand(observe.New())
	RootCmd.AddCommand(peer.New())
	RootCmd.AddCommand(status.New())
	RootCmd.AddCommand(version.New())
}

// Execute adds all child commands to the root command sets flags
// appropriately. This is called by main.main(). It only needs to happen once
// to the RootCmd.
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	flags := RootCmd.PersistentFlags()
	flags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hubble.yaml)")
	flags.BoolP("debug", "D", false, "Enable debug messages")
	viper.BindPFlags(flags)
	RootCmd.AddCommand(newCmdCompletion(os.Stdout))
	RootCmd.SetErr(os.Stderr)

	RootCmd.SetVersionTemplate("{{with .Name}}{{printf \"%s \" .}}{{end}}{{printf \"v%s\" .Version}}\n")

	addSubcommands()
}
