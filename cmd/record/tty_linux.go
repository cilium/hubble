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

//go:build linux

package record

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

const resetSequence = "\033[A\033[2K"

// isTTY returns true if output f is a terminal
func isTTY(f *os.File) bool {
	_, err := unix.IoctlGetTermios(int(f.Fd()), unix.TCGETS)
	return err == nil
}

// resetLastLine clears the last line printed on the output f
func clearLastLine(f io.Writer) {
	fmt.Fprint(f, resetSequence)
}
