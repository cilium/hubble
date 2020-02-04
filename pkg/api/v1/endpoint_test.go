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

package v1

import (
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEndpoint_EqualsByID(t *testing.T) {
	type fields struct {
		Created      time.Time
		Deleted      *time.Time
		ContainerID  []string
		ID           uint64
		IPv4         net.IP
		IPv6         net.IP
		PodName      string
		PodNamespace string
	}
	type args struct {
		o *Endpoint
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "compare by a same ID and all other fields different should be considered equal",
			fields: fields{
				Created:      time.Unix(0, 1),
				ContainerID:  []string{"foo"},
				ID:           1,
				IPv4:         net.ParseIP("2.2.2.2"),
				PodName:      "",
				PodNamespace: "",
			},
			args: args{
				o: &Endpoint{
					Created:      time.Unix(0, 2),
					ContainerIDs: []string{"bar"},
					ID:           1,
					IPv4:         net.ParseIP("1.1.1.1"),
					PodName:      "",
					PodNamespace: "",
				},
			},
			want: true,
		},
		{
			name: "compare by a same ID, but different pod name should be considered different",
			fields: fields{
				Created:      time.Unix(0, 1),
				ContainerID:  []string{"foo"},
				ID:           1,
				IPv4:         net.ParseIP("2.2.2.2"),
				PodName:      "pod-bar",
				PodNamespace: "",
			},
			args: args{
				o: &Endpoint{
					Created:      time.Unix(0, 2),
					ContainerIDs: []string{"bar"},
					ID:           1,
					IPv4:         net.ParseIP("1.1.1.1"),
					PodName:      "pod-foo",
					PodNamespace: "",
				},
			},
			want: false,
		},
		{
			name: "compare by a same ID, but different namespace should be considered different",
			fields: fields{
				Created:      time.Unix(0, 1),
				ContainerID:  []string{"foo"},
				ID:           1,
				IPv4:         net.ParseIP("2.2.2.2"),
				PodName:      "pod-bar",
				PodNamespace: "kube-system",
			},
			args: args{
				o: &Endpoint{
					Created:      time.Unix(0, 2),
					ContainerIDs: []string{"bar"},
					ID:           1,
					IPv4:         net.ParseIP("1.1.1.1"),
					PodName:      "pod-bar",
					PodNamespace: "cilium",
				},
			},
			want: false,
		},
		{
			name: "compare by a same ID where podname and podnamespace are empty should be considered equal",
			fields: fields{
				Created:      time.Unix(0, 1),
				ContainerID:  []string{"foo"},
				ID:           1,
				IPv4:         net.ParseIP("2.2.2.2"),
				PodName:      "",
				PodNamespace: "",
			},
			args: args{
				o: &Endpoint{
					Created:      time.Unix(0, 2),
					ContainerIDs: []string{"bar"},
					ID:           1,
					IPv4:         net.ParseIP("1.1.1.1"),
					PodName:      "foo",
					PodNamespace: "default",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Endpoint{
				Created:      tt.fields.Created,
				Deleted:      tt.fields.Deleted,
				ContainerIDs: tt.fields.ContainerID,
				ID:           tt.fields.ID,
				IPv4:         tt.fields.IPv4,
				IPv6:         tt.fields.IPv6,
				PodName:      tt.fields.PodName,
				PodNamespace: tt.fields.PodNamespace,
			}
			if got := e.EqualsByID(tt.args.o); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEndpoint_SetFrom(t *testing.T) {
	type fields struct {
		Created      time.Time
		Deleted      *time.Time
		ContainerIDs []string
		ID           uint64
		IPv4         net.IP
		IPv6         net.IP
		PodName      string
		PodNamespace string
		Labels       []string
	}
	type args struct {
		o *Endpoint
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Endpoint
	}{
		{
			name: "all fields are copied except the time based ones",
			fields: fields{
				Created:      time.Unix(0, 0),
				ContainerIDs: nil,
				ID:           0,
				IPv4:         nil,
				IPv6:         nil,
				PodName:      "",
				PodNamespace: "",
				Labels:       nil,
			},
			args: args{
				o: &Endpoint{
					Created:      time.Unix(1, 1),
					Deleted:      &time.Time{},
					ContainerIDs: []string{"foo"},
					ID:           1,
					IPv4:         net.ParseIP("1.1.1.1"),
					IPv6:         net.ParseIP("fd00::"),
					PodName:      "pod-bar",
					PodNamespace: "cilium",
					Labels:       []string{"a", "b"},
				},
			},
			want: &Endpoint{
				Created:      time.Unix(0, 0),
				ContainerIDs: []string{"foo"},
				ID:           1,
				IPv4:         net.ParseIP("1.1.1.1"),
				IPv6:         net.ParseIP("fd00::"),
				PodName:      "pod-bar",
				PodNamespace: "cilium",
				Labels:       []string{"a", "b"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Endpoint{
				Created:      tt.fields.Created,
				Deleted:      tt.fields.Deleted,
				ContainerIDs: tt.fields.ContainerIDs,
				ID:           tt.fields.ID,
				IPv4:         tt.fields.IPv4,
				IPv6:         tt.fields.IPv6,
				PodName:      tt.fields.PodName,
				PodNamespace: tt.fields.PodNamespace,
			}
			e.SetFrom(tt.args.o)
			if !reflect.DeepEqual(e, tt.want) {
				t.Errorf("SetFrom() got = %v, want %v", e, tt.want)
			}
		})
	}
}

func TestEndpoints_SyncEndpoints(t *testing.T) {
	es := &Endpoints{
		mutex: sync.RWMutex{},
		eps:   []*Endpoint{},
	}

	eps := []*Endpoint{
		{
			ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb479"},
			ID:           1,
			Created:      time.Unix(0, 0),
			IPv4:         net.ParseIP("1.1.1.1").To4(),
			IPv6:         net.ParseIP("fd00::1").To16(),
			PodName:      "foo",
			PodNamespace: "default",
		},
		{
			ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb471"},
			ID:           2,
			Created:      time.Unix(0, 1),
			IPv4:         net.ParseIP("1.1.1.2").To4(),
			IPv6:         net.ParseIP("fd00::2").To16(),
			PodName:      "bar",
			PodNamespace: "default",
		},
		{
			ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb471"},
			ID:           3,
			Created:      time.Unix(0, 2),
			IPv4:         net.ParseIP("1.1.1.3").To4(),
			IPv6:         net.ParseIP("fd00::3").To16(),
			PodName:      "bar",
			PodNamespace: "kube-system",
		},
	}

	endpointsWanted := []*Endpoint{
		{
			ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb479"},
			ID:           1,
			Created:      time.Unix(0, 0),
			IPv4:         net.ParseIP("1.1.1.1").To4(),
			IPv6:         net.ParseIP("fd00::1").To16(),
			PodName:      "foo",
			PodNamespace: "default",
		},
		{
			ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb471"},
			ID:           2,
			Created:      time.Unix(0, 1),
			IPv4:         net.ParseIP("1.1.1.2").To4(),
			IPv6:         net.ParseIP("fd00::2").To16(),
			PodName:      "bar",
			PodNamespace: "default",
		},
		{
			ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb471"},
			ID:           3,
			Created:      time.Unix(0, 2),
			IPv4:         net.ParseIP("1.1.1.3").To4(),
			IPv6:         net.ParseIP("fd00::3").To16(),
			PodName:      "bar",
			PodNamespace: "kube-system",
		},
	}

	// add 2 new endpoints
	es.SyncEndpoints(eps[0:2])

	es.mutex.RLock()
	// check if the endpoints were added
	assert.EqualValues(t, endpointsWanted[0:2], es.eps)
	es.mutex.RUnlock()

	// Add only the first endpoint, meaning the 2nd one will be marked as deleted
	es.SyncEndpoints(eps[0:1])

	es.mutex.Lock()
	assert.NotNil(t, es.eps[1].Deleted)
	es.eps[1].Deleted = nil
	assert.EqualValues(t, endpointsWanted[0:2], es.eps)
	es.mutex.Unlock()

	// Re-add all endpoints
	es.SyncEndpoints(endpointsWanted)
	es.mutex.RLock()
	// check if the endpoints were added
	assert.EqualValues(t, endpointsWanted, es.eps)
	es.mutex.RUnlock()
}

func TestEndpoints_FindEPs(t *testing.T) {
	type fields struct {
		mutex sync.RWMutex
		eps   []*Endpoint
	}
	type args struct {
		epID      uint64
		namespace string
		podName   string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []Endpoint
	}{
		{
			name: "return all eps in a particular namespace",
			fields: fields{
				mutex: sync.RWMutex{},
				eps: []*Endpoint{
					{
						ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb479"},
						ID:           1,
						Created:      time.Unix(0, 0),
						IPv4:         net.ParseIP("1.1.1.1").To4(),
						IPv6:         net.ParseIP("fd00::1").To16(),
						PodName:      "foo",
						PodNamespace: "default",
					},
					{
						ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb471"},
						ID:           2,
						Created:      time.Unix(0, 0),
						IPv4:         net.ParseIP("1.1.1.2").To4(),
						IPv6:         net.ParseIP("fd00::2").To16(),
						PodName:      "bar",
						PodNamespace: "default",
					},
					{
						ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb473"},
						ID:           3,
						Created:      time.Unix(0, 0),
						IPv4:         net.ParseIP("1.1.1.3").To4(),
						IPv6:         net.ParseIP("fd00::3").To16(),
						PodName:      "bar",
						PodNamespace: "kube-system",
					},
				},
			},
			args: args{
				namespace: "default",
			},
			want: []Endpoint{
				{
					ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb479"},
					ID:           1,
					Created:      time.Unix(0, 0),
					IPv4:         net.ParseIP("1.1.1.1").To4(),
					IPv6:         net.ParseIP("fd00::1").To16(),
					PodName:      "foo",
					PodNamespace: "default",
				},
				{
					ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb471"},
					ID:           2,
					Created:      time.Unix(0, 0),
					IPv4:         net.ParseIP("1.1.1.2").To4(),
					IPv6:         net.ParseIP("fd00::2").To16(),
					PodName:      "bar",
					PodNamespace: "default",
				},
			},
		},
		{
			name: "return the ep of a pod",
			fields: fields{
				mutex: sync.RWMutex{},
				eps: []*Endpoint{
					{
						ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb479"},
						ID:           1,
						Created:      time.Unix(0, 0),
						IPv4:         net.ParseIP("1.1.1.1").To4(),
						IPv6:         net.ParseIP("fd00::1").To16(),
						PodName:      "foo",
						PodNamespace: "default",
					},
					{
						ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb471"},
						ID:           2,
						Created:      time.Unix(0, 0),
						IPv4:         net.ParseIP("1.1.1.2").To4(),
						IPv6:         net.ParseIP("fd00::2").To16(),
						PodName:      "bar",
						PodNamespace: "default",
					},
					{
						ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb473"},
						ID:           3,
						Created:      time.Unix(0, 0),
						IPv4:         net.ParseIP("1.1.1.3").To4(),
						IPv6:         net.ParseIP("fd00::3").To16(),
						PodName:      "bar",
						PodNamespace: "kube-system",
					},
				},
			},
			args: args{
				podName:   "bar",
				namespace: "kube-system",
			},
			want: []Endpoint{
				{
					ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb473"},
					ID:           3,
					Created:      time.Unix(0, 0),
					IPv4:         net.ParseIP("1.1.1.3").To4(),
					IPv6:         net.ParseIP("fd00::3").To16(),
					PodName:      "bar",
					PodNamespace: "kube-system",
				},
			},
		},
		{
			name: "return eps with the given pod name and namespace",
			fields: fields{
				mutex: sync.RWMutex{},
				eps: []*Endpoint{
					{
						ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb479"},
						ID:           1,
						Created:      time.Unix(0, 0),
						IPv4:         net.ParseIP("1.1.1.1").To4(),
						IPv6:         net.ParseIP("fd00::1").To16(),
						PodName:      "foo",
						PodNamespace: "default",
					},
					{
						ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb471"},
						ID:           2,
						Created:      time.Unix(0, 0),
						IPv4:         net.ParseIP("1.1.1.2").To4(),
						IPv6:         net.ParseIP("fd00::2").To16(),
						PodName:      "bar",
						PodNamespace: "default",
					},
					{
						ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb473"},
						ID:           3,
						Created:      time.Unix(0, 0),
						IPv4:         net.ParseIP("1.1.1.3").To4(),
						IPv6:         net.ParseIP("fd00::3").To16(),
						PodName:      "bar",
						PodNamespace: "kube-system",
					},
				},
			},
			args: args{
				epID: 2,
			},
			want: []Endpoint{
				{
					ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb471"},
					ID:           2,
					Created:      time.Unix(0, 0),
					IPv4:         net.ParseIP("1.1.1.2").To4(),
					IPv6:         net.ParseIP("fd00::2").To16(),
					PodName:      "bar",
					PodNamespace: "default",
				},
			},
		},
		{
			name: "do not return deleted endpoint",
			fields: fields{
				mutex: sync.RWMutex{},
				eps: []*Endpoint{
					{
						ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb479"},
						ID:           1,
						Created:      time.Unix(0, 0),
						Deleted:      &time.Time{},
						IPv4:         net.ParseIP("1.1.1.1").To4(),
						IPv6:         net.ParseIP("fd00::1").To16(),
						PodName:      "foo",
						PodNamespace: "default",
					},
					{
						ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb471"},
						ID:           2,
						Created:      time.Unix(0, 0),
						IPv4:         net.ParseIP("1.1.1.2").To4(),
						IPv6:         net.ParseIP("fd00::2").To16(),
						PodName:      "foo",
						PodNamespace: "default",
					},
					{
						ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb473"},
						ID:           3,
						Created:      time.Unix(0, 0),
						IPv4:         net.ParseIP("1.1.1.3").To4(),
						IPv6:         net.ParseIP("fd00::3").To16(),
						PodName:      "bar",
						PodNamespace: "kube-system",
					},
				},
			},
			args: args{
				podName:   "foo",
				namespace: "default",
			},
			want: []Endpoint{
				{
					ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb471"},
					ID:           2,
					Created:      time.Unix(0, 0),
					IPv4:         net.ParseIP("1.1.1.2").To4(),
					IPv6:         net.ParseIP("fd00::2").To16(),
					PodName:      "foo",
					PodNamespace: "default",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &Endpoints{
				mutex: tt.fields.mutex,
				eps:   tt.fields.eps,
			}
			if got := es.FindEPs(tt.args.epID, tt.args.namespace, tt.args.podName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindEPs() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test that we can modify the endpoint without disrupting the original
	// endpoint
	epWant := Endpoint{
		ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb479"},
		ID:           1,
		Created:      time.Unix(0, 0),
		IPv4:         net.ParseIP("1.1.1.1").To4(),
		IPv6:         net.ParseIP("fd00::1").To16(),
		PodName:      "foo",
		PodNamespace: "default",
	}
	eps := []*Endpoint{&epWant}
	es := &Endpoints{
		mutex: sync.RWMutex{},
		eps:   eps,
	}
	gotEps := es.FindEPs(1, "", "")
	assert.Len(t, gotEps, 1)
	assert.Equal(t, gotEps[0], epWant, 1)
	gotEps[0].Created = time.Unix(1, 1)
	gotEps[0].ContainerIDs = append(gotEps[0].ContainerIDs, "foo")

	epWantModified := Endpoint{
		ContainerIDs: []string{"313c63b8b164a19ec0fe42cd86c4159f3276ba8a415d77f340817fcfee2cb479", "foo"},
		ID:           1,
		Created:      time.Unix(1, 1),
		IPv4:         net.ParseIP("1.1.1.1").To4(),
		IPv6:         net.ParseIP("fd00::1").To16(),
		PodName:      "foo",
		PodNamespace: "default",
	}

	assert.NotEqual(t, gotEps[0], epWant, 1)
	assert.Equal(t, gotEps[0], epWantModified, 1)
}

func TestEndpoints_MarkDeleted(t *testing.T) {
	t1 := time.Unix(0, 0)
	es := &Endpoints{
		mutex: sync.RWMutex{},
		eps:   nil,
	}
	ep1 := &Endpoint{
		Created:      time.Time{},
		Deleted:      &t1,
		ContainerIDs: nil,
		ID:           1,
		IPv4:         nil,
		IPv6:         nil,
		PodName:      "",
		PodNamespace: "",
	}
	ep2 := &Endpoint{
		Created:      time.Time{},
		ContainerIDs: nil,
		ID:           2,
		IPv4:         net.ParseIP("1.1.1.1").To4(),
		IPv6:         nil,
		PodName:      "",
		PodNamespace: "",
	}
	wantedEPs := []*Endpoint{ep1}
	es.MarkDeleted(ep1)

	es.mutex.Lock()
	// check if the endpoints were added
	assert.EqualValues(t, wantedEPs, es.eps)

	// manually add one more endpoint
	es.eps = append(es.eps, ep2)
	es.mutex.Unlock()

	ep2Copy := &Endpoint{
		Created:      time.Time{},
		Deleted:      &t1,
		ContainerIDs: nil,
		ID:           2,
		IPv4:         nil,
		IPv6:         nil,
		PodName:      "",
		PodNamespace: "",
	}
	// Make sure we only mark the endpoint as deleted, i.e. we don't copy
	// any other fields into the internal copy
	es.MarkDeleted(ep2Copy)

	ep2Wanted := &Endpoint{
		Created:      time.Time{},
		Deleted:      &t1,
		ContainerIDs: nil,
		ID:           2,
		IPv4:         net.ParseIP("1.1.1.1").To4(),
		IPv6:         nil,
		PodName:      "",
		PodNamespace: "",
	}
	wantedEPs = append(wantedEPs, ep2Wanted)
	assert.EqualValues(t, wantedEPs, es.eps)
}

func TestEndpoints_GetEndpoint(t *testing.T) {
	type fields struct {
		mutex sync.RWMutex
		eps   []*Endpoint
	}
	type args struct {
		ip net.IP
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantEndpoint *Endpoint
		wantOk       bool
	}{
		{
			name: "found",
			fields: fields{
				eps: []*Endpoint{
					{
						ID:   15,
						IPv4: net.ParseIP("1.1.1.1"),
					},
				},
			},
			args: args{
				ip: net.ParseIP("1.1.1.1"),
			},
			wantEndpoint: &Endpoint{
				ID:   15,
				IPv4: net.ParseIP("1.1.1.1"),
			},
			wantOk: true,
		},
		{
			name: "not found",
			fields: fields{
				eps: []*Endpoint{
					{
						ID:   15,
						IPv4: net.ParseIP("1.1.1.1"),
					},
				},
			},
			args: args{
				ip: net.ParseIP("1.1.1.2"),
			},
			wantEndpoint: nil,
			wantOk:       false,
		},
		{
			name: "deleted",
			fields: fields{
				eps: []*Endpoint{
					{
						ID:      15,
						IPv4:    net.ParseIP("1.1.1.1"),
						Deleted: &time.Time{},
					},
				},
			},
			args: args{
				ip: net.ParseIP("1.1.1.1"),
			},
			wantEndpoint: nil,
			wantOk:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := &Endpoints{
				mutex: tt.fields.mutex,
				eps:   tt.fields.eps,
			}
			gotEndpoint, gotOk := es.GetEndpoint(tt.args.ip)
			if !reflect.DeepEqual(gotEndpoint, tt.wantEndpoint) {
				t.Errorf("GetEndpoint() gotEndpoint = %v, want %v", gotEndpoint, tt.wantEndpoint)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetEndpoint() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestEndpoints_GarbageCollect(t *testing.T) {
	es := &Endpoints{
		eps: []*Endpoint{
			{
				ID:   1,
				IPv4: net.ParseIP("1.1.1.1"),
			},
			{
				ID:   2,
				IPv4: net.ParseIP("1.1.1.2"),
			},
			{
				ID:   3,
				IPv4: net.ParseIP("1.1.1.3"),
			},
		},
	}

	del := time.Unix(0, 0)
	es.MarkDeleted(&Endpoint{ID: 2, Deleted: &del})
	assert.Equal(t, 3, len(es.eps))

	es.GarbageCollect()
	assert.Equal(t, 2, len(es.eps))

	_, ok := es.GetEndpoint(net.ParseIP("1.1.1.1"))
	assert.True(t, ok)
	_, ok = es.GetEndpoint(net.ParseIP("1.1.1.2"))
	assert.False(t, ok)
	_, ok = es.GetEndpoint(net.ParseIP("1.1.1.3"))
	assert.True(t, ok)

	es.MarkDeleted(&Endpoint{ID: 1, Deleted: &del})
	es.MarkDeleted(&Endpoint{ID: 3, Deleted: &del})
	es.GarbageCollect()
	assert.Equal(t, 0, len(es.eps))
}

func TestEndpoints_GetEndpointByContainerID(t *testing.T) {
	type fields struct {
		mutex sync.RWMutex
		eps   []*Endpoint
	}
	type args struct {
		id string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantEndpoint *Endpoint
		wantOk       bool
	}{
		{
			name: "found",
			fields: fields{
				eps: []*Endpoint{
					{
						ID:           15,
						ContainerIDs: []string{"c0", "c1"},
					},
				},
			},
			args: args{
				id: "c1",
			},
			wantEndpoint: &Endpoint{
				ID:           15,
				ContainerIDs: []string{"c0", "c1"},
			},
			wantOk: true,
		},
		{
			name: "not found",
			fields: fields{
				eps: []*Endpoint{
					{
						ID:           15,
						ContainerIDs: []string{"c0", "c1"},
					},
				},
			},
			args: args{
				id: "c2",
			},
			wantEndpoint: nil,
			wantOk:       false,
		},
		{
			name: "deleted",
			fields: fields{
				eps: []*Endpoint{
					{
						ID:           15,
						ContainerIDs: []string{"c0", "c1"},
						Deleted:      &time.Time{},
					},
				},
			},
			args: args{
				id: "c0",
			},
			wantEndpoint: nil,
			wantOk:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := Endpoints{
				mutex: tt.fields.mutex,
				eps:   tt.fields.eps,
			}
			gotEndpoint, gotOk := es.GetEndpointByContainerID(tt.args.id)
			assert.Equal(t, tt.wantEndpoint, gotEndpoint)
			assert.Equal(t, tt.wantOk, gotOk)
		})
	}
}

func TestEndpoint_Copy(t *testing.T) {
	ep := &Endpoint{
		Created:      time.Unix(1, 2),
		Deleted:      nil,
		ContainerIDs: nil,
		ID:           0,
		IPv4:         nil,
		IPv6:         nil,
		PodName:      "",
		PodNamespace: "",
		Labels:       nil,
	}
	cp := ep.DeepCopy()
	assert.Equal(t, ep, cp)

	deleted := time.Unix(3, 4)
	ep = &Endpoint{
		Created:      time.Unix(1, 2),
		Deleted:      &deleted,
		ContainerIDs: []string{"c1", "c2"},
		ID:           3,
		IPv4:         net.ParseIP("1.1.1.1"),
		IPv6:         net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
		PodName:      "pod1",
		PodNamespace: "ns1",
		Labels:       []string{"a=b", "c=d"},
	}
	cp1 := ep.DeepCopy()
	cp2 := ep.DeepCopy()
	assert.Equal(t, ep, cp2)
	assert.Equal(t, cp1, cp2)
	cp1.Created = time.Unix(5, 6)
	*cp1.Deleted = time.Unix(7, 8)
	cp1.ContainerIDs = []string{"c3", "c4"}
	cp1.ID = 4
	cp1.IPv4 = net.ParseIP("2.2.2.2")
	cp1.IPv4 = net.ParseIP("eeee:eeee:eeee:eeee:eeee:eeee:eeee:eeee")
	cp1.PodName = "pod1"
	cp1.PodNamespace = "ns1"
	cp1.Labels = []string{"e=f"}
	assert.Equal(t, ep, cp2)
	assert.NotEqual(t, cp1, cp2)
}
