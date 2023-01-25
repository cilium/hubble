// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

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

// IOReaderObserver implements ObserverClient interface. It reads flows
// in jsonpb format from an io.Reader.
type IOReaderObserver struct {
	scanner *bufio.Scanner
}

// NewIOReaderObserver reads flows in jsonpb format from an io.Reader and
// returns a IOReaderObserver that implements the ObserverClient interface.
func NewIOReaderObserver(reader io.Reader) *IOReaderObserver {
	return &IOReaderObserver{
		scanner: bufio.NewScanner(reader),
	}
}

// GetFlows returns flows
func (o *IOReaderObserver) GetFlows(ctx context.Context, in *observer.GetFlowsRequest, _ ...grpc.CallOption) (observer.Observer_GetFlowsClient, error) {
	return newIOReaderClient(ctx, o.scanner, in)
}

// GetAgentEvents is not implemented, and will throw an error if used.
func (o *IOReaderObserver) GetAgentEvents(_ context.Context, _ *observer.GetAgentEventsRequest, _ ...grpc.CallOption) (observer.Observer_GetAgentEventsClient, error) {
	return nil, status.Errorf(codes.Unimplemented, "GetAgentEvents not implemented")
}

// GetDebugEvents is not implemented, and will throw an error if used.
func (o *IOReaderObserver) GetDebugEvents(_ context.Context, _ *observer.GetDebugEventsRequest, _ ...grpc.CallOption) (observer.Observer_GetDebugEventsClient, error) {
	return nil, status.Errorf(codes.Unimplemented, "GetDebugEvents not implemented")
}

// GetNodes is not implemented, and will throw an error if used.
func (o *IOReaderObserver) GetNodes(_ context.Context, _ *observer.GetNodesRequest, _ ...grpc.CallOption) (*observer.GetNodesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "GetNodes not implemented")
}

// ServerStatus is not implemented, and will throw an error if used.
func (o *IOReaderObserver) ServerStatus(_ context.Context, _ *observer.ServerStatusRequest, _ ...grpc.CallOption) (*observer.ServerStatusResponse, error) {
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

func newIOReaderClient(ctx context.Context, scanner *bufio.Scanner, request *observer.GetFlowsRequest) (*ioReaderClient, error) {
	allow, err := filters.BuildFilterList(ctx, request.Whitelist, filters.DefaultFilters)
	if err != nil {
		return nil, err
	}
	deny, err := filters.BuildFilterList(ctx, request.Blacklist, filters.DefaultFilters)
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
