# Navigation

In **Doors**, navigation usually happens in one of two ways:

- declaratively with `doors.ALink`
- programmatically by updating the current page's path `doors.Source[Path]`

Both use the same path model rules described in [Path Model](./04-path-model.md).

## ALink

Use `doors.ALink` for normal links.

It does two things:

- builds a real `href` from a path model
- adds **Doors** navigation behavior on top

```gox
<>
	~>doors.ALink{
		Model: Path{
			Home: true,
		},
	} <a>Home</a>
</>
```

`Model` is required.

## Link Behavior

`ALink` automatically chooses between same-page navigation and a normal page load.

If the target model has the same registered model type as the current page:

- the current page path source is updated
- the URL is updated
- the page rerenders without a full reload

If the target model belongs to a different registered model type:

- the generated `href` still works
- the browser performs a normal page load

That means you can usually write links in terms of models and let **Doors** decide how to navigate.

`ALink` always sets a real `href`, so the link remains a valid browser link even without the dynamic behavior.

## Fragment

Use `Fragment` to append `#...` to the generated URL:

```gox
<>
	~>doors.ALink{
		Model: Path{Docs: true},
		Fragment: "api",
	} <a>API</a>
</>
```

## Active

Use `Active` when a link should reflect the current location.

```gox
<>
	~>doors.ALink{
		Model: Path{
			Dashboard: true,
			ID:        f.id,
		},
		Active: doors.Active{
			Indicator: doors.IndicatorOnlyAttr("aria-current", "page"),
		},
	} <a>Current page</a>
</>
```

When the current location matches, **Doors** applies the given indicators to the link element.

If `Active.Indicator` is empty, there is no active-link behavior.

### Path

`Active.PathMatcher` controls how the path is compared.

Options:

- `doors.PathMatcherFull()`
- `doors.PathMatcherStarts()`
- `doors.PathMatcherSegments(i...)`

If you do not set one, **Doors** defaults to full-path matching.

### Query

`Active.QueryMatcher` controls query-string matching.

Available matchers:

- `doors.QueryMatcherIgnoreSome(params...)`
- `doors.QueryMatcherIgnoreAll()`
- `doors.QueryMatcherSome(params...)`
- `doors.QueryMatcherIfPresent(params...)`

The query matchers are applied in order, then **Doors** compares any remaining query parameters.

In practice:

- `IgnoreSome` removes some keys from comparison
- `IgnoreAll` ignores all remaining query parameters
- `Some` compares only the listed keys at that step
- `IfPresent` compares listed keys only when they exist

Helpers:

- `doors.QueryMatcherOnlyIgnoreSome(params...)`
- `doors.QueryMatcherOnlyIgnoreAll()`
- `doors.QueryMatcherOnlySome(params...)`
- `doors.QueryMatcherOnlyIfPresent(params...)`


### Fragment Match

Set `Active.FragmentMatch` when `#...` should also be part of active matching.

By default, fragments are ignored for active-link matching.

## Path Source

For programmatic navigation, update the current page's path source directly.

The `doors.Source[Path]` passed into your page is not just state. It is also the current route model for that page.

When you update or mutate that source:

- **Doors** re-encodes the path model
- the browser location is updated
- the page reacts as if the user navigated there

This is the right approach when navigation belongs to a button, wizard step, form flow, or other interaction that is not naturally an anchor tag.

## Mutate

Use `Mutate` when you want to preserve part of the current path and change the rest:

```gox
type App struct {
	path doors.Source[Path]
}

elem (a *App) goToCity(cityID int) {
	<button
		(doors.AClick{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				a.path.Mutate(ctx, func(p Path) Path {
					p.Selector = false
					p.Dashboard = true
					p.ID = cityID
					return p
				})
				return false
			},
		})>
		Open city
	</button>
}
```

## Update

Use `Update` when you already know the full target model:

```go
a.path.Update(ctx, Path{
	Dashboard: true,
	ID:        cityID,
})
```

## Actions

When an `ALink` upgrades to same-model dynamic navigation, it participates in
the same request pipeline as other **Doors** attributes.

That is why it supports:

- `Scope`
- `Indicator`
- `Before`
- `After`
- `OnError`

One useful special case: if `OnError` is left `nil` on a dynamic `ALink`,
**Doors** falls back to a location reload.

For the action types themselves, see [Actions](./12-actions.md).
