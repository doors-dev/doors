package renderer

import (
	"context"

	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

func Render(spawner *shredder.Spawner, ctx context.Context, comp gox.Comp) Renderer {
	q := getPipe(spawner)
	spawner.Go(func() {
		defer q.close()
		comp.Main().Print(ctx, q)
	})
	return newRenderer(q)
}

func newRenderer(root *spipe) Renderer {
	return Renderer{
		stack: []*spipe{root},
	}
}

type Renderer struct {
	stack []*spipe
}

func (p *Renderer) last() *spipe {
	return p.stack[len(p.stack)-1]
}

func (p *Renderer) pop() bool {
	if len(p.stack) == 0 {
		return false
	}
	last := p.last()
	p.stack[len(p.stack)-1] = nil
	p.stack = p.stack[:len(p.stack)-1]
	putPipe(last)
	return len(p.stack) != 0
}

func (p *Renderer) push(q *spipe) {
	p.stack = append(p.stack, q)
}

func (p *Renderer) Next() (gox.Job, bool) {
	item, closed := p.last().get()
	if closed {
		if !p.pop() {
			return nil, false
		}
		return p.Next()
	}
	switch i := item.(type) {
	case *spipe:
		p.push(i)
		return p.Next()
	case gox.Job:
		return i, true
	default:
		panic("invalid type")
	}
}
