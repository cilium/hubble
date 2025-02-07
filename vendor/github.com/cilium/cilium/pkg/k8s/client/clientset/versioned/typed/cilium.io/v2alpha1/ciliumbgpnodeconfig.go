// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

// Code generated by client-gen. DO NOT EDIT.

package v2alpha1

import (
	context "context"

	ciliumiov2alpha1 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2alpha1"
	scheme "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// CiliumBGPNodeConfigsGetter has a method to return a CiliumBGPNodeConfigInterface.
// A group's client should implement this interface.
type CiliumBGPNodeConfigsGetter interface {
	CiliumBGPNodeConfigs() CiliumBGPNodeConfigInterface
}

// CiliumBGPNodeConfigInterface has methods to work with CiliumBGPNodeConfig resources.
type CiliumBGPNodeConfigInterface interface {
	Create(ctx context.Context, ciliumBGPNodeConfig *ciliumiov2alpha1.CiliumBGPNodeConfig, opts v1.CreateOptions) (*ciliumiov2alpha1.CiliumBGPNodeConfig, error)
	Update(ctx context.Context, ciliumBGPNodeConfig *ciliumiov2alpha1.CiliumBGPNodeConfig, opts v1.UpdateOptions) (*ciliumiov2alpha1.CiliumBGPNodeConfig, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, ciliumBGPNodeConfig *ciliumiov2alpha1.CiliumBGPNodeConfig, opts v1.UpdateOptions) (*ciliumiov2alpha1.CiliumBGPNodeConfig, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*ciliumiov2alpha1.CiliumBGPNodeConfig, error)
	List(ctx context.Context, opts v1.ListOptions) (*ciliumiov2alpha1.CiliumBGPNodeConfigList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *ciliumiov2alpha1.CiliumBGPNodeConfig, err error)
	CiliumBGPNodeConfigExpansion
}

// ciliumBGPNodeConfigs implements CiliumBGPNodeConfigInterface
type ciliumBGPNodeConfigs struct {
	*gentype.ClientWithList[*ciliumiov2alpha1.CiliumBGPNodeConfig, *ciliumiov2alpha1.CiliumBGPNodeConfigList]
}

// newCiliumBGPNodeConfigs returns a CiliumBGPNodeConfigs
func newCiliumBGPNodeConfigs(c *CiliumV2alpha1Client) *ciliumBGPNodeConfigs {
	return &ciliumBGPNodeConfigs{
		gentype.NewClientWithList[*ciliumiov2alpha1.CiliumBGPNodeConfig, *ciliumiov2alpha1.CiliumBGPNodeConfigList](
			"ciliumbgpnodeconfigs",
			c.RESTClient(),
			scheme.ParameterCodec,
			"",
			func() *ciliumiov2alpha1.CiliumBGPNodeConfig { return &ciliumiov2alpha1.CiliumBGPNodeConfig{} },
			func() *ciliumiov2alpha1.CiliumBGPNodeConfigList { return &ciliumiov2alpha1.CiliumBGPNodeConfigList{} },
		),
	}
}
