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
	"io"

	"github.com/spf13/cobra"
)

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
	source $HOME/.bash_profile

# Installing zsh completion on Linux/macOS
## Load the hubble completion code for zsh into the current shell
        source <(hubble completion zsh)
## Write zsh completion code to a file and source if from .zshrc
        hubble completion zsh > ~/.hubble/completion.zsh.inc
        printf "
          # Hubble shell completion
          source '$HOME/.hubble/completion.zsh.inc'
          " >> $HOME/.zshrc
        source $HOME/.zshrc

# Installing fish completion on Linux/macOS
## Load the hubble completion code for fish into the current shell
        hubble completion fish | source
## Write fish completion code to a file
        hubble completion fish > ~/.config/fish/completions/hubble.fish`
)

func newCmdCompletion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "completion [shell]",
		Short:   "Output shell completion code",
		Long:    ``,
		Example: completionExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompletion(out, cmd, args)
		},
		ValidArgs: []string{"bash", "fish", "powershell", "ps1", "zsh"},
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

	if len(args) == 0 {
		return cmd.Root().GenBashCompletion(out)
	}

	switch args[0] {
	case "bash":
		return cmd.Root().GenBashCompletion(out)
	case "zsh":
		return cmd.Root().GenZshCompletion(out)
	case "fish":
		return cmd.Root().GenFishCompletion(out, true)
	case "powershell", "ps1":
		return cmd.Root().GenPowerShellCompletion(out)
	}
	return fmt.Errorf("unsupported shell type: %s", args[0])
}
