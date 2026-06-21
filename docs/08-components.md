# Components

Components are Go values that implement `gox.Comp` by having a `Main() gox.Elem` method.

There is no virtual DOM in **Doors**. A component is static by default: `Main()` renders once and stays. Re-rendering is explicit and local. `Bind` and `Effect` re-render dynamic fragments, subscriptions can update Doors, and direct Door methods replace only the Door region. The rest of the tree is untouched.

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

## Fields

Keep component fields tied to ownership and reuse. Use an `elem` function for markup that does not need state or methods. Use a struct component when a value must persist after `Main()` returns: state handles, Door handles, dependencies called by handlers, or configuration shared by methods.

Model data with concrete types. If a component needs a content slot, make that slot explicit and document what callers may pass.

When a repeated block has one template shape plus data, make one component per item:

```gox
type ContentBlock struct {
	Title       string
	Description string
	Items       []string
}

elem (b ContentBlock) Main() {
	<section>
		<h2>~(b.Title)</h2>
		<p>~(b.Description)</p>
		<ul>
			~(for _, item := range b.Items {
				<li>~(item)</li>
			})
		</ul>
	</section>
}
```

For many items, prefer an explicit `for` loop that renders one component per item. It keeps wrappers, classes, IDs, and per-item attributes local to each item.

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

Components can be reused across route switches when the same Go value is passed to `.Comp(...)`. Wrap construction in an element when you want fresh state each time the route enters:

```gox
<>
	~(path.Route(
		doors.RouteMatch(func(p Path) bool { return p.Section == SectionSelector }).
			Comp(<>
				~(LocationSelector(applyPlace))
			</>),
		doors.RouteDefaultComp[Path](Dashboard{}),
	))
</>
```

Without the wrapper, the component value can be reused and its fields can keep their previous state.

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
