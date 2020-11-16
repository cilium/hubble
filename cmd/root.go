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
	"path/filepath"
	"strings"

	"github.com/cilium/hubble/cmd/common/conn"
	"github.com/cilium/hubble/cmd/common/validate"
	"github.com/cilium/hubble/cmd/completion"
	"github.com/cilium/hubble/cmd/config"
	"github.com/cilium/hubble/cmd/node"
	"github.com/cilium/hubble/cmd/observe"
	"github.com/cilium/hubble/cmd/peer"
	"github.com/cilium/hubble/cmd/reflect"
	"github.com/cilium/hubble/cmd/status"
	"github.com/cilium/hubble/cmd/version"
	"github.com/cilium/hubble/pkg"
	"github.com/cilium/hubble/pkg/defaults"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// defaultConfigDir is the default directory path to store Hubble
	// configuration files.
	defaultConfigDir string
	// fallbackConfigDir is the directory path to store Hubble configuration
	// files if defaultConfigDir is unset. Note that it might also be unset.
	fallbackConfigDir string
	// defaultConfigFile is the path to an optional configuration file.
	// It might be unset.
	defaultConfigFile string
)

func init() {
	// honor user config dir
	if dir, err := os.UserConfigDir(); err == nil {
		defaultConfigDir = filepath.Join(dir, "hubble")
	}
	// fallback to home directory
	if dir, err := os.UserHomeDir(); err == nil {
		fallbackConfigDir = filepath.Join(dir, ".hubble")
	}

	switch {
	case defaultConfigDir != "":
		defaultConfigFile = filepath.Join(defaultConfigDir, "config.yaml")
	case fallbackConfigDir != "":
		defaultConfigFile = filepath.Join(fallbackConfigDir, "config.yaml")
	}
}

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
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := validate.Flags(cmd, vp); err != nil {
				return err
			}
			return conn.Init(vp)
		},
	}

	cobra.OnInitialize(func() {
		if cfg := vp.GetString("config"); cfg != "" { // enable ability to specify config file via flag
			vp.SetConfigFile(cfg)
		}
		// if a config file is found, read it in.
		if err := vp.ReadInConfig(); err == nil && vp.GetBool("debug") {
			fmt.Fprintln(rootCmd.ErrOrStderr(), "Using config file:", vp.ConfigFileUsed())
		}
	})

	flags := rootCmd.PersistentFlags()
	flags.String("config", defaultConfigFile, "Optional config file")
	flags.BoolP("debug", "D", false, "Enable debug messages")
	flags.String("server", defaults.GetSocketPath(), "Address of a Hubble server")
	flags.Duration("timeout", defaults.DialTimeout, "Hubble server dialing timeout")
	flags.Bool(
		"tls",
		false,
		"Specify that TLS must be used when establishing a connection to a Hubble server.\r\n"+
			"By default, TLS is only enabled if the server address starts with 'tls://'.",
	)
	flags.Bool(
		"tls-allow-insecure",
		false,
		"Allows the client to skip verifying the server's certificate chain and host name.\r\n"+
			"This option is NOT recommended as, in this mode, TLS is susceptible to machine-in-the-middle attacks.\r\n"+
			"See also the 'tls-server-name' option which allows setting the server name.",
	)
	flags.StringSlice(
		"tls-ca-cert-files",
		nil,
		"Paths to custom Certificate Authority (CA) certificate files."+
			"The files must contain PEM encoded data.",
	)
	flags.String(
		"tls-client-cert-file",
		"",
		"Path to the public key file for the client certificate to connect to a Hubble server (implies TLS).\r\n"+
			"The file must contain PEM encoded data.",
	)
	flags.String(
		"tls-client-key-file",
		"",
		"Path to the private key file for the client certificate to connect a Hubble server (implies TLS).\r\n"+
			"The file must contain PEM encoded data.",
	)
	flags.String(
		"tls-server-name",
		"",
		"Specify a server name to verify the hostname on the returned certificate (eg: 'instance.hubble-relay.cilium.io').",
	)
	vp.BindPFlags(flags)

	rootCmd.SetErr(os.Stderr)
	rootCmd.SetVersionTemplate("{{with .Name}}{{printf \"%s \" .}}{{end}}{{printf \"v%s\" .Version}}\r\n")

	rootCmd.AddCommand(
		completion.New(),
		config.New(vp),
		node.New(vp),
		observe.New(vp),
		peer.New(vp),
		reflect.New(vp),
		status.New(vp),
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

	// read config from a file
	vp.SetConfigName("config") // name of config file (without extension)
	vp.SetConfigType("yaml")   // useful if the given config file does not have the extension in the name
	vp.AddConfigPath(".")      // look for a config in the working directory first
	if defaultConfigDir != "" {
		vp.AddConfigPath(defaultConfigDir)
	}
	if fallbackConfigDir != "" {
		vp.AddConfigPath(fallbackConfigDir)
	}

	// read config from environment variables
	vp.SetEnvPrefix("hubble") // env var must start with HUBBLE_
	// replace - by _ for environment variable names
	// (eg: the env var for tls-server-name is TLS_SERVER_NAME)
	vp.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	vp.AutomaticEnv() // read in environment variables that match
	return vp
}
