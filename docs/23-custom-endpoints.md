# Custom Endpoints

`doors.NewHook(ctx, resource)` returns a URL path served by your handler for as long as the content around it lives. Use it when browser code needs its own HTTP endpoint: a React island with an API, an `EventSource`, a `WebSocket`, a third-party widget.

```go
url, _ := doors.NewHook(ctx, doors.ResourceHandler(handler))
```

The second result is `false` only when the content is already cleared. That render never reaches the browser, so it is safe to ignore.

## Start

A JSON API behind one endpoint, routed with `http.ServeMux`:

```go
func (d *Dashboard) api() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(d.users.List())
	})
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		u, ok := d.users.Get(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(u)
	})
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		d.users.Add(u)
		w.WriteHeader(http.StatusCreated)
	})
	return mux
}
```

The island module, `web/dashboard/index.tsx`, takes the endpoint URL:

```tsx
import { createRoot } from "react-dom/client"
import { useEffect, useState } from "react"

export function mount(el: HTMLElement, api: string) {
	createRoot(el).render(<Users api={api} />)
}

function Users({ api }: { api: string }) {
	const [users, setUsers] = useState<User[]>([])
	useEffect(() => {
		fetch(`${api}/users`).then(r => r.json()).then(setUsers)
	}, [api])
	return <ul>{users.map(u => <li key={u.id}>{u.name}</li>)}</ul>
}
```

Bundle it with your npm toolchain and embed the output, so the Go binary does not need `node_modules`:

```sh
npx esbuild web/dashboard/index.tsx --bundle --format=esm --minify --outfile=web/dashboard/dist/dashboard.js
```

```go
//go:embed web/dashboard/dist/dashboard.js
var dashboardJS []byte
```

Register the module in the page head. The import map is sent with the page, so this keeps `import("dashboard")` working even when the island first appears in a later update:

```gox
<head>
	<link
		rel="modulepreload"
		href=(dashboardJS)
		specifier="dashboard">
</head>
```

Pass the URL to the island with `data:`:

```gox
elem (d *Dashboard) Main() {
	<div id="app"></div>
	<script data:api=({
		url, _ := doors.NewHook(ctx, doors.ResourceHandler(d.api().ServeHTTP))
		return url
	})>
		const { mount } = await import("dashboard")
		mount(document.getElementById("app"), $data("api"))
	</script>
}
```

The endpoint is same-origin, so `fetch` sends the session cookie and the default CSP allows the call. For `specifier` see [Modules](./15-javascript.md#modules).

Without React, any script calls it the same way:

```gox
<script data:api=(api)>
	const r = await fetch(`${$data("api")}/users/42?fields=name`)
	console.log(await r.json())
</script>
```

## Requests

- The returned path has no trailing slash. Append `/` and further segments: `url`, `url/`, `url/users/42`, `url?q=1`.
- The handler sees the part after the endpoint path as `r.URL.Path`, always starting with `/`. `url` and `url/` both give `/`. Escaped segments stay in `r.URL.RawPath`, the query is kept, and `r.RequestURI` is the original.
- The `Doors-Hook-Prefix` request header holds the endpoint path.
- Any method is accepted. Route on it in your handler, as in `"POST /users"`.
- Requests run concurrently. Synchronize shared state yourself.
- Values from `doors.WithValue` are on `r.Context()`, see [Context Values](./21-configuration.md#context-values).
- Use `doors.ResourceHook(func(ctx, w, r) bool)` when the handler needs the **Doors** `ctx`. Return `false` to keep the endpoint; `true` removes it after that request.

### Redirects

A redirect to `/files/` sends the browser to the site root, not the endpoint. Build the location from `Doors-Hook-Prefix`:

```go
mux.HandleFunc("GET /old", func(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.Header.Get("Doors-Hook-Prefix")+"/new", http.StatusFound)
})
```

`http.ServeMux` redirects on its own in two cases, and both leave the endpoint: a request for `/files` when only `/files/` is registered, and a path with `//` or `..`. Register the exact paths your client calls.

## Lifetime

The endpoint lives as long as the content it was created in. Create it in a component for a widget-scoped API, or in the page root for a page-wide one.

Once the content is cleared, or the page or session ends, new requests get `404` or `410`. Treat both as "the page is gone". Requests already running are not interrupted; the `ctx` of a `ResourceHook` handler is canceled at that moment, so long-lived handlers watch `ctx.Done()`.

## Streaming

The response writer supports `http.Flusher`, so server-sent events work:

```gox
<script data:events=({
	url, _ := doors.NewHook(ctx, doors.ResourceHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		fl := w.(http.Flusher)
		for {
			select {
			case <-ctx.Done():
				return false
			case <-r.Context().Done():
				return false
			case p := <-d.prices:
				fmt.Fprintf(w, "data: %s\n\n", p)
				fl.Flush()
			}
		}
	}))
	return url
})>
	const es = new EventSource($data("events"))
	es.onmessage = e => console.log(e.data)
	$sys.clean(() => es.close())
</script>
```

For updates rendered by **Doors**, use Door updates or `doors.ActionEmit` with `$on(...)`. Streaming is for data the client consumes itself.

## WebSocket

The response writer supports `http.Hijacker`, so a WebSocket library can upgrade on an endpoint. With `github.com/gorilla/websocket`:

```gox
<script data:chat=({
	url, _ := doors.NewHook(ctx, doors.ResourceHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return false
		}
		defer c.Close()
		stop := context.AfterFunc(ctx, func() { c.Close() })
		defer stop()
		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				return false
			}
			c.WriteMessage(mt, msg)
		}
	}))
	return url
})>
	const proto = location.protocol === "https:" ? "wss" : "ws"
	const ws = new WebSocket(`${proto}://${location.host}${$data("chat")}/room/1`)
	ws.onmessage = e => console.log(e.data)
	$sys.clean(() => ws.close())
</script>
```

Close the connection when `ctx` ends; **Doors** does not close it for you. WebSocket upgrades need HTTP/1.1, which browsers use for `ws://` and `wss://`. `app.Use(...)` middleware that wraps the response writer must keep `http.Hijacker` and `http.Flusher`.

## `$hook` or a custom endpoint

Use `AHook` or `ARawHook` with `$hook` to enhance the page with JavaScript where `doors.A*` event attrs are not enough, such as drag-and-drop processing or scroll position sync. Calls keep **Doors** scopes, indication, and error handling. See [Hooks](./15-javascript.md#hooks).

Use a custom endpoint when the client brings its own code and speaks plain HTTP:

- a React or other framework island with its own API
- a third-party widget or library that needs a URL
- `EventSource` and `WebSocket` connections
- file uploads and downloads driven by `fetch`, with progress or streaming
- web workers

For static files the client loads, such as a stylesheet, image, or font, use a cached URL from `doors.NewResource` instead, see [URL in code](./14-resources.md#url-in-code).

## Rules

- route sub-paths with `http.ServeMux` on one `ResourceHandler`
- use `ResourceHook` and watch `ctx.Done()` for streams and sockets
- treat `404` and `410` as "the page is gone"
- enhance page logic with `$hook`; give self-contained clients a custom endpoint
