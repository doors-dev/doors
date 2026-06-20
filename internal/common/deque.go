// Copyright 2026 doors dev LLC
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

package common

import (
	"sync"

	"github.com/gammazero/deque"
)

const maxPooledDequeBufferCap = 4 * 1024

var dequeBufferPool = sync.Pool{
	New: func() any {
		return new(deque.Deque[any])
	},
}

func GetDequeBuffer() *deque.Deque[any] {
	return dequeBufferPool.Get().(*deque.Deque[any])
}

func FreezeDequeBuffer(buffer *deque.Deque[any]) {
	if buffer.Cap() == 0 {
		return
	}
	buffer.SetBaseCap(buffer.Cap())
}

func PutDequeBuffer(buffer *deque.Deque[any]) {
	if buffer == nil {
		return
	}
	buffer.Clear()
	if buffer.Cap() > maxPooledDequeBufferCap {
		return
	}
	dequeBufferPool.Put(buffer)
}

func ReleaseDequeBuffer(buffer *deque.Deque[any]) {
	if buffer == nil {
		return
	}
	for item := range buffer.Iter() {
		if child, ok := item.(*deque.Deque[any]); ok {
			ReleaseDequeBuffer(child)
		}
	}
	PutDequeBuffer(buffer)
}
