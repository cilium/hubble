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

package conn

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cilium/hubble/pkg/defaults"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// GRPCOptionFunc is a function that configures a gRPC dial option.
type GRPCOptionFunc func(vp *viper.Viper) (grpc.DialOption, error)

// GRPCOptionFuncs is a combination of multiple gRPC dial option.
var GRPCOptionFuncs []GRPCOptionFunc

func init() {
	GRPCOptionFuncs = append(
		GRPCOptionFuncs,
		grpcOptionBlock,
		grpcOptionFailOnNonTempDialError,
		grpcOptionConnError,
	)
}

func grpcOptionBlock(_ *viper.Viper) (grpc.DialOption, error) {
	return grpc.WithBlock(), nil
}

func grpcOptionFailOnNonTempDialError(_ *viper.Viper) (grpc.DialOption, error) {
	return grpc.FailOnNonTempDialError(true), nil
}

func grpcOptionConnError(_ *viper.Viper) (grpc.DialOption, error) {
	return grpc.WithReturnConnectionError(), nil
}

var grpcDialOptions []grpc.DialOption

// Init initializes common connection options. It MUST be called prior to any
// other package functions.
func Init(vp *viper.Viper) error {
	for _, fn := range GRPCOptionFuncs {
		dialOpt, err := fn(vp)
		if err != nil {
			return err
		}
		grpcDialOptions = append(grpcDialOptions, dialOpt)
	}
	return nil
}

// New creates a new gRPC client connection to the target.
func New(ctx context.Context, target string, dialTimeout time.Duration) (*grpc.ClientConn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()

	t := strings.TrimPrefix(target, defaults.TargetTLSPrefix)
	conn, err := grpc.DialContext(dialCtx, t, grpcDialOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to '%s': %w", target, err)
	}
	return conn, nil
}
