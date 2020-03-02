// Copyright 2020 Authors of Hubble
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

package k8s

import (
	"context"
	"net"
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// Client is a high level clienet for the Kubernetes API.
type Client struct {
	client rest.Interface
}

// NewClient creates a new Client using the provided config.
func NewClient(config *rest.Config) (*Client, error) {
	client, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

// GetServiceEndpoints returns a map of endpoints of service in namespace.
// This map has node names as keys and endpoint addresses as values.
// An address is in the form "host:port" for IPv4 addresses and "[host]:port"
// for IPv6 addresses.
func (c *Client) GetServiceEndpoints(ctx context.Context, namespace, service string) (map[string][]string, error) {
	endpoints := &v1.Endpoints{}
	err := c.client.Get().
		Context(ctx).
		Namespace(namespace).
		Resource("endpoints").
		Name(service).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do().
		Into(endpoints)
	if err != nil {
		return nil, err
	}
	eps := make(map[string][]string)
	for _, epSubset := range endpoints.Subsets {
		for _, addr := range epSubset.Addresses {
			if addr.NodeName == nil {
				continue
			}
			for _, port := range epSubset.Ports {
				if port.Port == 0 {
					continue
				}
				eps[*addr.NodeName] = append(eps[*addr.NodeName], net.JoinHostPort(addr.IP, strconv.Itoa(int(port.Port))))
			}
		}
	}
	return eps, nil
}
