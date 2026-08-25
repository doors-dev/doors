# Components

Components are Go values that implement `gox.Comp` by having a `Main() gox.Elem` method.

There is no virtual DOM in **Doors**. A component is static by default: `Main()` renders once and stays. Re-rendering is explicit and local. `Bind` and `Effect` re-render dynamic fragments, subscriptions can update Doors, and direct Door methods replace only the Door region. 

## Model

Use a struct component when the UI owns state, dependencies, or several methods:

```gox
func NewCounter() gox.Comp {
	return counter{count: doors.NewSource(0)}
}

type counter struct {
	count doors.Source[int]
}

elem (c counter) Main() {
	<button
		(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestPointer) bool {
				c.count.Mutate(ctx, func(i int) int { return i + 1 })
				return false
			},
		})>
		Add
	</button>

	~(c.count.Bind(c.countView))
}

elem (c counter) countView(i int) {
	<span>~(i)</span>
}
```

The `counter` value is created when it is placed in the tree. `Main()` renders once. The button stays mounted. Only the `Bind` fragment re-renders when `count` changes.

Use constructors to keep initialization private. Without `NewCounter`, callers would need to know that `count` must be initialized with `doors.NewSource(0)`.

## Rendering

Three tools cover most dynamic component rendering:

| Tool | Use |
| --- | --- |
| `Bind` | One source or beam drives one fragment. |
| `Effect` | One small container reads multiple related values inline. |
| `Door` | A handler, subscription, or background task explicitly replaces a region. |

### Bind

Use `Bind` when the relationship is direct: this value renders this fragment.

```gox
<>
	~(c.count.Bind(c.countView))
</>
```

The callback receives the value and **Doors** handles the subscription and Door updates.

### Effect

Use `Effect` when one small fragment needs to read several related values together:

```gox
<>
	~>(new(doors.Door)) <section>
		~~
		days, _ := d.days.Effect(ctx)
		units, ok := d.units.Effect(ctx)
		~~
		~(if ok {
			~(WeatherChart(days, units))
		})
	</section>
</>
```

The same pattern can be written in expression style when it is clearer to return one value:

```gox
<>
	~>(new(doors.Door)) ~({
		days, _ := d.days.Effect(ctx)
		units, ok := d.units.Effect(ctx)
		if !ok {
			return nil
		}

		return <section>
			~(WeatherChart(days, units))
		</section>
	})
</>
```

Keep `Effect` boundaries small. The whole container re-renders when any value read with `Effect` changes.

It is enough to check the last `ok`. `Effect` fails only when the context is already canceled.

### Door

Use a `Door` field when a component needs explicit updates from handlers, subscriptions, or background work:

```gox
type panel struct {
	body doors.Door
}

elem (p *panel) Main() {
	<button
		(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestPointer) bool {
				p.body.Inner(ctx, "Updated")
				return false
			},
		})>
		Update
	</button>

	~>(p.body) <div>Initial</div>
}
```

See [Door](./06-door.md) for the low-level API.

## Markup Encapsulation

A struct component's main job is to own a piece of markup — the tags, classes, and layout — and expose what varies as fields. Callers assemble pages from data-shaped literals; presentation stays in one place.

GoX renders any type, so a slot field is just `any`: it accepts a string, an element, another component, or a fragment of components.

```gox
type Bubble struct {
	Content any
	Accent  bool
}

elem (b Bubble) Main() {
	<div class=({
		if b.Accent {
			return "bubble accent"
		}
		return "bubble"
	})>~(b.Content)</div>
}
```

The same slot takes plain text or a markup tree:

```gox
elem examples() {
	~Bubble{Content: "Hello!"}
	~Bubble{Accent: true, Content: <>
		<h4>Links</h4>
		<a href="https://doors.dev">doors.dev</a>
	</>}
}
```

A nil slot renders nothing. Wrap dependent markup in `~(if ...)` when it should disappear together with the slot:

```gox
type Message struct {
	Avatar  any
	Content any
}

elem (m Message) Main() {
	<li class="message">
		~(if m.Avatar != nil {
			<div class="message-avatar">~(m.Avatar)</div>
		})
		<div class="message-body">~(m.Content)</div>
	</li>
}
```

Composition then reads as data:

```gox
elem conversation() {
	<ul class="chat">
		~Message{
			Avatar:  Avatar{Initials: "AZ"},
			Content: Bubble{Accent: true, Content: "What's Doors?"},
		}
		~Message{
			Content: Bubble{Content: "A Go framework for server-side interactive apps."},
		}
	</ul>
}
```

For repeated blocks, make one component per item and render them with an explicit `for` loop — wrappers, classes, IDs, and per-item attributes stay local to each item.

## State

Be explicit about who owns each source or beam:

- Local state: create it in the constructor and store it on the component.
- Shared state: accept a `Source` or `Beam` from the parent.
- Derived state: derive it from a parent source in the constructor.

Do not add a `doors.Source` or `doors.Beam` field and leave initialization implicit. A component with source fields should have a constructor or parent wiring that makes those fields non-nil before `Bind`, `Effect`, `Sub`, `Update`, or `Mutate` uses them.

```go
func LocationSelector(apply func(context.Context, Place)) gox.Comp {
	selected := doors.NewSource(Place{})
	return locationSelector{
		selected: selected,
		apply:    apply,
	}
}
```

Use derived sources and beams to keep updates narrow. One parent route or settings source can feed multiple small fragments without making the whole component re-render.

## Lifecycle

When a dynamic parent unmounts, **Doors** cancels everything inside it:

- `Bind`, `Effect`, and `Sub` subscriptions
- hook handlers
- mounted Doors
- scoped background work started with `doors.Go(...)`

Start timing and context semantics of `doors.Go(f)` are covered in [Core Concepts](./02-core-concepts.md).

### Disposable Components

Give `Main` a value receiver and initialize state inside it — every render then works on a fresh copy of the component value, and the declared value itself is never mutated:

```gox
type Counter struct {
	count doors.Source[int]
}

elem (c Counter) Main() {
	~~
	c.count = doors.NewSource(0)
	~~
	<div>
		~(c.count.Bind(func(v int) gox.Elem {
			return <span>~(v)</span>
		}))
		<button (doors.AClick{On: c.increment})>+</button>
	</div>
}

func (c *Counter) increment(ctx context.Context, _ doors.RequestPointer) bool {
	c.count.Mutate(ctx, func(v int) int { return v + 1 })
	return false
}
```

Three receivers working together:

- `Main` on a value: each render gets its own copy, so the top `~~ ~~` block acts as a constructor — no `NewCounter` needed.
- Handlers and helpers on a pointer: inside `Main`, `c.increment` binds to this render's copy — the one whose fields were just initialized. Everything in one mount shares that instance; two mounts share nothing.
- The declared value is a prototype: `~Counter{}` — or one shared `var counter Counter` — can be dropped anywhere, any number of times, with no stale state and no races through the shared value.

Where it pays off — a component value handed to a route branch is one Go value that renders again every time the route re-enters:

```gox
<>
	~(path.Route(
		doors.RouteMatch(func(p Path) bool { 
           return p.Section == SectionCounter 
        }).Comp(Counter{}),
		doors.RouteDefaultComp[Path](Dashboard{}),
	))
</>
```

A stateful component constructed outside the render would come back with its previous state; the disposable `Counter` re-initializes on every entry.

The flip side: state lives for one render of that fragment. Every re-render resets it. State that must survive re-renders needs an owner above the component — a pointer receiver and external construction.

## Rules

- Components are static unless you put dynamic fragments inside them.
- Prefer `Bind` for direct value-to-fragment rendering.
- Prefer small `Effect` containers for related values read together.
- Use `Door` when code must explicitly replace a region.
- Initialize every `Source`, `Beam`, and `Door` field deliberately.
- Keep state ownership clear: local, shared, or derived.
- Wrap component construction in `<>...</>` when route switches need fresh component state.

## Related

- [Template Syntax](./03-template-syntax.md) for `elem`, placeholders, and component syntax.
- [State](./07-state.md) for `Source`, `Beam`, `Bind`, `Effect`, and derivation.
- [Door](./06-door.md) for explicit dynamic regions.
- [Events](./08-events.md) for handlers attached to component markup.
