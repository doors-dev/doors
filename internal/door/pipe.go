package door

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
	"github.com/gammazero/deque"
)

var bufferPool = sync.Pool{
	New: func() any {
		return &deque.Deque[any]{}
	},
}

func newPipe(rootFrame shredder.SimpleFrame) *pipe {
	return &pipe{
		buffer:    bufferPool.Get().(*deque.Deque[any]),
		rootFrame: rootFrame,
	}
}

type Pipe interface {
	SendTo(gox.Printer)
	Error() error
}

type resourceState int

const (
	lookForResource resourceState = iota
	resourceContent
	closeResource
)

type pipe struct {
	mu               sync.Mutex
	closed           bool
	parent           *pipe
	buffer           *deque.Deque[any]
	tracker          *tracker
	renderFrame      shredder.Frame
	rootFrame        shredder.SimpleFrame
	printer          gox.Printer
	resourceState    resourceState
	resourceOpenHead *gox.JobHeadOpen
	resourceText     string
	renderingError   error
	printingError    error
}

func (p *pipe) renderProxy(parentCtx context.Context, view *view, takoverFrame *shredder.ValveFrame) {
	proxy := newProxyComponent(p.tracker.id, view, parentCtx, takoverFrame)
	p.renderAny(p.tracker.ctx, proxy)
}

func (p *pipe) Error() error {
	if p.renderingError != nil {
		return p.renderingError
	}
	return p.printingError
}

func (p *pipe) renderView(parentCtx context.Context, view *view) {
	p.submit(func(ok bool) {
		defer p.close()
		if !ok {
			return
		}
		cur := gox.NewCursor(p.tracker.ctx, p)
		open, close := view.headFrame(parentCtx, p.tracker.id, cur.NewID())
		p.renderingError = p.send(open)
		if p.renderingError != nil {
			return
		}
		if view.content != nil {
			if comp, ok := view.content.(gox.Comp); ok {
				p.renderingError = comp.Main()(cur)
			} else {
				p.renderingError = cur.Any(view.content)
			}

			if p.renderingError != nil {
				return
			}
		}
		p.renderingError = p.send(close)
	})
}

func (p *pipe) renderAny(ctx context.Context, any any) {
	p.submit(func(ok bool) {
		defer p.close()
		if !ok {
			return
		}
		if any == nil {
			return
		}
		cur := gox.NewCursor(ctx, p)
		if comp, ok := any.(gox.Comp); ok {
			p.renderingError = comp.Main()(cur)
		} else {
			p.renderingError = cur.Any(any)
		}
	})
}

func (p *pipe) SendTo(printer gox.Printer) {
	if p.parent != nil {
		panic("Can't initiate printing with owned renderer")
	}
	p.mu.Lock()
	readyToPrint := p.closed
	p.printer = printer
	p.mu.Unlock()
	if !readyToPrint {
		return
	}
	p.print()
}

func (p *pipe) Send(job gox.Job) error {
	switch p.resourceState {
	case lookForResource:
		return p.lookForResource(job)
	case resourceContent:
		return p.resourceContent(job)
	case closeResource:
		return p.closeResource(job)
	default:
		panic("invalid pipe resource state")
	}
}

func (p *pipe) submit(fun func(ok bool)) {
	p.renderFrame.Submit(p.tracker.ctx, p.tracker.root.runtime(), fun)
}

func (p *pipe) send(job gox.Job) error {
	switch job := job.(type) {
	case *node:
		job.render(p)
	case *gox.JobHeadOpen:
		if err := job.Attrs.ApplyMods(job.Ctx, job.Tag); err != nil {
			return err
		}
		p.job(job)
	case *gox.JobComp:
		comp := job.Comp
		ctx := job.Ctx
		gox.Release(job)
		newRenderer := p.branch()
		newRenderer.renderAny(ctx, comp)
	default:
		p.job(job)
	}
	return nil
}

func (p *pipe) lookForResource(job gox.Job) error {
	head, ok := job.(*gox.JobHeadOpen)
	switch true {
	case !ok:
		return p.send(job)
	case strings.EqualFold(head.Tag, "script"):
		if head.Attrs.Has("src") {
			return p.send(job)
		}
		if head.Attrs.Has("escape") {
			head.Attrs.Get("escape").SetBool(false)
			return p.send(job)
		}
		if head.Attrs.Has("type") {
			typ, _ := head.Attrs.Get("type").ReadString()
			if !strings.EqualFold(typ, "text/javascript") && !strings.EqualFold(typ, "application/javascript") {
				return p.send(job)
			}
		}
		p.resourceState = resourceContent
		p.resourceOpenHead = head
		return nil
	case strings.EqualFold(head.Tag, "style"):
		if escape := head.Attrs.Get("escape"); escape.IsSet() {
			escape.SetBool(false)
			return p.send(job)
		}
		p.resourceState = resourceContent
		p.resourceOpenHead = head
		return nil
	default:
		return p.send(job)
	}
}

func (p *pipe) resourceContent(job gox.Job) error {
	raw, ok := job.(*gox.JobRaw)
	if !ok {
		return errors.New("door: invalid resource content")
	}
	p.resourceText = raw.Text
	p.resourceState = closeResource
	gox.Release(raw)
	return nil
}

func (p *pipe) prepareResource(res *resources.Resource, mode resources.ResourceMode, name string, ext string) (string, bool) {
	if name == "" {
		name = "inline"
	}
	if mode == resources.ModeHost {
		return fmt.Sprintf("/~0/r/%s.%s.%s", res.HashString(), name, ext), true
	} else {
		hook, ok := p.tracker.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			res.Serve(w, r)
			return false
		}, nil)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("/~0/%s/%d/%d/%s.%s", p.tracker.root.inst.ID(), hook.DoorID, hook.HookID, name, ext), true
	}

}

func (p *pipe) closeResource(job gox.Job) error {
	openHead := p.resourceOpenHead
	content := p.resourceText
	p.resourceState = lookForResource
	p.resourceText = ""
	p.resourceOpenHead = nil
	close, ok := job.(*gox.JobHeadClose)
	if !ok || close.ID != openHead.ID {
		return errors.New("door: invalid resource close job")
	}
	nocacheAttr := openHead.Attrs.Get("nocache")
	local := openHead.Attrs.Get("local")
	var mode resources.ResourceMode
	switch true {
	case nocacheAttr.IsSet():
		nocacheAttr.SetBool(false)
		mode = resources.ModeNoCache
	case local.IsSet():
		local.SetBool(false)
		mode = resources.ModeCache
	default:
		mode = resources.ModeHost
	}
	name, _ := openHead.Attrs.Get("data-name").ReadString()
	registry := p.tracker.root.resourceRegistry()
	switch true {
	case strings.EqualFold(close.Tag, "script"):
		res, err := registry.Script(resources.ScriptInline{
			Content: content,
		}, resources.FormatDefault{}, "", mode)
		if err != nil {
			return err
		}
		src, ok := p.prepareResource(res, mode, name, "js")
		if !ok {
			return errors.New("door: can't prepare resource")
		}
		openHead.Attrs.Get("src").Set(src)
		if err := p.send(openHead); err != nil {
			return err
		}
		if err := p.send(close); err != nil {
			return err
		}
	case strings.EqualFold(close.Tag, "style"):
		res, err := registry.Style(resources.StyleString{
			Content: content,
		}, true, mode)
		if err != nil {
			return err
		}
		href, ok := p.prepareResource(res, mode, name, "css")
		if !ok {
			return errors.New("door: can't prepare resource")
		}
		openHead.Kind = gox.KindVoid
		openHead.Tag = "link"
		openHead.Attrs.Get("rel").Set("stylesheet")
		openHead.Attrs.Get("href").Set(href)
		gox.Release(close)
		return p.send(openHead)
	default:
		panic("unexpected resource tag: " + close.Tag)
	}
	return nil
}

func (p *pipe) print() {
	stack := []*pipe{p}
main:
	for len(stack) != 0 {
		rr := stack[len(stack)-1]
		rr.mu.Lock()
		rr.printer = p.printer
		if !rr.closed {
			rr.mu.Unlock()
			return
		}
		if rr.renderingError == nil {
			for rr.buffer.Len() != 0 && rr.printingError == nil {
				next := rr.buffer.PopFront()
				switch next := next.(type) {
				case gox.Job:
					rr.printingError = rr.printer.Send(next)
				case *pipe:
					rr.mu.Unlock()
					stack = append(stack, next)
					continue main
				}
			}
		}
		if err := rr.Error(); err != nil {
			id := rr.tracker.root.NewID()
			slog.Error("door rendering error", "error", err, "error_id", id)
			rr.printer.Send(gox.NewJobComp(p.tracker.parentContext(), renderError{err: err, id: id}))
		}
		rr.buffer.Clear()
		bufferPool.Put(rr.buffer)
		rr.buffer = nil
		rr.mu.Unlock()
		stack[len(stack)-1] = nil
		stack = stack[:len(stack)-1]
	}
	if p.parent == nil {
		return
	}
	p.parent.print()
}

func (p *pipe) close() {
	p.mu.Lock()
	if p.closed {
		panic("renderer is already closed")
	}
	p.closed = true
	p.renderFrame.Release()
	readyToPrint := p.printer != nil
	p.mu.Unlock()
	if !readyToPrint {
		return
	}
	p.print()
}

func (p *pipe) job(job gox.Job) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		panic("render is closed")
	}
	p.buffer.PushBack(job)
}

func (p *pipe) branch() *pipe {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		panic("render is closed")
	}
	newPipe := newPipe(p.rootFrame)
	newPipe.tracker = p.tracker
	newPipe.renderFrame = p.renderFrame
	newPipe.parent = p
	p.buffer.PushBack(newPipe)
	return newPipe
}
