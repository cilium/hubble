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
	"container/heap"
	"time"

	pb "github.com/cilium/hubble/api/v1/observer"
)

// A flowPriorityQueue implements heap.Interface and holds Items.
type flowPriorityQueue []*pb.Flow

func (pq flowPriorityQueue) Len() int { return len(pq) }

func (pq flowPriorityQueue) Less(i, j int) bool {
	if pq[i].Time.Seconds > pq[j].Time.Seconds {
		return false
	}

	if pq[i].Time.Seconds == pq[j].Time.Seconds {
		return pq[i].Time.Nanos < pq[j].Time.Nanos
	}

	return pq[i].Time.Seconds < pq[j].Time.Seconds
}

func (pq flowPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *flowPriorityQueue) Push(x interface{}) {
	item := x.(*pb.Flow)
	*pq = append(*pq, item)
}

func (pq *flowPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

// NewPriorityQueue orders the given pb.Flows by Timestamp and returns them
// by order in the 'outCh'
func NewPriorityQueue(inCh <-chan *pb.Flow, outCh chan<- *pb.Flow) {
	t := time.NewTicker(time.Second)
	// TODO write logic to return flows depending in the number of flows
	//  received per second. (i.e., if we receive 100 flows/s and we are not receiving
	//  1 flow/s we should keep returning 100 flows/s and decreasing the number
	//  of flows/s until we reach 1 flow/s
	NewPriorityQueueWith(inCh, outCh, t.C)
}

// NewPriorityQueueWith is similar to NewPriorityQueue with the exception
// that will also send a *pb.Flow through the 'outCh' every time the chan
// time.Time receives anything.
func NewPriorityQueueWith(inCh <-chan *pb.Flow, outCh chan<- *pb.Flow, t <-chan time.Time) {
	go func() {
		pq := flowPriorityQueue{}
		heap.Init(&pq)
		defer func() {
			// empty heap before exiting
			for pq.Len() != 0 {
				pl := heap.Pop(&pq).(*pb.Flow)
				outCh <- pl
			}
			close(outCh)
		}()
		i := 0
		for {
			select {
			case pl, ok := <-inCh:
				if !ok {
					return
				}
				heap.Push(&pq, pl)
				i++
				if i == cap(inCh) {
					i--
					pl := heap.Pop(&pq).(*pb.Flow)
					outCh <- pl
				}
			case <-t:
				if i == 0 {
					continue
				}
				i--
				pl := heap.Pop(&pq).(*pb.Flow)
				outCh <- pl
			}
		}
	}()
}
