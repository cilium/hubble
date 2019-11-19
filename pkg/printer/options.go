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

package printer

import (
	"io"
	"os"
	"syscall"
)

// Output enum of the printer.
type Output int

const (
	// TabOutput prints flows in even tab-aligned columns.
	TabOutput Output = iota
	// JSONOutput prints flows as json.
	JSONOutput
	// CompactOutput prints flows as compact as possible (similar to monitor).
	CompactOutput
	// DictOutput presents the same information as TabOutput, but each flow is
	// presented as a key:value dictionary, similar to \G output of mysql.
	DictOutput
)

// Options for the printer.
type Options struct {
	output      Output
	w           io.Writer
	werr        io.Writer
	withNewLine bool
	// Use json.Encoder instead of gojay
	withJSONEncoder bool
}

// Option ...
type Option func(*Options)

// JSON encoded output from the printer.
func JSON() Option {
	return func(opts *Options) {
		opts.output = JSONOutput
	}
}

// Compact ...
func Compact() Option {
	return func(opts *Options) {
		opts.output = CompactOutput
	}
}

// Dict ...
func Dict() Option {
	return func(opts *Options) {
		opts.output = DictOutput
	}
}

// Writer sets the custom destination for where the bytes are sent.
func Writer(w io.Writer) Option {
	return func(opts *Options) {
		opts.w = w
	}
}

// WithNewLine sets the output to have a new line appended on each line.
func WithNewLine() Option {
	return func(opts *Options) {
		opts.withNewLine = true
	}
}

// WithJSONEncoder configures the JSON output to use json.Encoder instead of gojay.
func WithJSONEncoder() Option {
	return func(opts *Options) {
		opts.withJSONEncoder = true
	}
}

// IgnoreStderr configures the output to not print any
func IgnoreStderr() Option {
	return func(opts *Options) {
		opts.werr = os.NewFile(uintptr(syscall.Stderr), os.DevNull)
	}
}
