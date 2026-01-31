package door

import (
	"context"
	"errors"
	"fmt"
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
}

func (r *pipe) SendTo(printer gox.Printer) {
	if r.parent != nil {
		panic("Can't initiate printing with owned renderer")
	}
	r.print(printer)
}

func (r *pipe) Send(job gox.Job) error {
	switch r.resourceState {
	case lookForResource:
		return r.lookForResource(job)
	case resourceContent:
		return r.resourceContent(job)
	case closeResource:
		return r.closeResource(job)
	default:
		panic("invalid pipe resource state")
	}
}

func (r *pipe) submit(fun func(ok bool)) {
	r.renderFrame.Submit(r.tracker.ctx, r.tracker.root.runtime(), fun)
}

func (r *pipe) send(job gox.Job) error {
	switch job := job.(type) {
	case *gox.JobHeadOpen:
		job.Attrs.ApplyMods()
		r.job(job)
	case *node:
		job.render(r)
	case *gox.JobComp:
		comp := job.Comp
		ctx := job.Ctx
		gox.Release(job)
		newRenderer := r.branch()
		newRenderer.submit(func(ok bool) {
			defer newRenderer.close()
			if !ok {
				return
			}
			comp.Main().Print(ctx, newRenderer)
		})
	default:
		r.job(job)
	}
	return nil
}

func (r *pipe) lookForResource(job gox.Job) error {
	head, ok := job.(*gox.JobHeadOpen)
	switch true {
	case !ok:
		return r.send(job)
	case strings.EqualFold(head.Tag, "script"):
		if src := head.Attrs.Get("src"); src.IsSet() {
			return r.send(job)
		}
		if escape := head.Attrs.Get("escape"); escape.IsSet() {
			escape.SetBool(false)
			return r.send(job)
		}
		if typ := head.Attrs.Get("type"); typ.IsSet() {
			typ, _ := typ.ReadString()
			if !strings.EqualFold(typ, "text/javascript") && !strings.EqualFold(typ, "application/javascript") {
				return r.send(job)
			}
		}
		r.resourceState = resourceContent
		r.resourceOpenHead = head
		return nil
	case strings.EqualFold(head.Tag, "style"):
		if escape := head.Attrs.Get("escape"); escape.IsSet() {
			escape.SetBool(false)
			return r.send(job)
		}
		r.resourceState = resourceContent
		r.resourceOpenHead = head
		return nil
	default:
		return r.send(job)
	}
}

func (r *pipe) resourceContent(job gox.Job) error {
	raw, ok := job.(*gox.JobRaw)
	if !ok {
		return errors.New("door: invalid resource content")
	}
	r.resourceText = raw.Text
	r.resourceState = closeResource
	gox.Release(raw)
	return nil
}

func (r *pipe) prepareResource(res *resources.Resource, mode resources.ResourceMode, name string, ext string) (string, bool) {
	if name == "" {
		name = "inline"
	}
	if mode == resources.ModeHost {
		return fmt.Sprintf("/d00r/r/%s.%s.%s", res.HashString(), name, ext), true
	} else {
		hook, ok := r.tracker.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			res.Serve(w, r)
			return false
		}, nil)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("/d00r/%s/%d/%d/%s.%s", r.tracker.root.inst.ID(), hook.DoorID, hook.HookID, name, ext), true
	}

}

func (r *pipe) closeResource(job gox.Job) error {
	openHead := r.resourceOpenHead
	content := r.resourceText
	r.resourceState = lookForResource
	r.resourceText = ""
	r.resourceOpenHead = nil
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
	registry := r.tracker.root.resourceRegistry()
	switch true {
	case strings.EqualFold(close.Tag, "script"):
		res, err := registry.Script(resources.ScriptInline{
			Content: content,
		}, resources.FormatDefault{}, "", mode)
		if err != nil {
			return err
		}
		src, ok := r.prepareResource(res, mode, name, "js")
		if !ok {
			return errors.New("door: can't prepare resource")
		}
		openHead.Attrs.Get("src").Set(src)
		if err := r.send(openHead); err != nil {
			return err
		}
		if err := r.send(close); err != nil {
			return err
		}
	case strings.EqualFold(close.Tag, "style"):
		res, err := registry.Style(resources.StyleString{
			Content: content,
		}, true, mode)
		if err != nil {
			return err
		}
		href, ok := r.prepareResource(res, mode, name, "css")
		if !ok {
			return errors.New("door: can't prepare resource")
		}
		openHead.Kind = gox.KindVoid
		openHead.Tag = "link"
		openHead.Attrs.Get("rel").Set("stylesheet")
		openHead.Attrs.Get("href").Set(href)
		gox.Release(close)
		return r.send(openHead)
	default:
		panic("unexpected resource tag: " + close.Tag)
	}
	return nil
}

func (r *pipe) print(printer gox.Printer) {
	stack := []*pipe{r}
	closed := false
main:
	for len(stack) != 0 {
		rr := stack[len(stack)-1]
		rr.mu.Lock()
		for rr.buffer.Len() != 0 {
			next := rr.buffer.PopFront()
			switch next := next.(type) {
			case gox.Job:
				printer.Send(next)
			case *pipe:
				rr.mu.Unlock()
				stack = append(stack, next)
				continue main
			}
		}
		closed = rr.closed
		if closed {
			bufferPool.Put(rr.buffer)
			rr.buffer = nil
			rr.mu.Unlock()
			stack[len(stack)-1] = nil
			stack = stack[:len(stack)-1]
			continue
		}
		rr.printer = printer
		rr.mu.Unlock()
		return
	}
	if !closed || r.parent == nil {
		return
	}
	r.parent.print(printer)
}

func (r *pipe) close() {
	r.mu.Lock()
	if r.closed {
		panic("renderer is already closed")
	}
	r.closed = true
	r.renderFrame.Release()
	done := r.printer != nil && r.buffer.Len() == 0
	r.mu.Unlock()
	if !done || r.parent == nil {
		return
	}
	r.parent.print(r.printer)
}

func (r *pipe) job(job gox.Job) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		panic("render is closed")
	}
	if r.printer != nil {
		r.printer.Send(job)
		return
	}
	r.buffer.PushBack(job)
}

func (p *pipe) branch() *pipe {
	p.mu.Lock()
	if p.closed {
		panic("render is closed")
	}
	newPipe := newPipe(p.rootFrame)
	newPipe.tracker = p.tracker
	newPipe.renderFrame = p.renderFrame
	newPipe.parent = p
	if p.printer != nil {
		printer := p.printer
		p.printer = nil
		p.mu.Unlock()
		newPipe.print(printer)
		return newPipe
	}
	p.buffer.PushBack(newPipe)
	p.mu.Unlock()
	return newPipe
}
