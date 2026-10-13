# State

In **Doors**, reactive values come in two public shapes:

- `Source[T]` is writable. It can own an original value or describe a writable derived view of another source.
- `Beam[T]` is read-only. It observes or derives a value without exposing writes.

A `Source` is also a `Beam`, so every source can be read, subscribed to, bound into UI, and routed like a read-only value.

The usual pattern is:

- keep one source of truth
- derive smaller, specific parts
- bind DOM to the smallest specific pieces

> Subscribers (incl. binds, effects) are triggered when the new value is not equal to the previous value for that source or beam.

## Original Sources

Create an original source with `doors.NewSource(...)` when normal `==` equality is enough:

```go
count := doors.NewSource(0)
```

Use `doors.NewSourceEqual(...)` when you need custom equality:

```go
import "reflect"

settings := doors.NewSourceEqual(Settings{}, func(old Settings, new Settings) bool {
	return reflect.DeepEqual(old, new)
})
```

The equality function should return `true` when the values should be treated as equal, which suppresses propagation.

> The `equal` function runs while the source's internal lock is held. It must not panic and must not call back into any source or beam.


## Derived Sources

Use `DeriveSource` when a part of the page should read and update a smaller part of a larger source value.

```go
type Settings struct {
	Units string
	Days  int
}

settings := doors.NewSource(Settings{
	Units: "metric",
	Days:  7,
})

units := doors.DeriveSource(settings,
	func(s Settings) string {
		return s.Units
	},
	func(s Settings, units string) Settings {
		s.Units = units
		return s
	},
)

days := doors.DeriveSource(settings,
	func(s Settings) int {
		return s.Days
	},
	func(s Settings, days int) Settings {
		s.Days = days
		return s
	},
)
```

`get` extracts the derived value. `set` receives the current parent value and the new derived value, then returns the next parent value.

Derived sources compose:

```go
isMetric := doors.DeriveSource(units,
	func(units string) bool {
		return units == "metric"
	},
	func(_ string, metric bool) string {
		if metric {
			return "metric"
		}
		return "imperial"
	},
)
```

Use `doors.DeriveSourceEqual(...)` when you need custom equality for the derived value.

> The `equal` function runs while the beam's internal lock is held. It must not panic and must not call back into any source or beam.

Updating a derived source writes back through its parent. The parent source remains the single stored value; the derived source only describes how to read and replace one piece of it.

## Tab State

Use `doors.TabState` for small per-tab UI state that must survive a reload but does not belong in the URL: an open panel, a draft filter, a wizard step.

It returns a `Source[*T]` whose value is kept by the browser tab. It survives reloads and in-app navigation, back and forward keep the latest value, other tabs do not see it, and it never appears in the URL.

The stored state reaches the server only after the page has been rendered, so during the initial render the value is `nil`. Render a placeholder and let the update fill it in:

```gox
elem (w Wizard) Main() {
	~~
	step := doors.TabState[int](ctx, "step")
	~~
	~(step.Bind(elem(v *int) {
		~(if v == nil {
			<div class="skeleton"></div>
		} else {
			~(Steps{Current: *v})
		})
	}))
}
```

After the sync the value is never `nil`: a missing key reads as the zero value. Update with `nil` to remove the key. A non-nil value set before the sync wins over the stored one.

Values are stored as JSON. Keep them small: browsers cap history state size, and the whole tab state travels on every page load and update.

Use `doors.TabStateEqual(...)` when you need custom equality. `ctx` must belong to a **Doors** render or handler.

## Beams

Use a `Beam` when a part of the page only needs a smaller read-only view of state.

```go
longRange := doors.DeriveBeam(settings, func(s Settings) bool {
	return s.Days > 7
})
```

Use `doors.DeriveBeamEqual(...)` when you need custom equality for the derived value.

> The `equal` function runs while the beam's internal lock is held. It must not panic and must not call back into any source or beam.

This is one of the main ways **Doors** keeps updates small. If only `Units` changes, a beam or source derived from `Days` can stay unchanged and the fragment using it does not need to rerender.

> Sources and beams are not limited to one page instance. They can be local to one page, shared across a session, or used even more broadly.

## Render

Four rendering strategies drive content from reactive values:

| Strategy | What it does | Works on |
| --- | --- | --- |
| `Bind` | Rerenders one fragment on every value change | `Beam` and `Source` |
| `Effect` | Rerenders the closest dynamic parent when any read value changes | `Beam` and `Source` |
| `RouteBeam` | Picks one of several read-only views based on the value | `Beam` and `Source` |
| `Route` | Picks one of several writable views based on the value | `Source` |

### Bind

`Bind` creates a dynamic fragment that rerenders whenever the value changes:

```gox
<>
	~(counter.Bind(elem(v int) {
		<span>~(v)</span>
	}))
</>
```

> Unmounting or updating a dynamic parent cancels old subscriptions inside it automatically.

Under the hood, `Bind` is the common shorthand for this pattern:

```gox
type CounterView struct {
	counter doors.Beam[int]
	body    doors.Door
}

elem (c *CounterView) Main() {
	~~
	c.counter.Sub(ctx, func(ctx context.Context, v int) bool {
		c.body.Inner(ctx, v)
		return false
	})
	~~

	~>(c.body) <span></span>
}
```

Reach for the manual form when you need explicit control over the door, the subscription lifecycle, or the update strategy.

### Effect

`Effect` reads a value directly inside a dynamic subtree and rerenders that subtree when the value changes. It feels more React-like than passing values through a `Bind` callback.

```gox
<>
	~>(new(doors.Door)) <section>
		~~
		settings, _ := settingsBeam.Effect(ctx)
		~~
		<span>Days: ~(settings.Days)</span>
	</section>
</>
```

Especially useful when the same DOM subtree depends on multiple reactive values:

```gox
type SearchView struct {
	query doors.Beam[string]
	page  doors.Beam[int]
}

elem (v *SearchView) Main() {
	~>(new(doors.Door)) <div>
		~~
		query, _ := v.query.Effect(ctx)
		page, ok := v.page.Effect(ctx)
		~~
		~(if ok {
			<p>Query: ~(query)</p>
			<p>Page: ~(page)</p>
		})
	</div>
}
```

It is enough to check only the last `ok`. `Effect` fails only when the context was already canceled, so if the last call succeeds, the earlier ones did too.

> Use multiple `Effect` calls when the values come from different parts of the application, such as route state and language settings. Do not split one logical state into many tiny `Source`s just to read each field with its own `Effect`. Keep one source of truth. Derive smaller sources or beams when you want a narrower update surface.

### Routing

Routing shows one of several branches depending on a reactive value. Think of it as a `switch` over a `Source` or `Beam` that keeps running: routes are tried against the current value, the first match is shown, and when the value changes, the branch is switched only if a different one should show.

```go
source.Route(routes...)   // gox.Elem; routes are SourceRoute[T]
beam.RouteBeam(routes...) // gox.Elem; routes are BeamRoute[T], every branch is read-only
doors.Route(routes...)    // gox.Elem; routes the page URL, a Source[doors.Location]
```

`Route` is available on every source, `RouteBeam` on every source and beam, derived ones included. `doors.Route(...)` routes `doors.Router(ctx)`; path models and URL helpers are covered in [Routing](./05-routing.md).

Each call renders one dynamic container, however deeply its routes nest. The container shows the chosen branch's content, or nothing when no route matches.

#### When to Use

Use routing when the value decides **which** view the user sees: an order's status, a member's role, a checkout step, a settings tab, a URL section. A branch that stays chosen is not rendered again: it reads later values through the `Source` or `Beam` it receives, so inputs, open panels, and component state inside survive the update. `RouteBind` is the one leaf that re-renders on purpose; see [Leaf Updates](#leaf-updates).

Reach for something else when:

- one fragment just displays the value: use [Bind](#bind). A `Bind` with an `if` chain also switches views, but it re-renders the fragment on every change.
- one small container reads several values together: use [Effect](#effect).
- a handler or background job decides what to show, not a value: use a [Door](./06-door.md).

#### Typical Use

An order page shows an editor while the order is a draft, tracking once it is placed, and a notice otherwise:

```gox
type Order struct {
	ID     int
	Status Status
	Note   string
}

elem OrderPage(order doors.Source[Order]) {
	~(order.Route(
		doors.RouteMatch(
			func(o Order) bool {
				return o.Status == StatusDraft
			},
			doors.RouteSource(OrderEditor),
		),
		doors.RouteMatch(
			func(o Order) bool {
				return o.Status == StatusPlaced
			},
			doors.RouteBeam(OrderTracking),
		),
		doors.RouteComp[Order](OrderClosed{}),
	))
}
```

- `OrderEditor` receives the writable order source, `OrderTracking` a read-only beam, and `OrderClosed` is a fixed component for every other status.
- The editor writes the note through its source. The order is still a draft, so the editor is not rendered again, and its textarea keeps its text and focus.
- When the editor sets the status to `StatusPlaced`, the draft route stops matching, the editor is unmounted, and tracking takes its place.

#### Leaves and Groups

Routes come in two kinds. **Leaves** render content and always match:

| Leaf | Content |
| --- | --- |
| `RouteComp[T](comp)` | a fixed component |
| `RouteBind(func(v T) gox.Elem)` | rendered from the value, again whenever it changes |
| `RouteSource(func(s Source[T]) gox.Elem)` | receives the writable source |
| `RouteBeam(func(b Beam[T]) gox.Elem)` | receives a read-only beam |

**Groups** hold more routes:

| Group | Matches when |
| --- | --- |
| `RouteMatch(pred, routes...)` | `pred` returns `true` and a child matches |
| `RouteValue(value, routes...)` | the value equals `value` and a child matches |
| `RouteDerive(derive, routes...)` | a child matches the derived value |
| `RouteModel[M](routes...)` | the `Location` decodes into path model `M` and a child matches |

Each group has a read-only twin; see [Read-Only Routes](#read-only-routes).

Routes are tried in order, depth first, and the first leaf reached wins. Since a leaf always matches, it is the fallback of its list: put it last, because routes after it are never tried. Without a leaf at the end, the container is empty when nothing matches.

```gox
<>
	~(role.Route(
		doors.RouteValue(RoleAdmin, doors.RouteComp[Role](AdminDashboard{})),
		doors.RouteValue(RoleSupport, doors.RouteComp[Role](TicketQueue{})),
		doors.RouteComp[Role](CustomerHome{}),
	))
</>
```

Routes below a `RouteDerive` work on the derived source, and routes below a `RouteModel` on the model source. Everywhere else, routes work on the source being routed.

#### Match

`RouteMatch` takes a predicate over the value. Use it for ranges, flags, and conditions on several fields:

```gox
elem QuotaBanner(usage doors.Source[Usage]) {
	~(usage.Route(
		doors.RouteMatch(
			func(u Usage) bool {
				return u.UsedGB >= u.LimitGB
			},
			doors.RouteComp[Usage](QuotaExceeded{}),
		),
		doors.RouteMatch(
			func(u Usage) bool {
				return u.UsedGB*5 >= u.LimitGB*4
			},
			doors.RouteBind(QuotaWarning),
		),
	))
}

elem QuotaWarning(u Usage) {
	<p class="warning">~(u.UsedGB) of ~(u.LimitGB) GB used</p>
}
```

Over the limit the banner shows `QuotaExceeded`, from 80% a warning with live numbers, and below that nothing matches, so the banner is empty.

#### Value

`RouteValue` matches one value. It fits enums: steps, plans, tabs.

```gox
<>
	~(step.Route(
		doors.RouteValue(StepCart, doors.RouteComp[Step](CartReview{})),
		doors.RouteValue(StepShipping, doors.RouteSource(ShippingForm)),
		doors.RouteValue(StepPayment, doors.RouteSource(PaymentForm)),
	))
</>
```

`ShippingForm` receives the step source and moves the checkout on with `step.Update(ctx, StepPayment)` once the address is saved.

`RouteValue` compares with the comparator of the source it routes: `==` for `NewSource`, `DeriveSource`, and `DeriveBeam`, your function for `NewSourceEqual`, `DeriveSourceEqual`, and `DeriveBeamEqual`. A custom comparator lets it match types that are not comparable. A source with a nil comparator never matches a `RouteValue`; route it with `RouteMatch`.

#### Derive

`RouteDerive` hands its children a narrower part of the value. It usually sits inside a `RouteMatch`: the match decides when the branch applies, and the derive decides what the routes inside it work on. The derive function turns the level's source into a derived one, usually with `DeriveSource` (see [Derived Sources](#derived-sources)).

Placed orders get delivery views that work on the delivery alone:

```gox
<>
	~(order.Route(
		doors.RouteMatch(
			func(o Order) bool {
				return o.Status == StatusPlaced
			},
			doors.RouteDerive(
				func(src doors.Source[Order]) doors.Source[Delivery] {
					return doors.DeriveSource(
						src,
						func(o Order) Delivery {
							return o.Delivery
						},
						func(o Order, d Delivery) Order {
							o.Delivery = d
							return o
						},
					)
				},
				doors.RouteMatch(
					func(d Delivery) bool {
						return d.Delivered
					},
					doors.RouteComp[Delivery](DeliveredNotice{}),
				),
				doors.RouteSource(DeliveryTracker),
			),
		),
		doors.RouteComp[Order](OrderSummary{}),
	))
</>
```

- The outer match keeps delivery views to placed orders; every other order gets `OrderSummary`.
- Inside it, routes work on `Delivery`: `DeliveredNotice` once the parcel arrives, `DeliveryTracker` until then.
- `DeliveryTracker` receives a `Source[Delivery]`. When it saves delivery instructions, the write goes through the derived source into the order.

`RouteDerive` has no condition of its own: it matches when one of its children matches. `DeliveryTracker` is a leaf, so without the outer match the derive would match every order and `OrderSummary` would never render. Use a bare `RouteDerive` only where it should take everything that reaches it, such as the last route of a list.

#### Read-Only Routes

The `RouteBeam` method starts a read-only tree: nothing below gets a `Source`, even when the routed value is one. Its routes are `BeamRoute[T]`:

- the leaves `RouteComp`, `RouteBind`, and `RouteBeam`
- the groups `RouteBeamMatch`, `RouteBeamValue`, `RouteBeamDerive`, and `RouteBeamModel`

They work like their writable twins. `RouteBeamDerive` takes a `func(Beam[T1]) Beam[T2]`, usually built with `DeriveBeam`, and `RouteBeamModel[M]` hands its children a `Beam[M]`.

```gox
<>
	~(order.RouteBeam(
		doors.RouteBeamMatch(
			func(o Order) bool {
				return o.Status == StatusPlaced
			},
			doors.RouteBeam(OrderTracking),
		),
		doors.RouteBind(OrderReceipt),
	))
</>
```

Don't confuse the two `RouteBeam`s: the method `x.RouteBeam(routes...)` starts a read-only tree, while the leaf `doors.RouteBeam(render)` hands its content a beam and fits in any tree.

`Route` accepts both kinds of routes. `RouteSource` and the writable groups (`RouteMatch`, `RouteValue`, `RouteDerive`, `RouteModel`) fit only under `Route` and other writable groups. A read-only group inside a `Route` tree makes its own subtree read-only.

#### Path Models

`RouteModel[M]` routes a `Location` through a path model. It matches when the `Location` decodes into `M`, and its children route on the decoded model. A `RouteSource` below receives a `Source[M]`: writing it encodes the model back into the `Location`. Under `doors.Route`, that changes the URL, and the page reroutes without a reload. The model syntax is covered in [Path Models](./05-routing.md#path-models).

```gox
type OrdersPath struct {
	View   OrdersView `/orders:" | :ID"`
	ID     int
	Status string `query:"status"`
}

type OrdersView int

const (
	OrdersList OrdersView = iota
	OrderDetails
)

elem App() {
	~(doors.Route(
		doors.RouteModel[OrdersPath](
			doors.RouteMatch(
				func(p OrdersPath) bool {
					return p.View == OrdersList
				},
				doors.RouteSource(OrderList),
			),
			doors.RouteBind(OrderDetailsPage),
		),
		doors.RouteComp[doors.Location](NotFound{}),
	))
}

elem OrderList(path doors.Source[OrdersPath]) {
	<button (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestPointer) bool {
			path.Mutate(ctx, func(p OrdersPath) OrdersPath {
				p.Status = "open"
				return p
			})
			return false
		},
	})>
		Open orders
	</button>
}
```

- `/orders` shows the list, `/orders/42` the order, and URLs that don't decode, like `/orders/abc` or `/billing`, get `NotFound`.
- **Open orders** writes the model, and the URL becomes `/orders?status=open`. The list route still matches, so `OrderList` stays.
- `OrderDetailsPage` is a `RouteBind`, so it re-renders when `/orders/42` becomes `/orders/43`. A URL change that decodes into an equal model doesn't reach the routes below.

Writes push a browser history entry; to replace the current one, see [Navigation](./09-navigation.md#history-replace).

#### Nesting

Groups nest to any depth, and the whole tree still renders into the one container. Here tabs sit inside a section inside the page:

```gox
type Path struct {
	Section Section `/:"dashboard | settings/:Tab"`
	Tab     string
}

<>
	~(doors.Route(
		doors.RouteModel[Path](
			doors.RouteMatch(
				func(p Path) bool {
					return p.Section == SectionDashboard
				},
				doors.RouteComp[Path](Dashboard{}),
			),
			doors.RouteMatch(
				func(p Path) bool {
					return p.Section == SectionSettings
				},
				doors.RouteBeamDerive(
					func(path doors.Beam[Path]) doors.Beam[string] {
						return doors.DeriveBeam(
							path,
							func(p Path) string {
								return p.Tab
							},
						)
					},
					doors.RouteBeamValue("profile", doors.RouteComp[string](ProfileTab{})),
					doors.RouteBeamValue("team", doors.RouteComp[string](TeamTab{})),
				),
			),
		),
		doors.RouteComp[doors.Location](NotFound{}),
	))
</>
```

Going from `/dashboard` to `/settings/team`, or from one tab to another, replaces the container content once.

The container shows only the chosen leaf, so switching tabs replaces everything in it. When a settings layout should stay put while tabs switch, render the layout from a leaf, for example `doors.RouteSource(SettingsLayout)`, and route the tabs inside it with another `Route` call. That inner call is a container of its own and starts over each time the settings branch is entered.

#### Fall-Through

A group matches only when one of its children matches. If none does, routing moves on to the group's next sibling, as if the group's condition had failed.

In the [Nesting](#nesting) example, `/settings/audit` passes the settings match, but no tab matches `"audit"`. The settings group falls through, so does the model route, and `NotFound` renders. To show a page of the section instead, end the tabs with their own leaf: `doors.RouteComp[string](UnknownTab{})`.

#### Rendering Nothing

When no route matches, the container is empty. To render nothing for some values **and** keep later routes from matching them, give the group an empty leaf, `RouteComp[T](nil)`:

```gox
<>
	~(account.Route(
		doors.RouteMatch(
			func(a Account) bool {
				return a.Enterprise
			},
			doors.RouteComp[Account](nil),
		),
		doors.RouteComp[Account](UpgradeBanner{}),
	))
</>
```

Enterprise accounts see no banner. A group with no children never matches, so `RouteMatch(pred)` alone would fall through and show them `UpgradeBanner`.

#### Leaf Updates

While a branch stays chosen, its leaf decides what a new value does:

- `RouteComp`, `RouteSource`, and `RouteBeam` render once when the branch is entered. Later values reach the content through its `Source` or `Beam`, so update the parts that change with `Bind`, `Effect`, or derived values inside.
- `RouteBind` re-renders its content whenever the value changes, judged by the comparator of its level.

```gox
<>
	~(shipment.Route(
		doors.RouteMatch(
			func(s Shipment) bool {
				return s.Delivered
			},
			doors.RouteBind(DeliveredNote),
		),
		doors.RouteBeam(TrackingMap),
	))
</>

elem TrackingMap(shipment doors.Beam[Shipment]) {
	~(MapWidget{})
	~(shipment.Bind(elem(s Shipment) {
		<p>Last seen in ~(s.City)</p>
	}))
}
```

While the parcel travels, `TrackingMap` renders once: the map stays, and only the last-seen line follows the shipment. `DeliveredNote` is small and holds no state, so re-rendering it on change is fine.

Use `RouteBind` for small views where re-rendering is cheap. Use `RouteSource` or `RouteBeam` when the content holds state, such as inputs or its own sources, or is expensive to render.

#### Sticky Branches

A chosen branch stays mounted while the conditions on its path keep holding, even if an earlier route would now match too. Updates that keep it chosen don't re-render it, so a form inside keeps what the user typed:

```gox
<>
	~(ticket.Route(
		doors.RouteMatch(
			func(t Ticket) bool {
				return !t.Closed
			},
			doors.RouteSource(ReplyForm),
		),
		doors.RouteComp[Ticket](ClosedNotice{}),
	))
</>
```

While the agent types a reply, new messages and reassignments keep updating `ticket`. The ticket is still open, so `ReplyForm` is not rendered again and the draft stays.

When someone closes the ticket, the condition fails and the container picks again from the top, as on the first render. Leaving a branch unmounts its content: its hooks and subscriptions end, `OnClean` callbacks run, and `ClosedNotice` shows.

When an earlier route must win whenever it matches, write the later conditions so they exclude it. Leaves are the exception to stickiness, covered next.

#### Fallbacks Give Way

While a leaf is showing, updates re-check the routes before it in its list, and the first one that matches takes over. A fallback is a placeholder until something more specific applies:

```gox
<>
	~(query.Route(
		doors.RouteMatch(
			func(q string) bool {
				return len(q) >= 3
			},
			doors.RouteBeam(SearchResults),
		),
		doors.RouteComp[string](SearchTips{}),
	))
</>
```

Tips show until the query reaches three characters, then the results take over. `SearchResults` follows further typing through its beam; deleting back below three characters ends its branch and the tips return.

A `RouteDerive` showing a leaf of its own list counts as a fallback too, so it gives way to its earlier siblings as well. `RouteMatch`, `RouteValue`, and `RouteModel` don't pass this on: while their condition holds, they stay, whatever they show.

#### Equal Derived Values

A derive level passes an update down only when the derived value changes by its comparator, the same rule beams use to suppress equal updates. While it stays equal, no predicate below runs and nothing below re-renders.

```gox
<>
	~(workspace.RouteBeam(
		doors.RouteBeamDerive(
			func(ws doors.Beam[Workspace]) doors.Beam[Role] {
				return doors.DeriveBeam(
					ws,
					func(w Workspace) Role {
						return w.Role
					},
				)
			},
			doors.RouteBeamValue(RoleOwner, doors.RouteComp[Role](OwnerDashboard{})),
			doors.RouteBeamValue(RoleMember, doors.RouteComp[Role](MemberDashboard{})),
			doors.RouteComp[Role](GuestHome{}),
		),
	))
</>
```

A workspace's unread and online counters change all the time, but the role stays equal, so the role routes are not even checked. Derive the narrowest value a subtree routes on.

With `DeriveSourceEqual` or `DeriveBeamEqual`, your comparator decides what counts as a change: a change it calls equal stays hidden from every route below, even one that would now pick another branch. A nil comparator passes every update down.

#### Details

- The derive functions of `RouteDerive` and `RouteBeamDerive` run once each time the container renders, for every derive in the tree, shown or not. Keep them to building the derived source or beam.
- `RouteComp` renders the same component value each time its branch is entered. Use a [disposable component](./08-components.md#disposable-components) when it should start fresh on every entry.

## Update

Use `Update` when you already know the next value:

```go
units.Update(ctx, "imperial")
```

Use `Mutate` when the new value naturally depends on the old one:

```go
days.Mutate(ctx, func(days int) int {
	return days + 1
})
```

When several fields must change as one logical update, mutate the original source directly.

`Update` and `Mutate` return a completion channel; the return value is optional to use and most code ignores it. It receives `nil` when propagation completes, or `context.Canceled` when a newer update supersedes it, then closes. It closes without a value when no propagation happens: the update is suppressed as equal or there are no subscribers.

The channel is useful when completion itself matters, especially for backpressure. For example, if updates arrive very quickly, waiting on it lets a producer send the next state only after the previous one finished propagating.

Never wait on it during render, in hooks, or in beam subscriptions. When you need to wait for propagation, do it in a goroutine of your own with `doors.DetachedContext(ctx)`, or inside `doors.Go(...)` when the wait should follow a rendered subtree.

If that work should outlive the current dynamic owner, use `doors.InstanceContext(ctx)`.

> For reference types such as slices, maps, pointers, or mutable structs, replace the stored value instead of mutating it in place. Doors compares committed values; in-place mutation can make old and new state indistinguishable.

## Read

Reading and subscribing work from any context. With a **Doors** context, such as the `ctx` you get in render code, handlers, subscriptions, or `doors.Go(...)`, subscriptions are owned by the surrounding dynamic content. With a context outside an instance (session or background context), a subscription lives until that context is canceled.

Updating a `Source` can be done from any context.

### Read

```go
value, ok := beam.Read(ctx)
```

Use `Read(ctx)` when you want the value that is consistent with the current **Doors** render/update cycle.

### Get

```go
value := beam.Get()
```

`Get()` returns the latest stored value without using a render context.

Use it when you want the current value directly. It is available on sources and beams. Do not use it when render consistency matters.

### Sub

```go
ok := beam.Sub(ctx, func(ctx context.Context, value T) bool {
	return false
})
```

`Sub` calls the callback immediately with the current value, then again on later updates.

The subscription ends when:

- your callback returns `true`
- the owning dynamic parent is unmounted
- for a context outside an instance (session or background context) — when the context is canceled

### ReadAndSub

```go
value, ok := beam.ReadAndSub(ctx, func(ctx context.Context, value T) bool {
	return false
})
```

This returns the current value first, then subscribes to future updates. The callback is for later updates only.

### Watcher

`Watch` is the low-level subscription API behind the helpers above. Most app code should use `Sub` or `ReadAndSub`.

## Consistency

The most important state guarantee in **Doors** is consistency.

During one render/update cycle, a Door subtree sees one coherent view of a `Source` and all sources and beams derived from it. A parent and its children do not see different versions halfway through the same render.

In practice, this means beam-driven rendering stays predictable even when several parts of the page are updating at once.

### Read-Freeze Semantics

`Read` and `Sub` capture the observable value for the current render cycle. After calling either, mutating the same source in the same cycle won't be visible until the next cycle.

This matters most in the page factory passed to `doors.NewApp`. If you need to redirect or adjust the initial location, use `Get()` to inspect the current value (it does not freeze), call `Update` or `Mutate` on `doors.Router(ctx)`, and place any `Sub` or `ReadAndSub` calls **after** the mutation.

## Skipping

By default, a `Source` is allowed to skip stale in-flight propagation.

That is usually what you want for UI state. If a newer value arrives before an older one finishes propagating, **Doors** prefers getting the UI to the latest useful state instead of insisting that every intermediate value must finish syncing.

If you need in-progress propagation to keep going even when a newer update arrives, create the source with a NoSkip constructor:

```go
source := doors.NewSourceNoSkip(0)
```

For custom equality, use:

```go
source := doors.NewSourceEqualNoSkip(init, equal)
```

NoSkip does not bypass equality. Equal updates are still suppressed; NoSkip only changes what happens after a committed update is already propagating.

Use NoSkip when an already-committed update should keep propagating even if a newer value arrives before it completes.

## Rules

- Prefer one `Source` plus derived sources and beams over many unrelated mutable sources.
- Keep state small and structural. Store IDs, filters, settings, selections, and route values.
- Use `Read(ctx)` when you need the render-consistent value.
- Use `Get()` only when you explicitly want the latest stored value outside render guarantees.
- Return a fresh value from `Mutate` or pass a fresh value to `Update` instead of mutating reference-type state in place.
- Use NoSkip constructors only when already-committed in-progress updates must finish.
- Use `TabState` for per-tab state that must survive a reload without going into the URL.

## Example

```gox
type SearchState struct {
	Query string
	Page  int
}

type Search struct {
	query doors.Source[string]
	page  doors.Source[int]
}

func NewSearch(state doors.Source[SearchState]) *Search {
	return &Search{
		query: doors.DeriveSource(state,
			func(s SearchState) string { return s.Query },
			func(s SearchState, query string) SearchState {
				s.Query = query
				return s
			},
		),
		page: doors.DeriveSource(state,
			func(s SearchState) int { return s.Page },
			func(s SearchState, page int) SearchState {
				s.Page = page
				return s
			},
		),
	}
}

elem (s *Search) Main() {
	<input
		type="search"
		(doors.AInput{
			On: func(ctx context.Context, ev doors.RequestEvent[doors.InputEvent]) bool {
				value := ev.Event().Value
				s.query.Update(ctx, value)
				return false
			},
		})/>

	<button
		(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestPointer) bool {
				s.page.Mutate(ctx, func(page int) int {
					return page + 1
				})
				return false
			},
		})>
		Next page
	</button>

	~(s.query.Bind(elem(q string) {
		<p>Query: ~(q)</p>
	}))

	~(s.page.Bind(elem(page int) {
		<p>Page: ~(page)</p>
	}))
}
```

This keeps one source of truth while letting the query and page fragments update independently.

See [Components](./08-components.md) for how to structure component fields around local, shared, and derived state.
