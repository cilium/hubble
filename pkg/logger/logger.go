// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

package logger

import (
	"sync"

	"golang.org/x/exp/slog" // TODO use std library once Go v1.21 is out
)

var (
	// Logger is a logger that is configured based on viper parameters.
	// Initialize() must be called before accessing it.
	Logger *slog.Logger
	once   sync.Once
)

// Initialize initializes Logger based on config values in viper.
func Initialize(handler slog.Handler) {
	once.Do(func() {
		Logger = slog.New(handler)
	})
}
