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

package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func newViewCommand(vp *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "Display merged configuration settings",
		Long:  "Display merged configuration settings",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runView(cmd, vp)
		},
	}
}

func runView(cmd *cobra.Command, vp *viper.Viper) error {
	bs, err := yaml.Marshal(vp.AllSettings())
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}
	_, err = fmt.Fprint(cmd.OutOrStdout(), string(bs))
	return err
}
