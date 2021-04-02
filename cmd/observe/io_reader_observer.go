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

package observe

import (
	"bufio"
	"context"
	"io"

	"github.com/cilium/cilium/api/v1/observer"
	v1 "github.com/cilium/cilium/pkg/hubble/api/v1"
	"github.com/cilium/cilium/pkg/hubble/filters"
	"github.com/cilium/hubble/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// ioReaderObserver implements ObserverClient interface. It reads flows
// in jsonpb format from an io.Reader.
type ioReaderObserver struct {
	scanner *bufio.Scanner
}

func newIOReaderObserver(reader io.Reader) *ioReaderObserver {
	return &ioReaderObserver{
		scanner: bufio.NewScanner(reader),
	}
}

func (o *ioReaderObserver) GetFlows(_ context.Context, in *observer.GetFlowsRequest, _ ...grpc.CallOption) (observer.Observer_GetFlowsClient, error) {
	return newIOReaderClient(o.scanner, in)
}

func (o *ioReaderObserver) GetNodes(_ context.Context, _ *observer.GetNodesRequest, _ ...grpc.CallOption) (*observer.GetNodesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "GetNodes not implemented")
}

func (o *ioReaderObserver) ServerStatus(_ context.Context, _ *observer.ServerStatusRequest, _ ...grpc.CallOption) (*observer.ServerStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ServerStatus not implemented")
}

// ioReaderClient implements Observer_GetFlowsClient.
type ioReaderClient struct {
	scanner *bufio.Scanner
	request *observer.GetFlowsRequest
	allow   filters.FilterFuncs
	deny    filters.FilterFuncs
	grpc.ClientStream
}

func newIOReaderClient(scanner *bufio.Scanner, request *observer.GetFlowsRequest) (*ioReaderClient, error) {
	allow, err := filters.BuildFilterList(context.Background(), request.Whitelist, filters.DefaultFilters)
	if err != nil {
		return nil, err
	}
	deny, err := filters.BuildFilterList(context.Background(), request.Blacklist, filters.DefaultFilters)
	if err != nil {
		return nil, err
	}
	return &ioReaderClient{
		scanner: scanner,
		request: request,
		allow:   allow,
		deny:    deny,
	}, nil
}

func (c *ioReaderClient) Recv() (*observer.GetFlowsResponse, error) {
	for c.scanner.Scan() {
		line := c.scanner.Text()
		var res observer.GetFlowsResponse
		err := protojson.Unmarshal(c.scanner.Bytes(), &res)
		if err != nil {
			logger.Logger.WithError(err).WithField("line", line).Warn("Failed to unmarshal json to flow")
			continue
		}
		if c.request.Since != nil && c.request.Since.AsTime().After(res.Time.AsTime()) {
			continue
		}
		if c.request.Until != nil && c.request.Until.AsTime().Before(res.Time.AsTime()) {
			continue
		}
		if !filters.Apply(c.allow, c.deny, &v1.Event{Timestamp: res.Time, Event: res.GetFlow()}) {
			continue
		}
		return &res, nil
	}
	if err := c.scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}
