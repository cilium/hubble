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
	"os"
	"runtime"
	"runtime/pprof"
)

var (
	cpuprofile, memprofile         string
	cpuprofileFile, memprofileFile *os.File
)

func pprofInit() error {
	var err error
	if cpuprofile != "" {
		cpuprofileFile, err = os.Create(cpuprofile)
		if err != nil {
			return fmt.Errorf("failed to create cpu profile: %v", err)
		}
		pprof.StartCPUProfile(cpuprofileFile)
	}
	if memprofile != "" {
		memprofileFile, err = os.Create(memprofile)
		if err != nil {
			return fmt.Errorf("failed to create memory profile: %v", err)
		}
	}
	return nil
}

func pprofTearDown() error {
	if cpuprofileFile != nil {
		pprof.StopCPUProfile()
		cpuprofileFile.Close()
	}
	if memprofileFile != nil {
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(memprofileFile); err != nil {
			return fmt.Errorf("failed to write memory profile: %v", err)
		}
		memprofileFile.Close()
	}
	return nil
}
