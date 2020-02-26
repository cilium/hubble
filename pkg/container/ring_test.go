// Copyright 2019-2020 Authors of Hubble
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

package container

import (
	"container/list"
	"container/ring"
	"context"
	"reflect"
	"sync"
	"testing"

	v1 "github.com/cilium/hubble/pkg/api/v1"
	"github.com/gogo/protobuf/types"
	"go.uber.org/goleak"
)

func BenchmarkRingWrite(b *testing.B) {
	entry := &v1.Event{}
	s := NewRing(b.N)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Write(entry)
	}
}

func BenchmarkRingRead(b *testing.B) {
	entry := &v1.Event{}
	s := NewRing(b.N)
	a := make([]*v1.Event, b.N, b.N)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Write(entry)
	}
	b.ResetTimer()
	lastWriteIdx := s.LastWriteParallel()
	for i := 0; i < b.N; i++ {
		a[i], _ = s.read(lastWriteIdx)
		lastWriteIdx--
	}
}

func BenchmarkTimeLibListRead(b *testing.B) {
	entry := &v1.Event{}
	s := list.New()
	a := make([]*v1.Event, b.N, b.N)
	i := 0
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.PushFront(entry)
	}
	b.ResetTimer()
	for e := s.Front(); e != nil; e = e.Next() {
		a[i], _ = e.Value.(*v1.Event)
	}
}

func BenchmarkTimeLibRingRead(b *testing.B) {
	entry := &v1.Event{}
	s := ring.New(b.N)
	a := make([]*v1.Event, b.N, b.N)
	i := 0
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Value = entry
		s.Next()
	}
	s.Do(func(e interface{}) {
		a[i], _ = e.(*v1.Event)
		i++
	})
}

func TestRing_Read(t *testing.T) {
	type fields struct {
		mask     uint64
		cycleExp uint8
		data     []*v1.Event
		write    uint64
	}
	type args struct {
		read uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *v1.Event
		want1  bool
	}{
		{
			name: "normal read for the index 7",
			fields: fields{
				mask:     0x7,
				cycleExp: 0x3, // 7+1=8=2^3
				data: []*v1.Event{
					0x0: {Timestamp: &types.Timestamp{Seconds: 0}},
					0x1: {Timestamp: &types.Timestamp{Seconds: 1}},
					0x2: {Timestamp: &types.Timestamp{Seconds: 2}},
					0x3: {Timestamp: &types.Timestamp{Seconds: 3}},
					0x4: {Timestamp: &types.Timestamp{Seconds: 4}},
					0x5: {Timestamp: &types.Timestamp{Seconds: 5}},
					0x6: {Timestamp: &types.Timestamp{Seconds: 6}},
					0x7: {Timestamp: &types.Timestamp{Seconds: 7}},
				},
				// next to be written: 0x9 (idx: 1), last written: 0x8 (idx: 0)
				write: 0x9,
			},
			args: args{
				read: 0x7,
			},
			want:  &v1.Event{Timestamp: &types.Timestamp{Seconds: 7}},
			want1: true,
		},
		{
			name: "we can't read index 0 since we just wrote into it",
			fields: fields{
				mask:     0x7,
				cycleExp: 0x3, // 7+1=8=2^3
				data: []*v1.Event{
					0x0: {Timestamp: &types.Timestamp{Seconds: 0}},
					0x1: {Timestamp: &types.Timestamp{Seconds: 1}},
					0x2: {Timestamp: &types.Timestamp{Seconds: 2}},
					0x3: {Timestamp: &types.Timestamp{Seconds: 3}},
					0x4: {Timestamp: &types.Timestamp{Seconds: 4}},
					0x5: {Timestamp: &types.Timestamp{Seconds: 5}},
					0x6: {Timestamp: &types.Timestamp{Seconds: 6}},
					0x7: {Timestamp: &types.Timestamp{Seconds: 7}},
				},
				// next to be written: 0x9 (idx: 2), last written: 0x8 (idx: 0)
				write: 0x9,
			},
			args: args{
				read: 0x0,
			},
			want:  nil,
			want1: false,
		},
		{
			name: "we can't read index 0x7 since we are one writing cycle ahead",
			fields: fields{
				mask:     0x7,
				cycleExp: 0x3, // 7+1=8=2^3
				data: []*v1.Event{
					0x0: {Timestamp: &types.Timestamp{Seconds: 0}},
					0x1: {Timestamp: &types.Timestamp{Seconds: 1}},
					0x2: {Timestamp: &types.Timestamp{Seconds: 2}},
					0x3: {Timestamp: &types.Timestamp{Seconds: 3}},
					0x4: {Timestamp: &types.Timestamp{Seconds: 4}},
					0x5: {Timestamp: &types.Timestamp{Seconds: 5}},
					0x6: {Timestamp: &types.Timestamp{Seconds: 6}},
					0x7: {Timestamp: &types.Timestamp{Seconds: 7}},
				},
				// next to be written: 0x10 (idx: 0), last written: 0x0f (idx: 7)
				write: 0x10,
			},
			args: args{
				// The next possible entry that we can read is 0x10-0x7-0x1 = 0x8 (idx: 0)
				read: 0x7,
			},
			want:  nil,
			want1: false,
		},
		{
			name: "we can read index 0x8 since it's the last entry that we can read in this cycle",
			fields: fields{
				mask:     0x7,
				cycleExp: 0x3, // 7+1=8=2^3
				data: []*v1.Event{
					0x0: {Timestamp: &types.Timestamp{Seconds: 0}},
					0x1: {Timestamp: &types.Timestamp{Seconds: 1}},
					0x2: {Timestamp: &types.Timestamp{Seconds: 2}},
					0x3: {Timestamp: &types.Timestamp{Seconds: 3}},
					0x4: {Timestamp: &types.Timestamp{Seconds: 4}},
					0x5: {Timestamp: &types.Timestamp{Seconds: 5}},
					0x6: {Timestamp: &types.Timestamp{Seconds: 6}},
					0x7: {Timestamp: &types.Timestamp{Seconds: 7}},
				},
				// next to be written: 0x10 (idx: 0), last written: 0x0f (idx: 7)
				write: 0x10,
			},
			args: args{
				// The next possible entry that we can read is 0x10-0x7-0x1 = 0x8 (idx: 0)
				read: 0x8,
			},
			want:  &v1.Event{Timestamp: &types.Timestamp{Seconds: 0}},
			want1: true,
		},
		{
			name: "we overflow write and we are trying to read the previous writes, that we can't",
			fields: fields{
				mask:     0x7,
				cycleExp: 0x3, // 7+1=8=2^3
				data: []*v1.Event{
					0x0: {Timestamp: &types.Timestamp{Seconds: 0}},
					0x1: {Timestamp: &types.Timestamp{Seconds: 1}},
					0x2: {Timestamp: &types.Timestamp{Seconds: 2}},
					0x3: {Timestamp: &types.Timestamp{Seconds: 3}},
					0x4: {Timestamp: &types.Timestamp{Seconds: 4}},
					0x5: {Timestamp: &types.Timestamp{Seconds: 5}},
					0x6: {Timestamp: &types.Timestamp{Seconds: 6}},
					0x7: {Timestamp: &types.Timestamp{Seconds: 7}},
				},
				// next to be written: 0x0 (idx: 0), last written: 0xffffffffffffffff (idx: 7)
				write: 0x0,
			},
			args: args{
				// We can't read this index because we might be still writing into it
				// next to be read: ^uint64(0) (idx: 7), last read: 0xfffffffffffffffe (idx: 6)
				read: ^uint64(0),
			},
			want:  nil,
			want1: false,
		},
		{
			name: "we overflow write and we are trying to read the previous writes, that we can",
			fields: fields{
				mask:     0x7,
				cycleExp: 0x3, // 7+1=8=2^3
				data: []*v1.Event{
					0x0: {Timestamp: &types.Timestamp{Seconds: 0}},
					0x1: {Timestamp: &types.Timestamp{Seconds: 1}},
					0x2: {Timestamp: &types.Timestamp{Seconds: 2}},
					0x3: {Timestamp: &types.Timestamp{Seconds: 3}},
					0x4: {Timestamp: &types.Timestamp{Seconds: 4}},
					0x5: {Timestamp: &types.Timestamp{Seconds: 5}},
					0x6: {Timestamp: &types.Timestamp{Seconds: 6}},
					0x7: {Timestamp: &types.Timestamp{Seconds: 7}},
				},
				// next to be written: 0x1 (idx: 1), last written: 0x0 (idx: 0)
				write: 0x1,
			},
			args: args{
				// next to be read: ^uint64(0) (idx: 7), last read: 0xfffffffffffffffe (idx: 6)
				read: ^uint64(0),
			},
			want:  &v1.Event{Timestamp: &types.Timestamp{Seconds: 7}},
			want1: true,
		},
		{
			name: "we overflow write and we are trying to read the 2 previously cycles",
			fields: fields{
				mask:     0x7,
				cycleExp: 0x3, // 7+1=8=2^3
				data: []*v1.Event{
					0x0: {Timestamp: &types.Timestamp{Seconds: 0}},
					0x1: {Timestamp: &types.Timestamp{Seconds: 1}},
					0x2: {Timestamp: &types.Timestamp{Seconds: 2}},
					0x3: {Timestamp: &types.Timestamp{Seconds: 3}},
					0x4: {Timestamp: &types.Timestamp{Seconds: 4}},
					0x5: {Timestamp: &types.Timestamp{Seconds: 5}},
					0x6: {Timestamp: &types.Timestamp{Seconds: 6}},
					0x7: {Timestamp: &types.Timestamp{Seconds: 7}},
				},
				// next to be written: 0x8 (idx: 1), last written: 0xffffffffffffffff (idx: 7)
				write: 0x8,
			},
			args: args{
				// next to be read: ^uint64(0)-0x7 (idx: 0), last read: 0xfffffffffffffff7 (idx: 7)
				// read is: ^uint64(0)-0x7 which should represent index 0x0 but
				// with a cycle that was already overwritten
				read: ^uint64(0) - 0x7,
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Ring{
				mask:      tt.fields.mask,
				data:      tt.fields.data,
				write:     tt.fields.write,
				dataLen:   uint64(len(tt.fields.data)),
				cycleExp:  tt.fields.cycleExp,
				cycleMask: ^uint64(0) >> tt.fields.cycleExp,
			}
			got, got1 := r.read(tt.args.read)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ring.read() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Ring.read() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestRing_Write(t *testing.T) {
	type fields struct {
		len   uint64
		data  []*v1.Event
		write uint64
	}
	type args struct {
		flow *v1.Event
	}
	tests := []struct {
		name   string
		fields fields
		want   fields
		args   args
	}{
		{
			name: "normal write",
			args: args{
				flow: &v1.Event{Timestamp: &types.Timestamp{Seconds: 5}},
			},
			fields: fields{
				len:   0x3,
				write: 0,
				data: []*v1.Event{
					0x0: {Timestamp: &types.Timestamp{Seconds: 0}},
					0x1: {Timestamp: &types.Timestamp{Seconds: 1}},
					0x2: {Timestamp: &types.Timestamp{Seconds: 2}},
					0x3: {Timestamp: &types.Timestamp{Seconds: 3}},
				},
			},
			want: fields{
				len:   0x3,
				write: 1,
				data: []*v1.Event{
					{Timestamp: &types.Timestamp{Seconds: 5}},
					{Timestamp: &types.Timestamp{Seconds: 1}},
					{Timestamp: &types.Timestamp{Seconds: 2}},
					{Timestamp: &types.Timestamp{Seconds: 3}},
				},
			},
		},
		{
			name: "overflow write",
			args: args{
				flow: &v1.Event{Timestamp: &types.Timestamp{Seconds: 5}},
			},
			fields: fields{
				len:   0x3,
				write: ^uint64(0),
				data: []*v1.Event{
					{Timestamp: &types.Timestamp{Seconds: 0}},
					{Timestamp: &types.Timestamp{Seconds: 1}},
					{Timestamp: &types.Timestamp{Seconds: 2}},
					{Timestamp: &types.Timestamp{Seconds: 3}},
				},
			},
			want: fields{
				len:   0x3,
				write: 0,
				data: []*v1.Event{
					{Timestamp: &types.Timestamp{Seconds: 0}},
					{Timestamp: &types.Timestamp{Seconds: 1}},
					{Timestamp: &types.Timestamp{Seconds: 2}},
					{Timestamp: &types.Timestamp{Seconds: 5}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Ring{
				mask:  tt.fields.len,
				data:  tt.fields.data,
				write: tt.fields.write,
				cond:  sync.NewCond(&sync.RWMutex{}),
			}
			r.Write(tt.args.flow)
			want := &Ring{
				mask:  tt.want.len,
				data:  tt.want.data,
				write: tt.want.write,
			}
			reflect.DeepEqual(want, r)
		})
	}
}

func TestRing_LastWriteParallel(t *testing.T) {
	type fields struct {
		len   uint64
		data  []*v1.Event
		write uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		{
			fields: fields{
				len:   0x3,
				write: 2,
				data:  []*v1.Event{},
			},
			want: 0,
		},
		{
			fields: fields{
				len:   0x3,
				write: 1,
				data:  []*v1.Event{},
			},
			want: ^uint64(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Ring{
				mask:  tt.fields.len,
				data:  tt.fields.data,
				write: tt.fields.write,
			}
			if got := r.LastWriteParallel(); got != tt.want {
				t.Errorf("Ring.LastWriteParallel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRing_LastWrite(t *testing.T) {
	type fields struct {
		len   uint64
		data  []*v1.Event
		write uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		{
			fields: fields{
				len:   0x3,
				write: 1,
				data:  []*v1.Event{},
			},
			want: 0,
		},
		{
			fields: fields{
				len:   0x3,
				write: 0,
				data:  []*v1.Event{},
			},
			want: ^uint64(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Ring{
				mask:  tt.fields.len,
				data:  tt.fields.data,
				write: tt.fields.write,
			}
			if got := r.LastWrite(); got != tt.want {
				t.Errorf("Ring.LastWrite() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRingFunctionalityInParallel(t *testing.T) {
	r := NewRing(0xf)
	if len(r.data) != 0x10 {
		t.Errorf("r.data should have a lenght of 0x10. Got %x", len(r.data))
	}
	if r.mask != 0xf {
		t.Errorf("r.mask should be 0xf. Got %x", r.mask)
	}
	if r.cycleExp != 4 {
		t.Errorf("r.cycleExp should be 4. Got %x", r.cycleExp)
	}
	if r.cycleMask != 0xfffffffffffffff {
		t.Errorf("r.cycleMask should be 0xfffffffffffffff. Got %x", r.cycleMask)
	}
	lastWrite := r.LastWriteParallel()
	if lastWrite != ^uint64(0)-1 {
		t.Errorf("lastWrite should be %x. Got %x", ^uint64(0)-1, lastWrite)
	}

	r.Write(&v1.Event{Timestamp: &types.Timestamp{Seconds: 0}})
	lastWrite = r.LastWriteParallel()
	if lastWrite != ^uint64(0) {
		t.Errorf("lastWrite should be %x. Got %x", ^uint64(0), lastWrite)
	}

	r.Write(&v1.Event{Timestamp: &types.Timestamp{Seconds: 1}})
	lastWrite = r.LastWriteParallel()
	if lastWrite != 0x0 {
		t.Errorf("lastWrite should be 0x0. Got %x", lastWrite)
	}

	entry, ok := r.read(lastWrite)
	if !ok {
		t.Errorf("Should be able to read position %x", lastWrite)
	}
	if !entry.Timestamp.Equal(&types.Timestamp{Seconds: 0}) {
		t.Errorf("Read Event should be %+v, got %+v instead", &types.Timestamp{Seconds: 0}, entry.Timestamp)
	}
	lastWrite--
	entry, ok = r.read(lastWrite)
	if !ok {
		t.Errorf("Should be able to read position %x", lastWrite)
	}
	if entry != nil {
		t.Errorf("Read Event should be %+v, got %+v instead", nil, entry)
	}
}

func TestRingFunctionalitySerialized(t *testing.T) {
	r := NewRing(0xf)
	if len(r.data) != 0x10 {
		t.Errorf("r.data should have a lenght of 0x10. Got %x", len(r.data))
	}
	if r.mask != 0xf {
		t.Errorf("r.mask should be 0xf. Got %x", r.mask)
	}
	lastWrite := r.LastWrite()
	if lastWrite != ^uint64(0) {
		t.Errorf("lastWrite should be %x. Got %x", ^uint64(0)-1, lastWrite)
	}

	r.Write(&v1.Event{Timestamp: &types.Timestamp{Seconds: 0}})
	lastWrite = r.LastWrite()
	if lastWrite != 0x0 {
		t.Errorf("lastWrite should be %x. Got %x", 0x0, lastWrite)
	}

	r.Write(&v1.Event{Timestamp: &types.Timestamp{Seconds: 1}})
	lastWrite = r.LastWrite()
	if lastWrite != 0x1 {
		t.Errorf("lastWrite should be 0x1. Got %x", lastWrite)
	}

	entry, ok := r.read(lastWrite)
	if ok {
		t.Errorf("Should not be able to read position %x", lastWrite)
	}
	lastWrite--
	entry, ok = r.read(lastWrite)
	if !ok {
		t.Errorf("Should be able to read position %x", lastWrite)
	}
	if !entry.Timestamp.Equal(&types.Timestamp{Seconds: 0}) {
		t.Errorf("Read Event should be %+v, got %+v instead", &types.Timestamp{Seconds: 0}, entry.Timestamp)
	}
}

func TestRing_ReadFrom_Test_1(t *testing.T) {
	defer goleak.VerifyNone(t)
	r := NewRing(0xf)
	if len(r.data) != 0x10 {
		t.Errorf("r.data should have a lenght of 0x10. Got %x", len(r.data))
	}
	if r.dataLen != 0x10 {
		t.Errorf("r.dataLen should have a lenght of 0x10. Got %x", r.dataLen)
	}
	if r.mask != 0xf {
		t.Errorf("r.mask should be 0xf. Got %x", r.mask)
	}
	lastWrite := r.LastWrite()
	if lastWrite != ^uint64(0) {
		t.Errorf("lastWrite should be %x. Got %x", ^uint64(0)-1, lastWrite)
	}

	// Add 5 flows
	for i := uint64(0); i < 5; i++ {
		r.Write(&v1.Event{Timestamp: &types.Timestamp{Seconds: int64(i)}})
		lastWrite = r.LastWrite()
		if lastWrite != i {
			t.Errorf("lastWrite should be %x. Got %x", i, lastWrite)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	ch := r.readFrom(ctx, 0)
	i := int64(0)
	for entry := range ch {
		if !entry.Timestamp.Equal(&types.Timestamp{Seconds: i}) {
			t.Errorf("Read Event should be %+v, got %+v instead", &types.Timestamp{Seconds: i}, entry.Timestamp)
		}
		i++
		if i == 4 {
			break
		}
	}
	cancel()
	flow, ok := <-ch
	if ok {
		t.Errorf("Channel should have been closed, received %+v", flow)
	}
}

func TestRing_ReadFrom_Test_2(t *testing.T) {
	defer goleak.VerifyNone(t)
	r := NewRing(0xf)
	if len(r.data) != 0x10 {
		t.Errorf("r.data should have a lenght of 0x10. Got %x", len(r.data))
	}
	if r.dataLen != 0x10 {
		t.Errorf("r.dataLen should have a lenght of 0x10. Got %x", r.dataLen)
	}
	if r.mask != 0xf {
		t.Errorf("r.mask should be 0xf. Got %x", r.mask)
	}
	lastWrite := r.LastWrite()
	if lastWrite != ^uint64(0) {
		t.Errorf("lastWrite should be %x. Got %x", ^uint64(0)-1, lastWrite)
	}

	// Add 5 flows
	for i := uint64(0); i < 5; i++ {
		r.Write(&v1.Event{Timestamp: &types.Timestamp{Seconds: int64(i)}})
		lastWrite = r.LastWrite()
		if lastWrite != i {
			t.Errorf("lastWrite should be %x. Got %x", i, lastWrite)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	// We should be able to read from a previou 'cycles' and ReadFrom will
	// be able to catch up with the writer.
	ch := r.readFrom(ctx, ^uint64(0)-15)
	i := int64(0)
	for entry := range ch {
		// Given the buffer length is 16 and there are no more writes being made,
		// we will receive 16-5=11 nil flows and 5 non-nil flows
		//
		//   ReadFrom +           +----------------valid read------------+  +position possibly being written
		//            |           |                                      |  |  +next position to be written (r.write)
		//            v           V                                      V  V  V
		// write: f0 f1 f2 f3 f4 f5 f6 f7 f8 f9 fa fb fc fd fe ff  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
		// index:  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
		// cycle: 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0
		if i < 16-5 {
			if entry != nil {
				t.Errorf("Read Event should be nil, got %+v instead", entry)
			}
		} else {
			if !entry.Timestamp.Equal(&types.Timestamp{Seconds: i - (16 - 5)}) {
				t.Errorf("Read Event should be %+v, got %+v instead", &types.Timestamp{Seconds: i - (16 - 5)}, entry.Timestamp)
			}
		}
		i++
		if i == 0xf {
			break
		}
	}
	cancel()
	flow, ok := <-ch
	if ok {
		t.Errorf("Channel should have been closed, received %+v", flow)
	}
}

func TestRing_ReadFrom_Test_3(t *testing.T) {
	defer goleak.VerifyNone(t)
	r := NewRing(0xf)
	if len(r.data) != 0x10 {
		t.Errorf("r.data should have a lenght of 0x10. Got %x", len(r.data))
	}
	if r.dataLen != 0x10 {
		t.Errorf("r.dataLen should have a lenght of 0x10. Got %x", r.dataLen)
	}
	if r.mask != 0xf {
		t.Errorf("r.mask should be 0xf. Got %x", r.mask)
	}
	lastWrite := r.LastWrite()
	if lastWrite != ^uint64(0) {
		t.Errorf("lastWrite should be %x. Got %x", ^uint64(0)-1, lastWrite)
	}

	// Add 5 flows
	for i := uint64(0); i < 5; i++ {
		r.Write(&v1.Event{Timestamp: &types.Timestamp{Seconds: int64(i)}})
		lastWrite = r.LastWrite()
		if lastWrite != i {
			t.Errorf("lastWrite should be %x. Got %x", i, lastWrite)
		}
	}

	// We should be able to read from a previous 'cycle' and ReadFrom will
	// be able to catch up with the writer.
	ctx, cancel := context.WithCancel(context.Background())
	ch := r.readFrom(ctx, ^uint64(0)-30)
	i := int64(0)
	for entry := range ch {
		// Given the buffer length is 16 and there are no more writes being made,
		// we will receive 16-5=11 nil flows and 5 non-nil flows
		//
		//   ReadFrom +           +----------------valid read------------+  +position possibly being written
		//            |           |                                      |  |  +next position to be written (r.write)
		//            v           V                                      V  V  V
		// write: f0 f1 f2 f3 f4 f5 f6 f7 f8 f9 fa fb fc fd fe ff  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
		// index:  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f  0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
		// cycle: 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f 1f  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0  0
		if i < 16-5 {
			if entry != nil {
				t.Errorf("Read Event should be nil, got %+v instead", entry)
			}
		} else {
			if !entry.Timestamp.Equal(&types.Timestamp{Seconds: i - (16 - 5)}) {
				t.Errorf("Read Event should be %+v, got %+v instead", &types.Timestamp{Seconds: i - (16 - 5)}, entry.Timestamp)
			}
		}
		i++
		if i == 0xf {
			break
		}
	}
	cancel()
	flow, ok := <-ch
	if ok {
		t.Errorf("Channel should have been closed, received %+v", flow)
	}
}
