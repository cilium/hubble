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

package container

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/types"

	pb "github.com/cilium/hubble/api/v1/flow"
)

func newTimeFlow(sec int64, nano int32) *pb.Flow {
	return &pb.Flow{
		Time: &types.Timestamp{
			Seconds: sec,
			Nanos:   nano,
		},
	}
}

func TestNewPriorityQueue(t *testing.T) {
	waitEmptyCh := func(inCh chan *pb.Flow) {
		for {
			// wait until all payloads were processed
			if len(inCh) == 0 {
				break
			}
			time.Sleep(time.Millisecond)
		}
	}
	waitFillCh := func(inCh chan *pb.Flow) {
		for {
			// wait until all payloads were processed
			if len(inCh) != 0 {
				break
			}
			time.Sleep(time.Millisecond)
		}
	}

	inCh := make(chan *pb.Flow, 10)
	outCh := make(chan *pb.Flow, 10)
	ti := make(chan time.Time)
	NewPriorityQueueWith(inCh, outCh, ti)

	for i := int32(0); i < 9; i++ {
		select {
		case inCh <- newTimeFlow(100, 100+i):
		default:
			t.Error("Should have accepted incoming Payload")
		}
		select {
		case <-outCh:
			t.Error("Should not have received any incoming Payload")
		default:
		}
	}

	waitEmptyCh(inCh)

	select {
	case <-outCh:
		t.Error("Should not have received any incoming Payload")
	default:
	}

	select {
	case inCh <- newTimeFlow(100, 102):
	default:
		t.Error("Should have accepted incoming Payload")
	}

	waitEmptyCh(inCh)
	waitFillCh(outCh)

	select {
	case p := <-outCh:
		if !p.Time.Equal(newTimeFlow(100, 100).Time) {
			t.Error("Should have received the oldest payload")
		}
	default:
		t.Error("Should have received an incoming Payload")
	}

	ti <- time.Now()
	waitFillCh(outCh)

	select {
	case p := <-outCh:
		if !p.Time.Equal(newTimeFlow(100, 101).Time) {
			t.Errorf("Should have received the oldest payload, received %s", p)
		}
	default:
		t.Error("Should have received an incoming Payload")
	}

	ti <- time.Now()
	waitFillCh(outCh)

	select {
	case p := <-outCh:
		if !p.Time.Equal(newTimeFlow(100, 102).Time) {
			t.Errorf("Should have received the oldest payload, received %s", p)
		}
	default:
		t.Error("Should have received an incoming Payload")
	}

	select {
	case inCh <- newTimeFlow(99, 101):
	default:
		t.Error("Should have accepted incoming Payload")
	}

	select {
	case inCh <- newTimeFlow(101, 101):
	default:
		t.Error("Should have accepted incoming Payload")
	}

	waitEmptyCh(inCh)
	ti <- time.Now()
	waitFillCh(outCh)

	select {
	case p := <-outCh:
		if !p.Time.Equal(newTimeFlow(99, 101).Time) {
			t.Errorf("Should have received the oldest payload, received %s", p)
		}
	default:
		t.Error("Should have received an incoming Payload")
	}

	ti <- time.Now()
	waitFillCh(outCh)

	select {
	case p := <-outCh:
		if !p.Time.Equal(newTimeFlow(100, 102).Time) {
			t.Errorf("Should have received the oldest payload, received %s, want 100.102", p)
		}
	default:
		t.Error("Should have received an incoming Payload")
	}

	for i := 0; i < 100; i++ {
		ti <- time.Now()
	}
}
