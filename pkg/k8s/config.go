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
	"fmt"
	"time"

	"github.com/cilium/hubble/pkg"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ConfigOptions may be used to specify options when creating a new config with
// NewConfig, NewConfigFromKubeConfig or NewInClusterConfig.
type ConfigOptions struct {
	// QPS indicates the maximum QPS to the master from this client.
	QPS float32
	// Burst is the maximum burst for throttle.
	Burst int
	// Timeout is  the maximum length of time to wait before giving up on a
	// server request. A value of zero means no timeout.
	Timeout time.Duration
}

// NewConfig creates a config suitable to use with NewClient.
// The host parameter must be a host string, a host:port pair, or a URL to the
// base of the kubernetes apiserver.
func NewConfig(host string, opts *ConfigOptions) (*rest.Config, error) {
	config := &rest.Config{Host: host}
	return applyOptions(applyDefaults(config), opts)
}

// NewConfigFromKubeConfig creates a config suitable to use with NewClient.
// It uses the kube config pointed to by path to create it.
func NewConfigFromKubeConfig(path string, opts *ConfigOptions) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}
	return applyOptions(applyDefaults(config), opts)
}

// NewInClusterConfig creates a config suitable to use with NewClient.
// It assumes that the client is running inside a pod running on kubernetes
func NewInClusterConfig(opts *ConfigOptions) (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return applyOptions(applyDefaults(config), opts)
}

func applyDefaults(config *rest.Config) *rest.Config {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/api"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	return config
}

func applyOptions(config *rest.Config, opts *ConfigOptions) (*rest.Config, error) {
	config.UserAgent = fmt.Sprintf("hubble v%s", pkg.Version)
	if opts != nil {
		config.QPS = opts.QPS
		config.Burst = opts.Burst
		config.Timeout = opts.Timeout
	}
	if err := rest.SetKubernetesDefaults(config); err != nil {
		return nil, err
	}
	return config, nil
}
