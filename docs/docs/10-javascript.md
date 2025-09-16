# JavaScript

For on-page interactivity, extended event handling, and integration with frontend frameworks.

## `$d` Reference

The magic `$d` variable provides access to framework features:

* `$d.data(name: string): any`  
  Retrieve data provided via `doors.AData`.

* `$d.hook(name: string, arg: any): Promise<any>`  
  Call a Go hook defined by `doors.AHook`, serialize/deserialize input/output.

* `$d.rawHook(name: string, arg: any): Promise<Response>`  
  Call a Go hook (`doors.AHook` or `doors.ARawHook`) and get the raw `Response`.

* `$d.HookErr`  
  Reference to the error class thrown by hooks. You can check errors with  
  `if (err instanceof $d.HookErr) { … }`.

* `$d.on(name: string, handler: (arg: any, err?: $d.HookErr) => any): void`  
  Register a handler callable from Go with `ActionEmit` or `Call`.  
  (No async support in the handler itself.)

* `$d.G: { [key: string]: any }`  
  Persistent global object to share state between scripts.

* `$d.clean(handler: () => void | Promise<void>): void`  
  Register a cleanup function to run when the script is removed.

## Script Component

The Script Component integrates inline scripts with framework features:

* Provides the magic `$d` variable
* Converts inline scripts into loadable, cacheable resources
* Minifies by default
* Compiles TypeScript (when `type="application/typescript"` or `"text/typescript"`)
* Wraps code in an async IIFE, so you can safely `await` and avoid polluting global scope

```templ
@doors.Script() {
	<script>
		console.log("hello world!")
	</script>
}
```

> ⚠️ `type="module"` is not supported in `doors.Script()`.  
> For ES modules, use **Imports** (see ref/imports article) and load via `await import("specifier")`.

### Variants

* `doors.Script()` — public, cacheable script resource  
* `doors.ScriptPrivate()` — session-protected, cached script resource  
* `doors.ScriptDisposable()` — session-protected, not cached, avoids leaks with dynamic scripts  

## Pass Data To JavaScript

Use **doors.AData** to expose server-side values:

```go
type AData struct {
	// Name of the data entry, accessed in JS via $d.data(name).
	// Required.
	Name string

	// Value to expose. Marshaled to JSON.
	// Required.
	Value any
}
```

Example:

```templ
@doors.Script() {
	@doors.AData{
		Name:  "user_profile",
		Value: user, // any Go struct
	}
	<script>
		const user = $d.data("user_profile") // parsed JSON
		console.log(user.name)
	</script>
}
```

## Call Go from JavaScript

### Structured Hook

`doors.AHook` handles Go hooks with automatic JSON (un)marshaling:

```go
type AHook[T any] struct {
	// Name of the hook to call from JavaScript via $d.hook(name, ...).
	// Required.
	Name string
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running.
	// Optional.
	Indicator []Indicator
	// Backend handler for the hook.
	// Receives typed input (T, unmarshaled from JSON) through RHook,
	// and returns any output which will be marshaled to JSON.
	// Should return true when the hook is complete and can be removed.
	// Required.
	On func(ctx context.Context, r RHook[T]) (any, bool)
}
```

#### Example

```templ
@doors.AHook[string]{
	Name: "length",
	On: func(ctx context.Context, r doors.RHook[string]) (any, bool) {
		return len(r.Data()), false
	},
}

@doors.Script() {
	<script>
		const length = await $d.hook("length", "hello!") 
		console.log(length) // 6
	</script>
}
```

> ⚠️ On errors, hooks throw `$d.HookErr`. Handle exceptions explicitly.

### Raw Hook

`doors.ARawHook` provides low-level access to the HTTP request:

```go
type ARawHook struct {
	// Hook name, called with $d.rawHook(name, arg).
	// Required.
	Name string

	// Backend handler with direct access to body/form.
	// Required.
	On func(ctx context.Context, r RRawHook) bool

	// Optional scope control.
	Scope []Scope

	// Optional visual indicators.
	Indicator []Indicator
}
```

### Data Conversion Rules

When calling `$d.hook` or `$d.rawHook`, the `arg` is converted to the request body:

* `undefined` → no body  
* `FormData` → multipart body  
* `URLSearchParams` → form-urlencoded body  
* `Blob` → raw blob with `Content-Type`  
* `File` / `ReadableStream` → octet-stream  
* Any other value → JSON  

> ℹ️ `$d.rawHook` (JS) is not related to `doors.ARawHook` (Go).  

## Call JavaScript from Go

You can invoke JS handlers registered with `$d.on`.  

⚠️ Handlers must be synchronous. For long tasks, use hooks to report back.

### Register a Handler

```templ
@doors.Script() {
	<script>
		$d.on("alert", (msg) => {
			alert(msg)
			return true
		})
	</script>
}
```

### Invoke a Handler

From Go, use `ActionEmit` or the Call API:

```go
// Fire-and-forget
doors.Call(ctx, ActionEmit{Name: "alert", Arg: "Hello!"})

// Await a return value (best-effort cancelation)
ch, cancel := doors.XCall[string](ctx, ActionEmit{Name: "alert", Arg: "Hello!"})
defer cancel()
res := <-ch
if res.Err == nil {
	log.Println("Handler returned:", res.Ok)
}
```

### Call Lookup Rules

Handler resolution is scoped to the closest dynamic parent (Door).  
When invoking:

* A hook can only call handlers **in its parent door or above**  
* Inner handlers shadow outer ones if names collide  

## Clean Up

Use `$d.clean` for cleanup when a script is removed:

```js
<script>
	const ctrl = new AbortController()
	$d.clean(() => ctrl.abort())
	// start background process
</script>
```

## `HookErr` Reference

Hook errors thrown in JS are instances of `$d.HookErr` (subclass of `Error`). If you need 

Error reasons:

* Canceled by scope  *(suppressed in event hooks)*
* Not Found  - hook expired, 404 code  *(suppressed in event hooks)*
* Unauthorized -  cause page reload (401 and 410 codes)
* Bad Request -  not expected to occur (400 code)
* Other - other 4XX code (not used by framework, if you issued)
* **Network Error **
* **Server Error  (5XX code)**
* Capture Error - not expected to occur (means js side event handling issue)

### Fields

* `status: number | undefined` — HTTP status code (if available)  
* `cause: Error | undefined` — original error object  
* `kind: "canceled" | "unauthorized" | "not_found" | "network" | "server" | "bad_request" | "capture"`
* `message: string` ` 

### Methods

Convenience type checks:

* `canceled(): bool`  
* `unauthorized(): bool`  
* `notFound(): bool`  
* `network(): bool`  
* `server(): bool`  
* `badRequest(): bool`  
* `capture(): bool`  
* `isOther(): bool`
