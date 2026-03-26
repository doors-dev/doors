# Navigation

Doors navigation has two main forms:

- declarative navigation with `doors.AHref`
- programmatic navigation by mutating the path `doors.Source[Path]`

Both are based on the same path model encoding rules from routing.

## `AHref`

`doors.AHref` builds an internal `href` from a path model and adds Doors navigation behavior.

```gox
~>doors.AHref{
	Model: Path{
		Home: true,
	},
} <a>Home</a>
```

Fields:

- `Model`
- `Fragment`
- `Active`
- `StopPropagation`
- `Scope`
- `Indicator`
- `Before`
- `After`
- `OnError`

`Model` is required. Everything else is optional.

## Dynamic

`AHref` can navigate in two different ways.

If the target model has the same registered model type as the current page source, the link is dynamic:

- the page source is updated
- Doors rerenders the affected fragments
- the browser URL is updated without a full page load

If the target model belongs to a different registered model type, the link falls back to plain navigation:

- the generated `href` still works
- the browser performs a normal page load

So `AHref` automatically chooses between same-page navigation and full navigation based on the model type.

## Fragments

Use `Fragment` to append `#...` to the generated URL:

```gox
~>doors.AHref{
	Model: Path{Docs: true},
	Fragment: "api",
} <a>API</a>
```

## Active

Active-link indication is configured through `Active`.

```gox
~>doors.AHref{
	Model: Path{
		Dashboard: true,
		Id:        f.id,
	},
	Active: doors.Active{
		Indicator: doors.IndicatorOnlyAttr("aria-current", "page"),
	},
} <a>Current page</a>
```

When the current location matches the configured rules, Doors applies the provided indicators on the link element.

If `Active.Indicator` is empty, there is no active behavior.

## Path Match

`Active.PathMatcher` controls how the path is compared.

Options:

- `doors.PathMatcherFull()`
- `doors.PathMatcherStarts()`
- `doors.PathMatcherSegments(i...)`

Behavior:

- `Full` compares the full path
- `Starts` checks whether the current path starts with the link path
- `Segments` compares only specific path segment positions

If no path matcher is provided, active matching defaults to full path matching.

## Query Match

`Active.QueryMatcher` controls query-string matching.

Matchers are applied in order.

Available matchers:

- `doors.QueryMatcherIgnoreSome(params...)`
- `doors.QueryMatcherIgnoreAll()`
- `doors.QueryMatcherSome(params...)`
- `doors.QueryMatcherIfPresent(params...)`

Helpers:

- `doors.QueryMatcherOnlyIgnoreSome(params...)`
- `doors.QueryMatcherOnlyIgnoreAll()`
- `doors.QueryMatcherOnlySome(params...)`
- `doors.QueryMatcherOnlyIfPresent(params...)`

Practical meaning:

- `IgnoreSome` removes some keys from the comparison and continues
- `IgnoreAll` ignores all remaining query parameters
- `Some` compares only the listed keys at that step
- `IfPresent` compares listed keys only when they exist

Doors finishes active matching by comparing any remaining query parameters, so the match stays stable unless you explicitly ignore them.

## Fragment Match

Set `Active.FragmentMatch` to include the fragment in active matching.

By default, fragment identity does not participate.

## Example

```gox
~>doors.AHref{
	Model: Path{
		Dashboard: true,
		Id:        f.id,
	},
	Active: doors.Active{
		PathMatcher:  doors.PathMatcherFull(),
		QueryMatcher: doors.QueryMatcherOnlyIgnoreSome("days"),
		Indicator:    doors.IndicatorOnlyAttr("aria-current", "page"),
	},
} <a>Overview</a>
```

This marks the link active when:

- the path matches fully
- all query params match except `days`

## Path Source

For programmatic navigation, use the page path source directly.

The `doors.Source[Path]` passed into your page or fragment is not only state. It is also the current route model for that page.

When you update or mutate that source:

- Doors re-encodes the path model
- the browser location is updated
- the page reacts as if the user navigated to that route

## Mutation

This is the usual programmatic navigation pattern:

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
					p.Id = cityID
					return p
				})
				return false
			},
		})>
		Open city
	</button>
}
```

This is useful when navigation belongs to a button, a wizard step, a form flow, or some other action that is not naturally an anchor tag.

## Update

You can also replace the whole path model at once:

```gox
a.path.Update(ctx, Path{
	Dashboard: true,
	Id:        cityID,
})
```

Use `Mutate` when you want to preserve and edit the existing path.  
Use `Update` when you already know the full target model.

## Source Rules

Manual path-source navigation works when you are mutating the current page's own path source.

That is the same source that `AHref` uses internally for same-model navigation.

If you want to navigate to a different model type from event handlers, use link navigation or actions instead.

## Choosing

- Use `AHref` for normal links.
- Use `Active` when a link should reflect the current location.
- Mutate the path source when navigation is part of a button or interaction flow.
- Use `Fragment` when the URL should include `#...`.

## Notes

- `AHref` still sets a real `href`, so the link remains valid as a normal browser link.
- `AHref` also supports `Scope`, `Indicator`, `Before`, `After`, and `OnError`, because it participates in the same event pipeline as other Doors attributes.
- Actions-based navigation is covered in the next doc.
