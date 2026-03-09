package pipe

import (
	"context"

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
		case chan any:
			v, ok := <-item
			if !ok {
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
