# Navigation

In **Doors**, in-app navigation happens in one of two ways:

- declaratively with `doors.ALink`
- programmatically by updating the `Source` received from `RouteModel`

Both update the current page instance's location source. The router (see [Routing](./05-routing.md)) reacts by re-matching routes and updating the view in place. No full reload is involved.

## ALink

`doors.ALink` is a real anchor with **Doors** behavior on top.

```gox
<>
	~>doors.ALink{
		Model: Path{
			Section: SectionHome,
		},
	} <a>Home</a>
</>
```

`Model` is required. It accepts:

- a path-model struct
- a `doors.Location` value
- a custom type implementing `doors.LocationEncoder`

The same three forms work in `ActionLocationAssign`, `ActionLocationReplace`, and `doors.NewLocation`.

## LocationEncoder

`doors.LocationEncoder` is a one-method interface for custom navigation models:

```go
type LocationEncoder interface {
	Encode() (doors.Location, error)
}
```

Implement it on a domain type when you have one-off URL shapes that don't justify a full path-model struct, or when an existing type already knows its own URL form. It pairs naturally with `RouteDerive` on `Location` for the matching side — together they round-trip a custom value through the URL.

```go
type CustomRoute struct {
	ID  string
	Tab string
}

func (r CustomRoute) Encode() (doors.Location, error) {
	q := url.Values{}
	if r.Tab != "" {
		q.Set("tab", r.Tab)
	}
	return doors.Location{
		Segments: []string{"custom", r.ID},
		Query:    q,
	}, nil
}
```

`CustomRoute{ID: "a", Tab: "info"}` now works wherever a model is accepted.

## Link Behavior

ALink always navigates dynamically. A normal click is intercepted, a hook updates the current instance's location source, and the page re-routes in place. `Route(...)` matches the new `Location` — the active route swaps if a different route now matches; otherwise the existing fragment stays and reacts to the new value via its `Source` or `Beam`.

The element is also a real anchor — `href` is set from the encoded model. That's there for browser features (middle-click, Cmd-click, "open in new tab", "copy link"), not as a fallback for clicks. Each of those opens the URL in a fresh browser context, handled like any other initial request.

On a normal dynamic click, the client updates the browser URL to the link `href`
**before** the hook request runs. If an error occurs:

- **Canceled** (blocked by scope) or **Not found (404)** (hook removed on server):
  The URL reverts to its previous value. Silent — `OnError` does not run.
- **Gone (410)** (instance stopped): The page reloads to the new URL. `OnError`
  does not run; the reload supersedes it.
- **Network error** or **Server error (5xx)**: The hook request likely never
  reached the server, so the URL reverts. `OnError` runs — you can show an error
  message (for example via `ActionEmit` and a `$on` handler). If the request did
  reach the server despite the error, the server will restore the URL to the new
  path on reconnect.
- **Bad request (400)** or other unexpected errors: The URL stays at the new
  path. `OnError` runs. Rare, usually from third-party proxies.

## Fragment

Use `Fragment` to append `#...` to the generated URL:

```gox
<>
	~>doors.ALink{
		Model: Path{Section: SectionDocs},
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
			Section: SectionDashboard,
			ID:      f.id,
		},
		Active: doors.Active{
			Indicator: doors.IndicateAttr("aria-current", "page"),
		},
	} <a>Current page</a>
</>
```

When the current location matches, **Doors** applies the indicator to the link element.

If `Active.Indicator` is empty, there is no active-link behavior.

### Path

`Active.PathMatcher` controls how the path is compared.

- `doors.PathMatcherFull()` — full path (default)
- `doors.PathMatcherStarts()` — current path starts with the link path
- `doors.PathMatcherSegments(i...)` — only listed segment indexes (zero-based)

### Query

`Active.QueryMatcher` controls query-string matching. The matchers are applied in order, then **Doors** compares any remaining query parameters.

- `doors.QueryMatcherIgnoreSome(params...)` — drop the listed keys from comparison
- `doors.QueryMatcherIgnoreAll()` — drop all remaining keys
- `doors.QueryMatcherSome(params...)` — compare only the listed keys at this step
- `doors.QueryMatcherIfPresent(params...)` — compare the listed keys only when present

Chain matchers with `.And(...)` when several steps are needed:

```gox
Active: doors.Active{
	QueryMatcher: doors.QueryMatcherSome("mode").And(doors.QueryMatcherIgnoreAll()),
	Indicator:    doors.IndicateClass("active"),
}
```

This example compares only `mode` and ignores every other query parameter.

### Fragment Match

`Active.FragmentMatch` includes `#...` in matching. Off by default.

## Hook Fields

ALink is a dynamic hook, so it supports the same request-lifecycle fields as event attrs:

- `Scope` — request scheduling and de-duplication. See [Scopes](./10-scopes.md).
- `Indicator` — UI feedback while the request is in flight. See [Indication](./11-indication.md).
- `Before` — actions to run before the request.
- `After` — actions to run after a successful request.
- `OnError` — actions to run on failure. Defaults to `ActionLocationReload{}` if nil.

For action types, see [Actions](./12-actions.md).

## Programmatic

For navigation that doesn't fit an anchor (button, wizard step, form flow), update that `Source` directly.

The source you get from `doors.RouteModel(...)` is a typed view of the current URL. Updating it encodes the new path model back into the URL, like an `ALink` click.

For raw locations, update `doors.Router(ctx)` directly.

### Update

When you already know the full target model:

```go
a.path.Update(ctx, Path{
	Section: SectionDashboard,
	ID:      cityID,
})
```

### Mutate

When you want to preserve part of the current path and change the rest:

```gox
type App struct {
	path doors.Source[Path]
}

elem (a *App) goToCity(cityID int) {
	<button
		(doors.AClick{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				a.path.Mutate(ctx, func(p Path) Path {
					p.Section = SectionDashboard
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

### History Replace

A programmatic update pushes a new browser history entry by default. To replace the current entry instead (like `history.replaceState`), wrap the context with `doors.HistoryReplaceContext`:

```go
a.path.Update(doors.HistoryReplaceContext(ctx), Path{
	Section: SectionDashboard,
	ID:      cityID,
})
```

It works with any URL-backed source — `doors.Router(ctx)`, a `RouteModel` source, or a derived route source — and with both `Update` and `Mutate`. Browser back then skips the replaced entry.
