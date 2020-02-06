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

package cilium

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
func (s *State) syncEndpoints() {
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
		s.endpoints.GarbageCollect()
	}
}

func (s *State) consumeEndpointEvents() {
	for an := range s.endpointEvents {
		switch an.Type {
		case monitorAPI.AgentNotifyEndpointCreated, monitorAPI.AgentNotifyEndpointRegenerateSuccess:
			// When a new endpoint is created, or an endpoint is successfully
			// updated, we consult the Cilium API to fetch additional endpoint
			// information such as the endpoint IP address.
			ern := monitorAPI.EndpointRegenNotification{}
			err := json.Unmarshal([]byte(an.Text), &ern)
			if err != nil {
				s.log.Error("Unable to unmarshal EndpointRegenNotification", zap.String("EndpointRegenNotification", an.Text))
				continue
			}

			ciliumEP, err := s.ciliumClient.GetEndpoint(ern.ID)
			if err != nil {
				s.log.Error("Updated or created endpoint not found!", zap.Uint64("id", ern.ID), zap.Error(err))
				continue
			}
			ep := endpoint.ParseEndpointFromModel(ciliumEP)
			s.endpoints.UpdateEndpoint(ep)
		case monitorAPI.AgentNotifyEndpointDeleted:
			// When a deleted endpoint is found in the local endpoint cache,
			// sets the time when the endpoint was deleted. If not found, stores
			// a new endpoint in the cache, as well with the time when the
			// endpoint was deleted.
			edn := monitorAPI.EndpointDeleteNotification{}
			err := json.Unmarshal([]byte(an.Text), &edn)
			if err != nil {
				s.log.Error("Unable to unmarshal EndpointDeleteNotification", zap.String("EndpointDeleteNotification", an.Text))
				continue
			}

			ep := endpoint.ParseEndpointFromEndpointDeleteNotification(edn)
			s.endpoints.MarkDeleted(ep)
		default:
			s.log.Debug("Ignoring unknown endpoint event",
				zap.Int("type", int(an.Type)), zap.String("notification", an.Text))
		}
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
