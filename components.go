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

// Node represents a dynamic placeholder in the DOM tree that can be updated,
// replaced, or removed at runtime.
//
// It is a fundamental building block of the framework, used to manage dynamic HTML content.
// All changes made to a Node are automatically synchronized with the frontend DOM.
//
// A Node is itself a templ.Component and can be used directly in templates:
//
//	@node
//	// or
//	@node {
//	    // initial HTML content
//	}
//
// Nodes start inactive and become active when rendered. Operations on inactive nodes
// are stored virtually and applied when the node becomes active. If a node is removed
// or replaced, it becomes inactive again, but operations continue to update its virtual
// state for potential future rendering.
//
// The context used when rendering a Node's content follows the Node's lifecycle.
// This allows you to safely use `ctx.Done()` inside background goroutines
// that depend on the Node's presence in the DOM.
//
// Extended methods (prefixed with X) return a channel that can be used to track
// when operations complete. The channel receives nil on success or an error on failure,
// then closes. For inactive nodes, the channel closes immediately without sending a value.
//
// During a single render cycle, Nodes and their children are guaranteed to observe
// consistent state (Beam), ensuring stable and predictable rendering.
type Node = node.Node

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
// that re-renders part of the fragment's content manually.
//
// By default, a Fragment is static — its output does not change after rendering.
// To enable dynamic behavior, use a root-level Node to support targeted updates.
//
// Example:
//
//	type Counter struct {
//	    node  doors.Node
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
//	    <button { doors.A(ctx, doors.AClick{
//	        On: func(ctx context.Context, _ doors.REvent[doors.PointerEvent]) bool {
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
//	    @doors.F(&Counter{})
//	}
func F(f Fragment) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return f.Render().Render(ctx, w)
	})
}

// Sub creates a reactive component that automatically updates when a Beam value changes.
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
//	    @doors.Sub(beam, func(v int) templ.Component {
//	        return display(v)
//	    })
//	}
//
// Parameters:
//   - beam: the reactive Beam to observe
//   - render: a function that maps the current Beam value to a templ.Component
//
// Returns:
//   - A templ.Component that updates reactively as the Beam value changes
func Sub[T any](beam Beam[T], render func(T) templ.Component) templ.Component {
	return E(func(ctx context.Context) templ.Component {
		node := &Node{}
		ok := beam.Sub(ctx, func(ctx context.Context, v T) bool {
			node.Update(ctx, render(v))
			return false
		})
		if !ok {
			return nil
		}
		return node
	})
}


// Inject creates a reactive component that injects Beam values into the context for child components.
//
// It subscribes to the Beam and re-renders its children whenever the value changes,
// making the current value available to child components via Extract().
//
// This enables passing reactive values down the component tree without explicit prop drilling.
//
// Example:
//
//	@Inject("user", userBeam) {
//	    @UserProfile() // Can use ctx.Value("user").(User) to get current user
//	}
func Inject[T any](key any, beam Beam[T]) templ.Component {
	return E(func(ctx context.Context) templ.Component {
		children := templ.GetChildren(ctx)
		ctx = templ.ClearChildren(ctx)
		node := &Node{}
		ok := beam.Sub(ctx, func(ctx context.Context, v T) bool {
			node.Update(
				ctx,
				templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
					ctx = context.WithValue(ctx, key, v)
					return children.Render(ctx, w)
				}),
			)
			return false
		})
		if !ok {
			return nil
		}
		return node
	})
}

// E evaluates the provided function at render time and returns the resulting templ.Component.
//
// This is useful when rendering logic is complex or better expressed in plain Go code,
// rather than templ syntax. The function is called with the current render context.
//
// Example:
//
//	@doors.E(func(ctx context.Context) templ.Component {
//	    user, err := db.Get(id)
//	    if err != nil {
//	        return RenderError(err)
//	    }
//	    return RenderUser(user)
//	})
//
// Parameters:
//   - f: a function that returns a templ.Component, given the current render context
//
// Returns:
//   - A templ.Component produced by evaluating f during rendering
func E(f func(context.Context) templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		content := f(ctx)
		if content == nil {
			return nil
		}
		return content.Render(ctx, w)
	})
}

// Run runs function at render time
// useful for intitialization logic
func Run(f func(context.Context)) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, _ io.Writer) error {
		f(ctx)
		return nil
	})
}

// Go starts a goroutine at render time using a blocking-safe context tied to the component's lifecycle.
//
// The goroutine runs only if the component is rendered. The context is canceled when the component
// is unmounted, allowing for proper cleanup. You must explicitly listen to ctx.Done() to stop work.
//
// The context allows safe blocking operations, making it safe to use with X* operations (e.g., XUpdate, XRemove).
//
// Example:
//
//	@doors.Go(func(ctx context.Context) {
//	    for {
//	        select {
//	        case <-time.After(time.Second):
//	            node.Update(ctx, currentTime())
//	        case <-ctx.Done():
//	            return
//	        }
//	    }
//	})
//
// Parameters:
//   - f: a function to run in a goroutine, scoped to the component's render lifecycle
//
// Returns:
//   - A non-visual templ.Component that starts the goroutine when rendered
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

// Script converts inline script content to an external resource.
// The JavaScript/TypeScript content is processed by esbuild and served as an external
// resource with a src attribute. This is the default mode - the resource is publicly
// accessible as a static asset.
//
// The script content is automatically wrapped in an anonymous async function to support
// await and protect the global context. A special $d variable is provided to access
// frontend framework functions.
//
// The content must be wrapped in <script> tags. TypeScript is supported by adding
// type="text/typescript" attribute.
//
// Example:
//
//	@Script() {
//	    <script>
//	        console.log("Hello from [not] inline script!");
//	        // $d provides access to framework functions
//	        // await is supported due to async wrapper
//	        await $d.hook("hello","world");
//	    </script>
//	}
//
// Or with TypeScript:
//
//	@Script() {
//	    <script type="text/typescript">
//	        const message: string = "TypeScript works!";
//	        console.log(message);
//	    </script>
//	}
func Script() templ.Component {
	return script{
		mode: resources.InlineModeHost,
	}
}

// ScriptLocal converts inline script content to an external resource that is served
// securely within the current context scope. The script is processed and served with
// a src attribute, but not exposed as a publicly accessible static asset.
// The script content is wrapped in an anonymous async function and provides the $d variable.
// The content must be wrapped in <script> tags.
func ScriptLocal() templ.Component {
	return script{
		mode: resources.InlineModeLocal,
	}
}

// ScriptLocalNoCache converts inline script content to an external resource within
// the current context scope without caching. The script is processed on every render
// and served with a src attribute, but not exposed as a static resource.
// The script content is wrapped in an anonymous async function and provides the $d variable.
// The content must be wrapped in <script> tags.
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

// Style converts inline CSS content to an external resource.
// The CSS is minified and served as an external resource with an href attribute.
// This is the default mode - the resource is publicly accessible as a static asset.
//
// The content must be wrapped in <style> tags.
//
// Example:
//
//	@Style() {
//	    <style>
//	        .my-class {
//	            color: blue;
//	            font-size: 14px;
//	        }
//	    </style>
//	}
func Style() templ.Component {
	return style{
		mode: resources.InlineModeHost,
	}
}

// StyleLocal converts inline CSS content to an external resource that is served
// securely within the current context scope. The CSS is processed and served with
// an href attribute, but not exposed as a publicly accessible static asset.
// The content must be wrapped in <style> tags.
func StyleLocal() templ.Component {
	return style{
		mode: resources.InlineModeLocal,
	}
}

// StyleLocalNoCache converts inline CSS content to an external resource within
// the current context scope without caching. The CSS is processed on every render
// and served with an href attribute, but not exposed as a static resource.
// The content must be wrapped in <style> tags.
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

// Text converts any value to a component with escaped string using default formats.
//
// Example:
//
//	@Text("Hello <world>")  // Output: Hello &lt;world&gt;
//	@Text(42)               // Output: 42
//	@Text(user.Name)        // Output: John
func Text(value any) templ.Component {
	str := fmt.Sprint(value)
	escaped := templ.EscapeString(str)
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write(common.AsBytes(escaped))
		return err
	})
}
