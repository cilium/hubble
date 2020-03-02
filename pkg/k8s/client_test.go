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
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	utiltesting "k8s.io/client-go/util/testing"
)

var (
	hubbleEndpoints = &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hubble",
			Namespace: "kube-system",
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP:       "10.0.0.1",
						NodeName: stringPtr("k8s1"),
					}, {
						IP:       "fd00::1",
						NodeName: stringPtr("k8s1"),
					}, {
						IP:       "10.0.0.2",
						NodeName: stringPtr("k8s2"),
					}, {
						IP:       "fd00::2",
						NodeName: stringPtr("k8s2"),
					},
				},
				Ports: []v1.EndpointPort{
					{
						Port: 8181,
					},
				},
			}, {
				Addresses: []v1.EndpointAddress{
					{
						IP:       "10.0.0.3",
						NodeName: stringPtr("k8s3"),
					},
				},
				Ports: []v1.EndpointPort{
					{
						Port: 8183,
					},
				},
			}, {
				Addresses: []v1.EndpointAddress{
					{
						IP:       "10.0.0.4",
						NodeName: stringPtr("k8s4"),
					},
				},
				Ports: []v1.EndpointPort{
					{
						Port: 0,
					},
				},
			}, {
				Addresses: []v1.EndpointAddress{
					{
						IP:       "10.0.0.5",
						NodeName: nil,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Port: 8184,
					},
				},
			},
		},
	}
)

func stringPtr(s string) *string {
	return &s
}

func newTestClient(t testing.TB, srv *httptest.Server) *rest.RESTClient {
	baseURL, _ := url.Parse("http://localhost")
	if srv != nil {
		var err error
		baseURL, err = url.Parse(srv.URL)
		if err != nil {
			t.Fatalf("failed to parse test URL: %v", err)
		}
	}
	gv := v1.SchemeGroupVersion
	cfg := rest.ClientContentConfig{
		ContentType:  "application/json",
		GroupVersion: gv,
		Negotiator:   runtime.NewClientNegotiator(scheme.Codecs.WithoutConversion(), gv),
	}
	client, err := rest.NewRESTClient(baseURL, "/api/"+v1.SchemeGroupVersion.Version, cfg, nil, nil)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return client
}

func Test_GetServiceEndpoints(t *testing.T) {
	endpointsBody, err := runtime.Encode(scheme.Codecs.LegacyCodec(v1.SchemeGroupVersion), hubbleEndpoints)
	if err != nil {
		t.Fatalf("failed to encode endpoints: %v", err)
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	type h struct {
		statusCode   int
		responseBody string
	}
	tests := []struct {
		name      string
		handler   h
		context   context.Context
		namespace string
		service   string
		want      map[string][]string
		wantErr   bool
	}{
		{
			name:      "existing endpoints",
			handler:   h{200, string(endpointsBody)},
			context:   context.Background(),
			namespace: "kube-system",
			service:   "hubble",
			want: map[string][]string{
				"k8s1": {"10.0.0.1:8181", "[fd00::1]:8181"},
				"k8s2": {"10.0.0.2:8181", "[fd00::2]:8181"},
				"k8s3": {"10.0.0.3:8183"},
			},
		}, {
			name:      "non-existing endpoints",
			handler:   h{statusCode: 404},
			context:   context.Background(),
			namespace: "kube-system",
			service:   "hubble",
			want:      nil,
			wantErr:   true,
		}, {
			name:      "canceled context",
			handler:   h{statusCode: 200, responseBody: string(endpointsBody)},
			context:   canceledCtx,
			namespace: "kube-system",
			service:   "hubble",
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := utiltesting.FakeHandler{
				StatusCode:   tt.handler.statusCode,
				ResponseBody: tt.handler.responseBody,
				T:            t,
			}
			srv := httptest.NewServer(&handler)
			defer srv.Close()
			c := &Client{
				client: newTestClient(t, srv),
			}
			got, err := c.GetServiceEndpoints(tt.context, tt.namespace, tt.service)
			if err != nil && !tt.wantErr {
				t.Errorf("got error=%v, want <nil>", err)
			}
			assert.Equal(t, tt.want, got)
			if handler.RequestReceived != nil {
				assert.Equal(t, "/api/v1/namespaces/kube-system/endpoints/hubble", handler.RequestReceived.URL.String())
			}
		})
	}
}
