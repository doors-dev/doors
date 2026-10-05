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

package door

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/gox"
)

type failPrinter struct{}

func (failPrinter) Send(gox.Job) error {
	return errors.New("print failed")
}

// A print error releases the stack's buffers to the pool only after the
// iteration over them ends, so another goroutine reusing a pooled buffer never
// races with it. The race detector reports a violation.
func TestStackPrintErrorReleasesAfterIteration(t *testing.T) {
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				b := common.GetDequeBuffer()
				for k := range 40 {
					b.PushBack(k)
				}
				for b.Len() > 2 {
					b.PopFront()
				}
				common.PutDequeBuffer(b)
			}
		})
	}
	for range 5000 {
		root := common.GetDequeBuffer()
		child := common.GetDequeBuffer()
		for range 20 {
			child.PushBack(gox.NewJobText(context.Background(), "x"))
		}
		root.PushBack(child)
		stack := Stack{root}
		if err := stack.Print(failPrinter{}); err == nil {
			t.Fatal("expected the print error")
		}
	}
	close(stop)
	wg.Wait()
}
