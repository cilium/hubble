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
	"fmt"
	"os"

	"github.com/cilium/hubble/cmd/completion"
	"github.com/cilium/hubble/cmd/observe"
	"github.com/cilium/hubble/cmd/peer"
	"github.com/cilium/hubble/cmd/status"
	"github.com/cilium/hubble/cmd/version"
	"github.com/cilium/hubble/pkg"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// New create a new root command.
func New() *cobra.Command {
	vp := newViper()
	rootCmd := &cobra.Command{
		Use:           "hubble",
		Short:         "CLI",
		Long:          `Hubble is a utility to observe and inspect recent Cilium routed traffic in a cluster.`,
		SilenceErrors: true, // this is being handled in main, no need to duplicate error messages
		SilenceUsage:  true,
		Version:       pkg.Version,
	}

	cobra.OnInitialize(func() {
		if cfg := vp.GetString("config"); cfg != "" { // enable ability to specify config file via flag
			vp.SetConfigFile(cfg)
		}
		// if a config file is found, read it in.
		if err := vp.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", vp.ConfigFileUsed())
		}
	})

	flags := rootCmd.PersistentFlags()
	flags.String("config", "", "config file (default is $HOME/.hubble.yaml)")
	flags.BoolP("debug", "D", false, "Enable debug messages")
	vp.BindPFlags(flags)
	rootCmd.SetErr(os.Stderr)
	rootCmd.SetVersionTemplate("{{with .Name}}{{printf \"%s \" .}}{{end}}{{printf \"v%s\" .Version}}\n")

	rootCmd.AddCommand(
		completion.New(),
		observe.New(vp),
		peer.New(),
		status.New(),
		version.New(),
	)
	return rootCmd
}

// Execute creates the root command and executes it.
func Execute() error {
	return New().Execute()
}

// newViper creates a new viper instance configured for Hubble.
func newViper() *viper.Viper {
	vp := viper.New()
	vp.SetEnvPrefix("hubble")
	vp.SetConfigName(".hubble") // name of config file (without extension)
	vp.AddConfigPath("$HOME")   // adding home directory as first search path
	vp.AutomaticEnv()           // read in environment variables that match
	return vp
}
