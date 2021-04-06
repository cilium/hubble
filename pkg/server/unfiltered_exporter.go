// Copyright 2021 Authors of Cilium
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

package server

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/cilium/hubble/api/v1/observer"
	"github.com/cilium/hubble/pkg/logger"
	"google.golang.org/grpc/metadata"
	"gopkg.in/natefinch/lumberjack.v2"
)

// exportServer is an implementation of pb.Observer_GetFlowsServer that simply
// JSON-encode flows.
type exportServer struct {
	encoder *json.Encoder
}

func (l *exportServer) Send(event *observer.GetFlowsResponse) error {
	if event == nil {
		return nil
	}
	switch event.ResponseTypes.(type) {
	case *observer.GetFlowsResponse_Flow:
		return l.encoder.Encode(event)
	}
	return nil
}

func (l *exportServer) SetHeader(metadata.MD) error {
	return nil
}

func (l *exportServer) SendHeader(metadata.MD) error {
	return nil
}

func (l *exportServer) SetTrailer(metadata.MD) {
}

func (l *exportServer) Context() context.Context {
	return context.Background()
}

func (l *exportServer) SendMsg(_ interface{}) error {
	return nil
}

func (l *exportServer) RecvMsg(_ interface{}) error {
	return nil
}

// ExportUnfilteredFlows calls GetFlows() with exportServer that writes flows to a rotated file.
func ExportUnfilteredFlows(server GRPCServer, config string) {
	params := strings.Split(config, ";")
	if len(params) != 3 {
		logger.GetLogger().
			WithField("config", config).
			Error("Invalid flow export config")
		return
	}
	maxSizeMB, err := strconv.Atoi(params[1])
	if err != nil || maxSizeMB <= 0 {
		logger.GetLogger().
			WithField("max-size-mb", params[1]).
			Error("Invalid value for max-size in flow export config")
		return
	}
	maxBackups, err := strconv.Atoi(params[2])
	if err != nil || maxBackups < 0 {
		logger.GetLogger().
			WithField("max-backups", params[2]).
			Error("Invalid value for max-backups in flow export config")
		return
	}
	req := observer.GetFlowsRequest{Follow: true}
	jsonEncoder := json.NewEncoder(&lumberjack.Logger{
		Filename:   params[0],
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		Compress:   true,
	})

	for {
		status, err := server.ServerStatus(context.Background(), &observer.ServerStatusRequest{})
		if err == nil && status.NumFlows > 1 {
			break
		}
		logger.GetLogger().WithError(err).Info("Waiting for Hubble server to start...")
		time.Sleep(1 * time.Second)
	}
	logger.GetLogger().WithField("config", config).Info("Starting unfiltered export")
	if err = server.GetFlows(&req, &exportServer{jsonEncoder}); err != nil {
		logger.GetLogger().WithError(err).Error("Failed to start unfiltered json export")
	}
}
