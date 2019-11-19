// Copyright 2019 Authors of Hubble
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
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestObserveUsage(t *testing.T) {
	cmd := &cobra.Command{
		Use: "cmd subcmd [foo]",
		Run: func(cmd *cobra.Command, args []string) {
			// noop
		},
	}
	cmd.Flags().String("last", "", "last selector")
	cmd.Flags().String("to-fqdn", "", "to-fqdn usage")
	cmd.Flags().String("verdict", "", "verdict filter")
	cmd.Flags().String("something-else", "", "some other flag")
	customObserverHelp(cmd)

	var b bytes.Buffer
	cmd.SetOut(&b)
	cmd.Help()

	require.Equal(t, strings.TrimSpace(`Usage:
  cmd subcmd [foo] [flags]

Selectors (retrieve data from hubble):
      --last string   last selector

Filters (limit result set, not all are compatible with each other):
      --to-fqdn string   to-fqdn usage
      --verdict string   verdict filter

Other Flags:
      --something-else string   some other flag

Global Flags:`), strings.TrimSpace(b.String()))
}
