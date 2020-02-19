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
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/sirupsen/logrus"
)

// Look at the set observe flags, and optionally enable cpu, memory, or both,
// profiling.
//
// Returns a function which should be deferred to the end of the execution so
// profiles can be finalized.
func maybeProfile(log *logrus.Entry) func() {
	var cf, mf *os.File
	var err error
	if cpuprofile != "" {
		cf, err = os.Create(cpuprofile)
		if err != nil {
			log.WithError(err).Fatal("failed to create cpu profile")
		}
		pprof.StartCPUProfile(cf)
	}

	if memprofile != "" {
		mf, err = os.Create(memprofile)
		if err != nil {
			log.WithError(err).Fatal("failed to create memory profile")
		}
	}

	return func() {
		if cf != nil {
			pprof.StopCPUProfile()
			cf.Close()
		}
		if mf != nil {
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(mf); err != nil {
				log.WithError(err).Fatal("failed to write memory profile")
			}
			mf.Close()
		}
	}
}
