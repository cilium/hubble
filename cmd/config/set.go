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
	"encoding/csv"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newSetCommand(vp *viper.Viper) *cobra.Command {
	return &cobra.Command{
		Use:   "set KEY [VALUE]",
		Short: "Set an individual value in the hubble config file",
		Long: "Set an individual value in the hubble config file.\n" +
			"If VALUE is not provided, the value is reset to its default value.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var val string
			switch len(args) {
			case 2:
				val = args[1]
				fallthrough
			case 1:
				return runSet(cmd, vp, args[0], val)
			default:
				return fmt.Errorf("invalid arguments: set requires exactly 1 or 2 argument(s): got '%s'", strings.Join(args, " "))
			}
		},
	}
}

func runSet(cmd *cobra.Command, vp *viper.Viper, key, value string) error {
	if !isKey(vp, key) {
		return fmt.Errorf("unknown key: %s", key)
	}

	// each viper key/value entry should be bound to a command flag
	flag := cmd.Flag(key)
	if flag == nil {
		return fmt.Errorf("key=%s not bound to a flag", key)
	}

	val := value
	if value == "" {
		val = flag.DefValue
	}

	var err error
	var newVal interface{}
	typ := flag.Value.Type()
	switch typ {
	case "bool":
		newVal, err = cast.ToBoolE(val)
	case "duration":
		newVal, err = cast.ToDurationE(val)
	case "int":
		newVal, err = cast.ToIntE(val)
	case "string":
		newVal = val
	case "stringSlice":
		val = strings.TrimSuffix(strings.TrimPrefix(val, "["), "]")
		if val == "" {
			newVal = []string{} // csv reader would return io.EOF
		} else {
			newVal, err = csv.NewReader(strings.NewReader(val)).Read()
		}
	default:
		return fmt.Errorf("unhandeld type %s, please open an issue", typ)
	}
	if err != nil {
		return fmt.Errorf("cannot assign value=%s for key=%s, expected type=%s: %w", value, key, typ, err)
	}

	// Create a file-only viper config from the configured file to avoid
	// writing defaults and/or values set via environment variables or flags.
	// This viper config is only used to write the resulting config.
	// This method also prevents from writing default values for all keys
	// therefore only writing key/value pairs explicitely set by the caller.
	configPath := vp.GetString("config")
	fileVP, err := newFileOnlyViper(configPath)
	if err != nil {
		return err
	}
	fileVP.Set(key, newVal)
	return fileVP.WriteConfigAs(configPath)
}

// newFileOnlyViper creates a new viper config that only reads from the given
// configuration file and is not bound to any environment variable or flag.
func newFileOnlyViper(configPath string) (*viper.Viper, error) {
	if configPath == "" {
		return nil, errors.New("config file undefined")
	}

	path := filepath.Clean(configPath)
	base := filepath.Base(path)
	ext := filepath.Ext(path)

	dir := filepath.Dir(path)
	filename := strings.TrimSuffix(base, ext)
	typ := strings.TrimPrefix(ext, ".")

	vp := viper.New()
	vp.SetConfigName(filename)
	vp.SetConfigType(typ)
	vp.AddConfigPath(dir)

	if err := vp.ReadInConfig(); err != nil {
		// it's OK so long as the failure is ConfigFileNotFound
		// for all other cases, failing to read the config should be an error
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}
	return vp, nil
}
