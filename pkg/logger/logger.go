// Copyright 2021 Authors of Hubble
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

package logger

import (
	"sync"

	"github.com/cilium/hubble/cmd/common/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	// Logger is a logger that is configured based on viper parameters.
	// Initialize() must be called before accessing it.
	Logger *logrus.Logger
	once   sync.Once
)

// Initialize initializes Logger based on config values in viper.
func Initialize(vp *viper.Viper) {
	once.Do(func() {
		Logger = logrus.New()
		Logger.SetFormatter(&logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
		if vp.GetBool(config.KeyDebug) {
			Logger.SetLevel(logrus.DebugLevel)
		} else {
			Logger.SetLevel(logrus.InfoLevel)
		}
	})
}
