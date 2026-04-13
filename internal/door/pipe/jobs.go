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

package pipe

import (
	"context"
	"sync/atomic"

	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

type Stack interface {
	Print(pr gox.Printer) error
}

type stack []*deque.Deque[any]

func (p *stack) Print(pr gox.Printer) error {
cycle:
	next := p.next()
	if next == nil {
		return nil
	}
	for item := range next.IterPopFront() {
		switch item := item.(type) {
		case *atomic.Value:
			v := item.Load()
			if v == nil {
				continue
			}
			switch v := v.(type) {
			case *deque.Deque[any]:
				p.push(v)
				goto cycle
			case error:
				if err := p.onErr(pr, v); err != nil {
					return err
				}
			}
		case gox.Job:
			if err := pr.Send(item); err != nil {
				return err
			}
		default:
			panic("unknown  item type in the buffer")
		}
	}
	p.pop()
	goto cycle
}

func (p stack) next() *deque.Deque[any] {
	if len(p) == 0 {
		return nil
	}
	return p[len(p)-1]
}

func (p *stack) push(buf *deque.Deque[any]) {
	*p = append(*p, buf)
}

func (p *stack) pop() {
	(*p)[len(*p)-1] = nil
	*p = (*p)[:len(*p)-1]
}

func (p *stack) onErr(pr gox.Printer, err error) error {
	return pr.Send(gox.NewJobComp(context.Background(), NewError(err)))
}
