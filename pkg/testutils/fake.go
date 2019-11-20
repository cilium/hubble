// Copyright 2019 Isovalent

package testutils

import (
	"net"

	v1 "github.com/cilium/hubble/pkg/api/v1"
	"github.com/cilium/hubble/pkg/ipcache"
)

// FakeDNSGetter is used for unit tests that needs DNSGetter.
type FakeDNSGetter struct {
	OnGetNamesOf func(sourceEpID uint64, ip net.IP) (names []string)
}

// GetNamesOf implements DNSGetter.GetNameOf.
func (f *FakeDNSGetter) GetNamesOf(sourceEpID uint64, ip net.IP) (fqdns []string) {
	if f.OnGetNamesOf != nil {
		return f.OnGetNamesOf(sourceEpID, ip)
	}
	panic("OnGetNamesOf not set")
}

// NoopDNSGetter always returns an empty response.
var NoopDNSGetter = FakeDNSGetter{
	OnGetNamesOf: func(sourceEpID uint64, ip net.IP) (fqdns []string) {
		return nil
	},
}

// FakeEndpointGetter is used for unit tests that needs EndpointGetter.
type FakeEndpointGetter struct {
	OnGetEndpoint func(ip net.IP) (endpoint *v1.Endpoint, ok bool)
}

// GetEndpoint implements EndpointGetter.GetEndpoint.
func (f *FakeEndpointGetter) GetEndpoint(ip net.IP) (endpoint *v1.Endpoint, ok bool) {
	if f.OnGetEndpoint != nil {
		return f.OnGetEndpoint(ip)
	}
	panic("OnGetEndpoint not set")
}

// NoopEndpointGetter always returns an empty response.
var NoopEndpointGetter = FakeEndpointGetter{
	OnGetEndpoint: func(ip net.IP) (endpoint *v1.Endpoint, ok bool) {
		return nil, false
	},
}

// FakeIPGetter is used for unit tests that needs IPGetter.
type FakeIPGetter struct {
	OnGetIPIdentity func(ip net.IP) (id ipcache.IPIdentity, ok bool)
}

// GetIPIdentity implements FakeIPGetter.GetIPIdentity.
func (f *FakeIPGetter) GetIPIdentity(ip net.IP) (id ipcache.IPIdentity, ok bool) {
	if f.OnGetIPIdentity != nil {
		return f.OnGetIPIdentity(ip)
	}
	panic("OnGetIPIdentity not set")
}

// NoopIPGetter always returns an empty response.
var NoopIPGetter = FakeIPGetter{
	OnGetIPIdentity: func(ip net.IP) (id ipcache.IPIdentity, ok bool) {
		return ipcache.IPIdentity{}, false
	},
}
