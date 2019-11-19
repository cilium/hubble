// Copyright 2019 Isovalent

package testutils

import (
	"net"

	v1 "github.com/cilium/hubble/pkg/api/v1"
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

// FakeK8sGetter is used for unit tests that needs K8sGetter.
type FakeK8sGetter struct {
	OnGetPodNameOf func(ip net.IP) (ns, pod string, ok bool)
}

// GetPodNameOf implements K8sGetter.GetPodNameOf.
func (f *FakeK8sGetter) GetPodNameOf(ip net.IP) (ns, pod string, ok bool) {
	if f.OnGetPodNameOf != nil {
		return f.OnGetPodNameOf(ip)
	}
	panic("OnGetPodNameOf not set")
}

// NoopK8sGetter always returns an empty response.
var NoopK8sGetter = FakeK8sGetter{
	OnGetPodNameOf: func(ip net.IP) (ns, pod string, ok bool) {
		return "", "", false
	},
}
