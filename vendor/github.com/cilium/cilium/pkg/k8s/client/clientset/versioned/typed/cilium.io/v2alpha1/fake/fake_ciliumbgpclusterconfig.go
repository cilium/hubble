// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v2alpha1 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2alpha1"
	ciliumiov2alpha1 "github.com/cilium/cilium/pkg/k8s/client/clientset/versioned/typed/cilium.io/v2alpha1"
	gentype "k8s.io/client-go/gentype"
)

// fakeCiliumBGPClusterConfigs implements CiliumBGPClusterConfigInterface
type fakeCiliumBGPClusterConfigs struct {
	*gentype.FakeClientWithList[*v2alpha1.CiliumBGPClusterConfig, *v2alpha1.CiliumBGPClusterConfigList]
	Fake *FakeCiliumV2alpha1
}

func newFakeCiliumBGPClusterConfigs(fake *FakeCiliumV2alpha1) ciliumiov2alpha1.CiliumBGPClusterConfigInterface {
	return &fakeCiliumBGPClusterConfigs{
		gentype.NewFakeClientWithList[*v2alpha1.CiliumBGPClusterConfig, *v2alpha1.CiliumBGPClusterConfigList](
			fake.Fake,
			"",
			v2alpha1.SchemeGroupVersion.WithResource("ciliumbgpclusterconfigs"),
			v2alpha1.SchemeGroupVersion.WithKind("CiliumBGPClusterConfig"),
			func() *v2alpha1.CiliumBGPClusterConfig { return &v2alpha1.CiliumBGPClusterConfig{} },
			func() *v2alpha1.CiliumBGPClusterConfigList { return &v2alpha1.CiliumBGPClusterConfigList{} },
			func(dst, src *v2alpha1.CiliumBGPClusterConfigList) { dst.ListMeta = src.ListMeta },
			func(list *v2alpha1.CiliumBGPClusterConfigList) []*v2alpha1.CiliumBGPClusterConfig {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v2alpha1.CiliumBGPClusterConfigList, items []*v2alpha1.CiliumBGPClusterConfig) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
