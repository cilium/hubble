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

package common

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// NewHubbleConn creates a new gRPC client connection to the configured Hubble
// target.
func NewHubbleConn(ctx context.Context, vp *viper.Viper) (*grpc.ClientConn, error) {
	target := vp.GetString("server")
	dialCtx, cancel := context.WithTimeout(ctx, vp.GetDuration("timeout"))
	defer cancel()
	conn, err := grpc.DialContext(dialCtx, target, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to '%s': %w", target, err)
	}
	return conn, nil
}
