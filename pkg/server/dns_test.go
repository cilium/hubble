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
	"net"
	"sync"
	"testing"
	"time"

	"github.com/cilium/cilium/api/v1/models"
	monitorAPI "github.com/cilium/cilium/pkg/monitor"
	"github.com/cilium/cilium/pkg/proxy/accesslog"
	"github.com/cilium/cilium/pkg/u8proto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type fakeFQDNCache struct {
	fakeInitializeFrom func(entries []*models.DNSLookup)
	fakeAddDNSLookup   func(epID uint64, lookupTime time.Time, domainName string, ips []net.IP, ttl uint32)
	fakeGetNamesOf     func(epID uint64, ip net.IP) []string
}

func (f *fakeFQDNCache) InitializeFrom(entries []*models.DNSLookup) {
	if f.fakeInitializeFrom != nil {
		f.fakeInitializeFrom(entries)
		return
	}
	panic("InitializeFrom([]*models.DNSLookup) should not have been called since it was not defined")
}

func (f *fakeFQDNCache) AddDNSLookup(epID uint64, lookupTime time.Time, domainName string, ips []net.IP, ttl uint32) {
	if f.fakeAddDNSLookup != nil {
		f.fakeAddDNSLookup(epID, lookupTime, domainName, ips, ttl)
		return
	}
	panic("AddDNSLookup(uint64, time.Time, string, []net.IP, uint32) should not have been called since it was not defined")
}

func (f *fakeFQDNCache) GetNamesOf(epID uint64, ip net.IP) []string {
	if f.fakeGetNamesOf != nil {
		return f.fakeGetNamesOf(epID, ip)
	}
	panic("GetNamesOf(uint64, net.IP) should not have been called since it was not defined")
}

func TestObserverServer_consumeLogRecordNotifyChannel(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	lr := monitorAPI.LogRecordNotify{
		LogRecord: accesslog.LogRecord{
			Type:             accesslog.TypeResponse,
			Timestamp:        "2006-01-02T15:04:05.999999999Z",
			ObservationPoint: accesslog.Ingress,
			SourceEndpoint: accesslog.EndpointInfo{
				ID:           123,
				IPv4:         "",
				IPv6:         "",
				Port:         0,
				Identity:     0,
				Labels:       nil,
				LabelsSHA256: "",
			},
			IPVersion:         accesslog.VersionIPV6,
			Verdict:           accesslog.VerdictForwarded,
			TransportProtocol: accesslog.TransportProtocol(u8proto.UDP),
			ServiceInfo:       nil,
			DropReason:        nil,
			DNS: &accesslog.LogRecordDNS{
				Query:             "deathstar.empire.svc.cluster.local.",
				IPs:               []net.IP{net.ParseIP("1.2.3.4")},
				TTL:               5,
				ObservationSource: accesslog.DNSSourceProxy,
				RCode:             0,
				QTypes:            []uint16{1},
			},
		},
	}
	fakeFQDNCache := &fakeFQDNCache{
		fakeAddDNSLookup: func(epID uint64, lookupTime time.Time, domainName string, ips []net.IP, ttl uint32) {
			defer wg.Done()
			assert.Equal(t, uint64(123), epID)
			assert.Equal(t, []net.IP{net.ParseIP("1.2.3.4")}, ips)
			assert.Equal(t, "deathstar.empire.svc.cluster.local.", domainName)
		},
	}

	s := &ObserverServer{
		fqdnCache: fakeFQDNCache,
		logRecord: make(chan monitorAPI.LogRecordNotify, 1),
		log:       zap.L(),
	}
	go s.consumeLogRecordNotifyChannel()

	s.getLogRecordNotifyChannel() <- lr
	wg.Wait()
}
