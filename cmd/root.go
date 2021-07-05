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

	"github.com/cilium/hubble/cmd/common/config"
	"github.com/cilium/hubble/cmd/common/conn"
	"github.com/cilium/hubble/cmd/common/template"
	"github.com/cilium/hubble/cmd/common/validate"
	cmdConfig "github.com/cilium/hubble/cmd/config"
	"github.com/cilium/hubble/cmd/list"
	"github.com/cilium/hubble/cmd/observe"
	"github.com/cilium/hubble/cmd/record"
	"github.com/cilium/hubble/cmd/reflect"
	"github.com/cilium/hubble/cmd/status"
	"github.com/cilium/hubble/cmd/version"
	"github.com/cilium/hubble/cmd/watch"
	"github.com/cilium/hubble/pkg"
	"github.com/cilium/hubble/pkg/logger"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// New create a new root command.
func New() *cobra.Command {
	return NewWithViper(config.NewViper())
}

// NewWithViper creates a new root command with the given viper.
func NewWithViper(vp *viper.Viper) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "hubble",
		Short:         "CLI",
		Long:          `Hubble is a utility to observe and inspect recent Cilium routed traffic in a cluster.`,
		SilenceErrors: true, // this is being handled in main, no need to duplicate error messages
		SilenceUsage:  true,
		Version:       pkg.Version,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := validate.Flags(cmd, vp); err != nil {
				return err
			}
			return conn.Init(vp)
		},
	}

	cobra.OnInitialize(func() {
		if cfg := vp.GetString(config.KeyConfig); cfg != "" { // enable ability to specify config file via flag
			vp.SetConfigFile(cfg)
		}
		// if a config file is found, read it in.
		err := vp.ReadInConfig()
		// initialize the logger after all the config parameters get loaded to viper.
		logger.Initialize(vp)
		if err == nil {
			logger.Logger.WithField("config-file", vp.ConfigFileUsed()).Debug("Using config file")
		}
	})

	flags := rootCmd.PersistentFlags()
	// config.GlobalFlags can be used with any command
	flags.AddFlagSet(config.GlobalFlags)
	// config.ServerFlags is added to the root command's persistent flags
	// so that "hubble --server foo observe" still works
	flags.AddFlagSet(config.ServerFlags)
	vp.BindPFlags(flags)

	// config.ServerFlags is only useful to a subset of commands so do not
	// add it by default in the help template
	// config.GlobalFlags is always added to the help template as it's global
	// to all commands
	rootCmd.SetUsageTemplate(template.Usage())

	rootCmd.SetErr(os.Stderr)
	rootCmd.SetVersionTemplate("{{with .Name}}{{printf \"%s \" .}}{{end}}{{printf \"v%s\" .Version}}\r\n")

	rootCmd.AddCommand(
		cmdConfig.New(vp),
		list.New(vp),
		observe.New(vp),
		record.New(vp),
		reflect.New(vp),
		status.New(vp),
		version.New(),
		watch.New(vp),
	)
	return rootCmd
}

// Execute creates the root command and executes it.
func Execute() error {
	return New().Execute()
}
