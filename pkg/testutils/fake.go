// Copyright 2019 Isovalent

package testutils

import (
	"net"

	"github.com/cilium/cilium/api/v1/models"

	pb "github.com/cilium/hubble/api/v1/flow"
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

// FakeServiceGetter is used for unit tests that need ServiceGetter.
type FakeServiceGetter struct {
	OnGetServiceByAddr func(ip net.IP, port uint16) (service pb.Service, ok bool)
	OnGetServiceByID   func(id int64) (service pb.Service, ok bool)
}

// GetServiceByAddr implements FakeServiceGetter.GetServiceByAddr.
func (f *FakeServiceGetter) GetServiceByAddr(ip net.IP, port uint16) (service pb.Service, ok bool) {
	if f.OnGetServiceByAddr != nil {
		return f.OnGetServiceByAddr(ip, port)
	}
	panic("OnGetServiceByAddr not set")
}

// GetServiceByID implements FakeServiceGetter.GetServiceByID.
func (f *FakeServiceGetter) GetServiceByID(id int64) (service pb.Service, ok bool) {
	if f.OnGetServiceByID != nil {
		return f.OnGetServiceByID(id)
	}
	panic("OnGetServiceByID not set")
}

// NoopServiceGetter always returns an empty response.
var NoopServiceGetter = FakeServiceGetter{
	OnGetServiceByAddr: func(ip net.IP, port uint16) (service pb.Service, ok bool) {
		return pb.Service{}, false
	},
	OnGetServiceByID: func(id int64) (service pb.Service, ok bool) {
		return pb.Service{}, false
	},
}

// FakeIdentityGetter is used for unit tests that need IdentityGetter.
type FakeIdentityGetter struct {
	OnGetIdentity func(securityIdentity uint64) (*models.Identity, error)
}

// GetIdentity implements IdentityGetter.GetIPIdentity.
func (f *FakeIdentityGetter) GetIdentity(securityIdentity uint64) (*models.Identity, error) {
	if f.OnGetIdentity != nil {
		return f.OnGetIdentity(securityIdentity)
	}
	panic("OnGetIdentity not set")
}

// NoopIdentityGetter always returns an empty response.
var NoopIdentityGetter = FakeIdentityGetter{
	OnGetIdentity: func(securityIdentity uint64) (*models.Identity, error) {
		return &models.Identity{}, nil
	},
}

// FakeEndpointsHandler implements EndpointsHandler interface for unit testing.
type FakeEndpointsHandler struct {
	FakeSyncEndpoints            func([]*v1.Endpoint)
	FakeUpdateEndpoint           func(*v1.Endpoint)
	FakeMarkDeleted              func(*v1.Endpoint)
	FakeFindEPs                  func(epID uint64, ns, pod string) []v1.Endpoint
	FakeGetEndpoint              func(ip net.IP) (endpoint *v1.Endpoint, ok bool)
	FakeGarbageCollect           func()
	FakeGetEndpointByContainerID func(id string) (endpoint *v1.Endpoint, ok bool)
	FakeGetEndpointByPodName     func(namespace string, name string) (*v1.Endpoint, bool)
}

// SyncEndpoints calls FakeSyncEndpoints.
func (f *FakeEndpointsHandler) SyncEndpoints(eps []*v1.Endpoint) {
	if f.FakeSyncEndpoints != nil {
		f.FakeSyncEndpoints(eps)
		return
	}
	panic("SyncEndpoints([]*v1.Endpoint) should not have been called since it was not defined")
}

// UpdateEndpoint calls FakeUpdateEndpoint.
func (f *FakeEndpointsHandler) UpdateEndpoint(ep *v1.Endpoint) {
	if f.FakeUpdateEndpoint != nil {
		f.FakeUpdateEndpoint(ep)
		return
	}
	panic("UpdateEndpoint(*v1.Endpoint) should not have been called since it was not defined")
}

// MarkDeleted calls FakeMarkDeleted.
func (f *FakeEndpointsHandler) MarkDeleted(ep *v1.Endpoint) {
	if f.FakeMarkDeleted != nil {
		f.FakeMarkDeleted(ep)
		return
	}
	panic("MarkDeleted(ep *v1.Endpoint) should not have been called since it was not defined")
}

// FindEPs calls FakeFindEPs.
func (f *FakeEndpointsHandler) FindEPs(epID uint64, ns, pod string) []v1.Endpoint {
	if f.FakeFindEPs != nil {
		return f.FakeFindEPs(epID, ns, pod)
	}
	panic(" FindEPs(epID uint64, ns, pod string) should not have been called since it was not defined")
}

// GetEndpoint calls FakeGetEndpoint.
func (f *FakeEndpointsHandler) GetEndpoint(ip net.IP) (ep *v1.Endpoint, ok bool) {
	if f.FakeGetEndpoint != nil {
		return f.FakeGetEndpoint(ip)
	}
	panic("GetEndpoint(ip net.IP) (ep *v1.Endpoint, ok bool) should not have been called since it was not defined")
}

// GetEndpointByContainerID calls FakeGetEndpointByContainerID.
func (f *FakeEndpointsHandler) GetEndpointByContainerID(id string) (ep *v1.Endpoint, ok bool) {
	if f.FakeGetEndpointByContainerID != nil {
		return f.FakeGetEndpointByContainerID(id)
	}
	panic("GetEndpointByContainerID(id string) (ep *v1.Endpoint, ok bool) should not have been called since it was not defined")
}

// GetEndpointByPodName calls FakeGetEndpointByPodName.
func (f *FakeEndpointsHandler) GetEndpointByPodName(namespace string, name string) (ep *v1.Endpoint, ok bool) {
	if f.FakeGetEndpointByPodName != nil {
		return f.FakeGetEndpointByPodName(namespace, name)
	}
	panic("GetEndpointByPodName(namespace string, name string) (ep *v1.Endpoint, ok bool) should not have been called since it was not defined")
}

// GarbageCollect calls FakeGarbageCollect.
func (f *FakeEndpointsHandler) GarbageCollect() {
	if f.FakeGarbageCollect != nil {
		f.FakeGarbageCollect()
		return
	}
	panic("GarbageCollect() should not have been called since it was not defined")
}

// FakeCiliumClient implements CliliumClient interface for unit testing.
type FakeCiliumClient struct {
	FakeEndpointList    func() ([]*models.Endpoint, error)
	FakeGetEndpoint     func(uint64) (*models.Endpoint, error)
	FakeGetIdentity     func(uint64) (*models.Identity, error)
	FakeGetFqdnCache    func() ([]*models.DNSLookup, error)
	FakeGetIPCache      func() ([]*models.IPListEntry, error)
	FakeGetServiceCache func() ([]*models.Service, error)
}

// EndpointList calls FakeEndpointList.
func (c *FakeCiliumClient) EndpointList() ([]*models.Endpoint, error) {
	if c.FakeEndpointList != nil {
		return c.FakeEndpointList()
	}
	panic("EndpointList() should not have been called since it was not defined")
}

// GetEndpoint calls FakeGetEndpoint.
func (c *FakeCiliumClient) GetEndpoint(id uint64) (*models.Endpoint, error) {
	if c.FakeGetEndpoint != nil {
		return c.FakeGetEndpoint(id)
	}
	panic("GetEndpoint(uint64) should not have been called since it was not defined")
}

// GetIdentity calls FakeGetIdentity.
func (c *FakeCiliumClient) GetIdentity(id uint64) (*models.Identity, error) {
	if c.FakeGetIdentity != nil {
		return c.FakeGetIdentity(id)
	}
	panic("GetIdentity(uint64) should not have been called since it was not defined")
}

// GetFqdnCache calls FakeGetFqdnCache.
func (c *FakeCiliumClient) GetFqdnCache() ([]*models.DNSLookup, error) {
	if c.FakeGetFqdnCache != nil {
		return c.FakeGetFqdnCache()
	}
	panic("GetFqdnCache() should not have been called since it was not defined")
}

// GetIPCache calls FakeGetIPCache.
func (c *FakeCiliumClient) GetIPCache() ([]*models.IPListEntry, error) {
	if c.FakeGetIPCache != nil {
		return c.FakeGetIPCache()
	}
	panic("GetIPCache() should not have been called since it was not defined")
}

// GetServiceCache calls FakeGetServiceCache.
func (c *FakeCiliumClient) GetServiceCache() ([]*models.Service, error) {
	if c.FakeGetServiceCache != nil {
		return c.FakeGetServiceCache()
	}
	panic("GetServiceCache() should not have been called since it was not defined")
}
