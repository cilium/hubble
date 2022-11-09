// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of Hubble

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
