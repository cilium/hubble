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

package server

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/cilium/cilium/api/v1/models"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	"go.uber.org/zap"

	v1 "github.com/cilium/hubble/pkg/api/v1"
	"github.com/cilium/hubble/pkg/parser/endpoint"
)

var (
	// refreshEndpointList is the time hubble will refresh current endpoints
	// with cilium's
	refreshEndpointList = time.Minute
)

// syncEndpoints sync all endpoints of Cilium with the hubble.
func (s *ObserverServer) syncEndpoints() {
	for {
		eps, err := s.ciliumClient.EndpointList()
		if err != nil {
			s.log.Error("Unable to get cilium endpoint list", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		for _, modelUpdateEP := range eps {
			updatedEp := endpoint.ParseEndpointFromModel(modelUpdateEP)
			s.log.Debug("Found pod", zap.String("namespace", updatedEp.PodNamespace), zap.String("pod-name", updatedEp.PodName))
			s.endpoints.UpdateEndpoint(updatedEp)
		}
		break
	}
	for {
		time.Sleep(refreshEndpointList)
		eps, err := s.ciliumClient.EndpointList()
		if err != nil {
			s.log.Error("Unable to get cilium endpoint list", zap.Error(err))
			continue
		}
		var parsedEPs []*v1.Endpoint
		for _, modelUpdateEP := range eps {
			parsedEPs = append(parsedEPs, endpoint.ParseEndpointFromModel(modelUpdateEP))
		}

		s.endpoints.SyncEndpoints(parsedEPs)
	}
}

// consumeEpAddEvents starts reading the s.epAdd channel and adds the endpoint
// to the observerServer's endpoints.
func (s *ObserverServer) consumeEpAddEvents() {
	for ep := range s.epAdd {
		ecn := monitorAPI.EndpointCreateNotification{}
		err := json.Unmarshal([]byte(ep), &ecn)
		if err != nil {
			s.log.Error("Unable to unmarshal EndpointCreateNotification", zap.String("EndpointCreateNotification", ep))
			continue
		}

		ciliumEP, err := s.ciliumClient.GetEndpoint(ecn.ID)
		if err != nil {
			s.log.Error("Endpoint not found!", zap.Error(err))
			continue
		}
		ep := endpoint.ParseEndpointFromModel(ciliumEP)
		s.endpoints.UpdateEndpoint(ep)
	}
}

// consumeEpAddEvents starts reading the s.epDel channel and, if found in
// observerServer, sets the time when the endpoint was deleted, if not found
// stores a new endpoint in the observerServer as well with the time when the
// endpoint was deleted.
func (s *ObserverServer) consumeEpDelEvents() {
	for epDeleted := range s.epDel {
		edn := monitorAPI.EndpointDeleteNotification{}
		err := json.Unmarshal([]byte(epDeleted), &edn)
		if err != nil {
			s.log.Error("Unable to unmarshal EndpointDeleteNotification", zap.String("EndpointDeleteNotification", epDeleted))
			continue
		}

		ep := endpoint.ParseEndpointFromEndpointDeleteNotification(edn)
		s.endpoints.MarkDeleted(ep)
	}
}

// GetNamespace returns the namespace the Endpoint belongs to.
func GetNamespace(ep *models.Endpoint) string {
	if ep.Status != nil && ep.Status.Identity != nil {
		for _, label := range ep.Status.Identity.Labels {
			kv := strings.Split(label, "=")
			if len(kv) == 2 && kv[0] == v1.K8sNamespaceTag {
				return kv[1]
			}
		}
	}
	return ""
}
