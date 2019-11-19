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
	"net"
	"time"

	"github.com/cilium/cilium/pkg/identity"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
	"go.uber.org/zap"

	v1 "github.com/cilium/hubble/pkg/api/v1"
	"github.com/cilium/hubble/pkg/cilium/client"
)

const (
	ipcacheInitRetryInterval = 5 * time.Second
	ipcacheRefreshInterval   = 5 * time.Minute
)

type podGetter interface {
	GetPodNameOf(ip net.IP) (ns, pod string, ok bool)
}

type endpointGetter interface {
	GetEndpoint(ip net.IP) (ep *v1.Endpoint, ok bool)
}

// LegacyPodGetter implements GetPodNameOf based on the IPCache-backed
// podGetter, but falls back on obtaining the pod information from the list
// of endpoints. This is intended to support Cilium 1.6 and older.
type LegacyPodGetter struct {
	PodGetter      podGetter
	EndpointGetter endpointGetter
}

// GetPodNameOf fetches K8s pod and namespace information.
func (l *LegacyPodGetter) GetPodNameOf(ip net.IP) (ns, pod string, ok bool) {
	if ns, pod, ok := l.PodGetter.GetPodNameOf(ip); ok {
		return ns, pod, ok
	}

	// fallback on local endpoints
	if ep, ok := l.EndpointGetter.GetEndpoint(ip); ok {
		return ep.PodNamespace, ep.PodName, ok
	}

	return "", "", false
}

// fetchIPCache copies over the IP cache from cilium agent
func (s *ObserverServer) fetchIPCache() error {
	entries, err := s.ciliumClient.GetIPCache()
	if err != nil {
		return err
	}
	err = s.ipcache.InitializeFrom(entries)
	if err != nil {
		return err
	}
	s.log.Debug("Fetched ipcache from cilium", zap.Int("entries", len(entries)))
	return nil
}

// processIPCacheEvent decodes and applies an IPCache update, returns true if
// it was applied to the local IPCache mirror.
func (s *ObserverServer) processIPCacheEvent(an monitorAPI.AgentNotify) bool {
	n := monitorAPI.IPCacheNotification{}
	err := json.Unmarshal([]byte(an.Text), &n)
	if err != nil {
		s.log.Error("Unable to unmarshal IPCacheNotification",
			zap.Int("type", int(an.Type)), zap.String("IPCacheNotification", an.Text))
		return false
	}

	switch an.Type {
	case monitorAPI.AgentNotifyIPCacheUpserted:
		newID := identity.NumericIdentity(n.Identity)
		var oldID *identity.NumericIdentity
		if n.OldIdentity != nil {
			id := identity.NumericIdentity(*n.OldIdentity)
			oldID = &id
		}

		return s.ipcache.UpsertChecked(n.CIDR, newID, oldID, n.HostIP, n.OldHostIP,
			n.EncryptKey, n.Namespace, n.PodName)
	case monitorAPI.AgentNotifyIPCacheDeleted:
		return s.ipcache.Delete(n.CIDR)
	default:
		s.log.Warn("Received unknown IPCache notification type", zap.Int("type", int(an.Type)))
	}

	return false
}

// syncIPCache initializes the IPCache by fetching an initial version from
// Cilium and then starts reading IPCacheNotification from the channel.
func (s *ObserverServer) syncIPCache(ipcacheEvents <-chan monitorAPI.AgentNotify) {
	for {
		err := s.fetchIPCache()
		if err != nil {
			// This is expected to fail on older versions of cilium, therefore
			// we emit a warning and will not try to synchronize the ipcache.
			if client.IsIPCacheNotFoundErr(err) {
				s.log.Warn("Failed to obtain IPCache from Cilium. If you are using Cilium 1.6 or older, " +
					"this is expected. Pod names of endpoints running on remote nodes will not be resolved.")
				return
			}
			s.log.Error("Failed to fetch IPCache from Cilium", zap.Error(err))
			time.Sleep(ipcacheInitRetryInterval)
			continue
		}

		break
	}

	refresh := time.NewTimer(ipcacheRefreshInterval)
	inSync := false

	for ipcacheEvents != nil {
		select {
		case <-refresh.C:
			err := s.fetchIPCache()
			if err != nil {
				s.log.Error("Failed to fetch IPCache from Cilium", zap.Error(err))
				refresh.Reset(ipcacheInitRetryInterval)
				continue
			}
			refresh.Reset(ipcacheRefreshInterval)
		case an, ok := <-ipcacheEvents:
			if !ok {
				return
			}
			// Initially we might see stale updates that were enqued before we
			// initialized the ipcache. Once we see the first applicable update
			// though, all subsequent updates must be applicable as well.
			updated := s.processIPCacheEvent(an)
			switch {
			case !updated && !inSync:
				s.log.Debug("Received stale ipcache update", zap.Int("type", int(an.Type)), zap.String("IPCacheNotification", an.Text))
			case !updated && inSync:
				s.log.Warn("Received unapplicable ipcache update", zap.Int("type", int(an.Type)), zap.String("IPCacheNotification", an.Text))
			case updated && !inSync:
				inSync = true
			}
		}
	}
}
