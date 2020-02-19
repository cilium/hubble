// Copyright 2017-2019 Authors of Hubble
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
	"io"
	"os"

	"github.com/cilium/hubble/cmd/observe"
	"github.com/cilium/hubble/cmd/serve"
	"github.com/cilium/hubble/cmd/status"
	"github.com/cilium/hubble/pkg/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile                string
	cpuprofile, memprofile string
	log                    *logrus.Entry
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hubble",
	Short: "CLI",
	Long:  `Hubble is a utility to observe and inspect recent Cilium routed traffic in a cluster.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	fin := maybeProfile(log)
	defer fin() // make sure update memory profile is written in the end

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	flags := rootCmd.PersistentFlags()
	flags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hubble.yaml)")
	flags.BoolP("debug", "D", false, "Enable debug messages")
	viper.BindPFlags(flags)
	rootCmd.AddCommand(newCmdCompletion(os.Stdout))
	rootCmd.SetErr(os.Stderr)

	rootCmd.PersistentFlags().StringVar(&cpuprofile,
		"cpuprofile", "", "Enable CPU profiling",
	)
	rootCmd.PersistentFlags().StringVar(&memprofile,
		"memprofile", "", "Enable memory profiling",
	)
	rootCmd.PersistentFlags().Lookup("cpuprofile").Hidden = true
	rootCmd.PersistentFlags().Lookup("memprofile").Hidden = true

	log = logger.GetLogger()

	// initialize all subcommands
	rootCmd.AddCommand(status.New())
	rootCmd.AddCommand(serve.New(log))
	rootCmd.AddCommand(observe.New())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetEnvPrefix("hubble")
	viper.SetConfigName(".hubble") // name of config file (without extension)
	viper.AddConfigPath("$HOME")   // adding home directory as first search path
	viper.AutomaticEnv()           // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

const copyRightHeader = `
# Copyright 2019 Authors of Hubble
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
`

var (
	completionExample = `
# Installing bash completion on macOS using homebrew
## If running Bash 3.2 included with macOS
	brew install bash-completion
## or, if running Bash 4.1+
	brew install bash-completion@2
## afterwards you only need to run
	hubble completion bash > $(brew --prefix)/etc/bash_completion.d/hubble


# Installing bash completion on Linux
## Load the hubble completion code for bash into the current shell
	source <(hubble completion bash)
## Write bash completion code to a file and source if from .bash_profile
	hubble completion bash > ~/.hubble/completion.bash.inc
	printf "
	  # Hubble shell completion
	  source '$HOME/.hubble/completion.bash.inc'
	  " >> $HOME/.bash_profile
	source $HOME/.bash_profile`
)

func newCmdCompletion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "completion [bash]",
		Short:   "Output shell completion code for bash",
		Long:    ``,
		Example: completionExample,
		Run: func(cmd *cobra.Command, args []string) {
			runCompletion(out, cmd, args)
		},
		ValidArgs: []string{"bash"},
	}

	return cmd
}

func runCompletion(out io.Writer, cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("too many arguments; expected only the shell type")
	}
	if _, err := out.Write([]byte(copyRightHeader)); err != nil {
		return err
	}

	return cmd.Parent().GenBashCompletion(out)
}
