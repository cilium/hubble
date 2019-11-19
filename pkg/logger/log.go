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

package logger

import (
	"sync"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	log  *zap.Logger
	once sync.Once
)

// GetLogger returns the logger properly set up accordingly with the debug flag.
func GetLogger() *zap.Logger {
	once.Do(func() {
		var err error
		if viper.GetBool("debug") {
			log, err = zap.NewDevelopment()
			if err != nil {
				panic(err)
			}
		} else {
			log, err = zap.NewProduction()
			if err != nil {
				panic(err)
			}
		}
	})

	return log
}
