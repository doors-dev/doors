package doors

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/node"
	"github.com/doors-dev/doors/internal/resources"
)

// Fragment is a helper interface for defining composable, stateful, and code-interactive components.
//
// A Fragment groups fields, methods, and rendering logic into a reusable unit.
// This is especially useful when a simple templ function is not sufficient — for example,
// when you need to manage internal state, expose multiple methods, or control updates from Go code.
//
// Fragments implement the Render method and can be rendered using the F() helper.
//
// A Fragment can be stored in a variable, rendered once, and later updated by calling custom methods.
// These methods typically encapsulate internal Node updates — such as a Refresh() function
// that re-renders part of the fragment’s content manually.
//
// By default, a Fragment is static — its output does not change after rendering.
// To enable dynamic behavior, use a root-level Node to support targeted updates.
//
// Example:
//
//	type Counter struct {
//	    node  Node
//	    count int
//	}
//
//	func (c *Counter) Refresh(ctx context.Context) {
//	    c.node.Update(ctx, c.display())
//	}
//
//	templ (c *Counter) Render() {
//	    @c.node {
//	        @c.display()
//	    }
//	    <button { d.A(ctx, d.Click{
//	        On: func(ctx context.Context, _ d.EventRequest[d.PointerEvent]) bool {
//	            c.count++
//	            c.Refresh(ctx)
//	            return false
//	        },
//	    })... }>
//	        Click Me!
//	    </button>
//	}
//
//	templ (c *Counter) display() {
//	    Clicked { fmt.Sprint(c.count) } time(s)!
//	}
type Fragment interface {
	Render() templ.Component
}

// F renders a Fragment as a templ.Component.
//
// This helper wraps a Fragment and returns a valid templ.Component,
// enabling Fragments to be used inside other templ components.
//
// Example:
//
//	templ Demo() {
//	    @F(&Counter{})
//	}
func F(f Fragment) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return f.Render().Render(ctx, w)
	})
}

// Node represents a dynamic placeholder in the DOM tree that can be updated,
// replaced, or removed at runtime.
//
// It is a fundamental building block of the framework, used to manage and update dynamic HTML content.
// All changes made to a Node are automatically synchronized with the frontend DOM.
// If synchronization is slower than the update rate (e.g., due to network latency),
// operations may overwrite each other. In such cases, only the latest state will be propagated,
// ensuring the frontend eventually reflects the correct result.
//
// Remove and Replace are terminal operations — once invoked, the Node becomes static
// and behaves like a regular templ.Component. No further updates or mutations are accepted.
//
// A Node is itself a templ.Component and can be used directly in templates:
//
//	@node
//	// or
//	@node {
//	    // initial HTML content
//	}
//
// Nodes remain valid even if they have not yet been rendered, or were removed due to a parent
// being unmounted. Updates in such cases are deferred and automatically applied the next time
// the Node is rendered. If the same Node is rendered multiple times on the page, the previously
// rendered instance will be silently removed.
//
// The context used when rendering a Node's content follows the Node's lifecycle.
// This allows you to safely use `ctx.Done()` inside background goroutines
// that depend on the Node's presence in the DOM.
//
// Extended methods (prefixed with X) return a channel that is closed once the operation
// is confirmed by the frontend, allowing for safe coordination, sequencing, or error handling.
//
// During a single render cycle, Nodes and their children are guaranteed to observe
// consistent Beam values, ensuring stable and predictable rendering.
type Node = node.Node

// Sub is a helper component that wraps a Node whose content updates reactively
// based on the current value of the provided Beam.
//
// It subscribes to the Beam and re-renders the inner content whenever the value changes.
// The render function is called with the current Beam value and must return a templ.Component.
//
// This is the preferred way to bind Beam values into the DOM in a declarative and reactive manner.
//
// Example:
//
//	templ display(value int) {
//	    <span>{strconv.Itoa(value)}</span>
//	}
//
//	templ demo(beam Beam[int]) {
//	    @Sub(beam, func(v int) templ.Component {
//	        return display(v)
//	    })
//	}
//
// Parameters:
//   - beam: the reactive Beam to observe.
//   - render: a function that maps the current Beam value to a templ.Component.
//
// Returns:
//   - A templ.Component that updates as the Beam value changes.
func Sub[T any](beam Beam[T], render func(T) templ.Component) templ.Component {
	return E(func(ctx context.Context) templ.Component {
		node := &Node{}
		ok := beam.Sub(ctx, func(ctx context.Context, v T) bool {
			return !node.Update(ctx, render(v))
		})
		if !ok {
			return nil
		}
		return node
	})
}

func Extract[T any](ctx context.Context, beam Beam[T]) (T, bool) {
	ref, ok := ctx.Value(beam).(*T)
	if !ok || ref == nil {
		var t T
		return t, false
	}
	return *ref, false
}

func Inject[T any](beam Beam[T]) templ.Component {
	return E(func(ctx context.Context) templ.Component {
		children := templ.GetChildren(ctx)
		ctx = templ.ClearChildren(ctx)
		node := &Node{}
		ok := beam.Sub(ctx, func(ctx context.Context, v T) bool {
			return !node.Update(
				ctx,
				templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
					ctx = context.WithValue(ctx, beam, &v)
					return children.Render(ctx, w)
				}))
		})
		if !ok {
			return nil
		}
		return node
	})
}

// E is a helper component that evaluates the provided function at render time
// and returns the resulting templ.Component.
//
// This is useful when rendering logic is complex or better expressed in plain Go code,
// rather than templ syntax.
//
// Example:
//
//	@E(func(ctx context.Context) templ.Component {
//	    user, err := db.Get(id)
//	    if err != nil {
//	        return RenderError(err)
//	    }
//	    return RenderUser(user)
//	})
//
// Parameters:
//   - f: a function that returns a templ.Component, given the current render context.
//
// Returns:
//   - A templ.Component produced by evaluating f during rendering.
func E(f func(context.Context) templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		content := f(ctx)
		if content == nil {
			return nil
		}
		return content.Render(ctx, w)
	})
}

// Go starts a goroutine at render time using a blocking-safe context tied to the component's lifecycle.
//
// The goroutine runs only if the component is rendered. The context is canceled when the component
// is unmounted, but you must explicitly listen to ctx.Done() to stop work.
//
// The context allows safe blocking, making it safe to use with X* operations (e.g., XUpdate, XRemove).
//
// Example:
//
//	@Go(func(ctx context.Context) {
//	    for {
//	        ch, ok := node.XUpdate(ctx, content())
//	        if !ok {
//	            return
//	        }
//	        select {
//	        case err := <-ch:
//	            log.Println("confirmed update:", err)
//	        case <-ctx.Done():
//	            return
//	        }
//	    }
//
// Parameters:
//   - f: a function to run in a goroutine, scoped to the component's render lifecycle.
//
// Returns:
//   - A non-visual templ.Component that starts the goroutine when rendered.
func Go(f func(context.Context)) templ.Component {
	return E(func(ctx context.Context) templ.Component {
		ctx = common.SetBlockingCtx(ctx)
		go f(ctx)
		return nil
	})
}

type script struct {
	mode resources.InlineMode
}

func Script() templ.Component {
	return script{
		mode: resources.InlineModeHost,
	}
}

func ScriptLocal() templ.Component {
	return script{
		mode: resources.InlineModeLocal,
	}
}
func ScriptLocalNoCache() templ.Component {
	return script{
		mode: resources.InlineModeNoCache,
	}
}

func (s script) Render(ctx context.Context, w io.Writer) error {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	script := templ.GetChildren(ctx)
	ctx = templ.ClearChildren(ctx)
	buf := &bytes.Buffer{}
	err := script.Render(ctx, buf)
	if err != nil {
		return err
	}
	resource, err := inst.ImportRegistry().InlineScript(buf.Bytes(), s.mode)
	if err != nil {
		return err
	}
	if resource == nil {
		return nil
	}
	nonce, inline := inst.InlineNonce()
	if inline && s.mode != resources.InlineModeHost {
		resource.Attrs["nonce"] = nonce
	}
	return scriptRender(resource, inline, s.mode).Render(ctx, w)
}

type style struct {
	mode resources.InlineMode
}

func Style() templ.Component {
	return style{
		mode: resources.InlineModeHost,
	}
}

func StyleLocal() templ.Component {
	return style{
		mode: resources.InlineModeLocal,
	}
}
func StyleLocalNoCache() templ.Component {
	return style{
		mode: resources.InlineModeNoCache,
	}
}

func (s style) Render(ctx context.Context, w io.Writer) error {
	inst := ctx.Value(common.InstanceCtxKey).(instance.Core)
	style := templ.GetChildren(ctx)
	ctx = templ.ClearChildren(ctx)
	buf := &bytes.Buffer{}
	err := style.Render(ctx, buf)
	if err != nil {
		return err
	}
	resource, err := inst.ImportRegistry().InlineStyle(buf.Bytes(), s.mode)
	if err != nil {
		return err
	}
	if resource == nil {
		return nil
	}
	nonce, inline := inst.InlineNonce()
	if inline && s.mode != resources.InlineModeHost {
		resource.Attrs["nonce"] = nonce
	}
	return styleRender(resource, inline, s.mode).Render(ctx, w)
}

func renderRaw(tag string, attrs templ.Attributes, content []byte) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write(fmt.Appendf(nil, "<%s", tag))
		if err != nil {
			return err
		}
		err = templ.RenderAttributes(ctx, w, attrs)
		if err != nil {
			return err
		}
		_, err = w.Write(common.AsBytes(">"))
		if err != nil {
			return err
		}
		_, err = w.Write(content)
		if err != nil {
			return err
		}
		_, err = w.Write(fmt.Appendf(nil, "</%s>", tag))
		return err
	})
}
