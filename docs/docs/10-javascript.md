# JavaScript

For on-page interactivity,  extended event handling, and frontend frameworks integration.

## `$d` Reference

Magic variable to access framework features

* `$d.data(name: string)`
  Retrieve data from `doors.AData` attribute
* `$d.hook(name: string, arg: any): Promise<any>`
  Call the hook from the `doors.AHook` or `doors.ARawHook` attributes and parse response
* `$d.rawHook(name: string, arg: any): Promise<Response>`
  Call the hook from the `doors.AHook` or `doors.ARawHook` and get raw response
* `$d.CaptureEr`
  Reference to the error class thrown by hooks, so in the exception, you can check `if(err instanceof $d.CaptureErr)`. A description of CaptureErr can be found in the [Error Handling](./ref/02-error-handling.md) article.
* `$d.on(name string,  handler: (arg: any) => any)`
  Register a handler to be called from Go or on error (checkout [Error Handling](./ref/02-error-handling.md) article)
* `$d.G: { [key: string]: any }`
  Getter for the global persistent object to share data between scripts.
* `$d.clean(handler: () => void | Promise<void>): void`
  Register a handler to be called when the script is removed from the DOM

## 1. Script Component

The Script Component provides integration with framework functions. It:

* **Provides the magic `$d` variable to access framework features**
* Converts an inline script to a loaded and cacheable one
* Minifies (by default)
* Compiles TypeScript (if attribute `type="application/typescript"` or `"text/typescript"` is provided)
* Wraps in an anonymous async function, so you can **use await and do not care about global scope pollution** with your variables

```templ
@doors.Script() {
	<script>
		console.log("hello world!")
	</script>
}
```

> Script type="module" is not supported by `doors.Script()` . If you need modules, include them via Imports (**ref/imports** article) and then require them in the script via `const module = await import("specifier")`
>
> *That's intentional, because modules stay in memory for the whole page lifetime.*

### Variants

* `doors.Script()`
  Creates a publicly accessible, cacheable static resource from inline script.
* `doors.ScriptPrivate()`
  Creates session-protected resource, internally caches script content.
* `doors.ScriptDisposable()`
  Creates a session-protected resource and does not cache content, preventing memory leaks with dynamic scripts.

## 2. Pass Data To JavaScript 

With **doors.AData** attribute

```templ
type AData struct {
  // name to access in js
	Name  string
	// any value that can be marshalled to json
	Value any
}
```

Example:
```templ
@doors.Script() {
	@doors.AData {
		Name: "user_profile",
		// user struct value
		Value: user,
  }
	<script>
	  // JSONParsed user object
		const user = $d.data("user_profile") 
	</script>
}
```

##  3. Call Go from JavaScript

### Custom Hook

To call Go from JavaScript with automatic serialization/deserialization of data

Attribute:

```templ
type AHook[I any, O any] struct {
  // returns output and a bool value indicating whether
  // you done (true, hook will be removed) or not (false, hook stays)
	On        func(ctx context.Context, r RHook[I]) (O, bool)
	Name      string
	Scope     []Scope // Scopes API
	Indicator []Indicator //Indicator API
}
```

#### Example

```templ
// attach hook to script (attachement allowed inside and outside doors.Script)
@doors.AHook[string, int]{
    Name: "length",
    On: func(ctx context.Context, r doors.RHook[string]) (int, bool) {
      return len(r.Data()), false
    },
}
@doors.Script() {	
  <script>
	  // JSONParsed user object
		const length = await $d.hook("length", "hello!") 
		/* 
		    // alternative, returns response instead of parced data
				const response = await $d.rawHook("length", "hello!") 
				const legth = await response.json()
		*/
		console.log(length) // 6
	</script>
}
```

> ⚠️ It's your responsibility to handle exceptions in custom hooks on the front-end. Any hook error will throw an instance of `$d.CaptureErr` class. Please refer to the [Error Handling](./ref/02-error-handling.md) article.

### Raw Custom Hook

Deal with http request yourself

```templ
type ARawHook struct {
	Name      string
	// RRawHook - request wrapper to parse/read body yourself
	On        func(ctx context.Context, r RRawHook) bool
	Scope     []Scope
	Indicator []Indicator
}

```

### Data Conversion

When you pass an argument to hook `$d.hook(name, arg) or $d.rawHook(name, arg)`,  arg is converted to the request body:

* `arg === undefined` - no request body
* `FormData` as is
* `URLSearchParams` as is with Content-Type "application/x-www-form-urlencoded;charset=UTF-8"
* `Blob` as is with Content-Type blob.type
* `File `and `ReadableStream` as octet-stream
* All the rest as JSON

> ℹ️ `$d.rawHook` on the frontend has no relation to  `doors.ARawHook` on the backend.

## 4. Call JavaScript from Go

You can register a JavaScript handler to be called at any time by the server. 

⚠️ Async functions are not supported, use hooks to report regarding long running operations

**Example:**

```templ
@doors.Script() {	
	<script>
    $d.on("my_call", (input) => {
      // logic
      return output
    })
  </script>
}
```

> Return value from the JavaScript handler will be serialized to JSON

## Invoke from Go (new API)

### Function

```templ
func Call[OUTPUT any](
  ctx context.Context,
  name string,
  arg any,
  onResult func(OUTPUT, error),
  onCancel func(),
) context.CancelFunc
```

- **ctx**: must be valid 
- **name**: JS handler name registered with `$d.on(name, fn)`.
- **arg**: any value; marshaled to JSON and passed to JS.
- **onResult**: called with the decoded JS return value (into `OUTPUT`) or an error. Pass `nil` if not needed.
- **onCancel**: called if provided context is canceled, the call is canceled, or cannot be scheduled/delivered due to instance shutdown.  Pass `nil` if not needed.
- **returns**: a `cancel` function (best-effort)

> Provide `json.RawMessage`  as `OUTPUT` if you don't need parsing

### Important: Call Registration/Invocation rules

You register a handler with  `$d.on(name, function)` to the **closest dynamic parent (Door)**. When you perform the Call from Go, it will look for a handler up the tree. 

To clarify how it works, let's examine this code:

```templ
@door_1 {
  @door_2 {
    @doors.Script() {
      /* handler 2 */
    }
    @Click{/* hook 2 */}
    <button>button</button>
  }
  @doors.Script() {
    /* handler 1 */
  }
  @Click{/* hook 1 */}
  <button>button</button>
}
@doors.Script() {
  /* handler 0 */
}
@Click{/* hook 0 */}
<button>button</button>
```

* **hook 0** is able to call only **handler 0**
* **hook 1** is able to call  **handler 1** and **handler 0**, with priority for **handler 1** (if the name is the same), but not **handler 2**
* **hook 2** is able to call **handler 2**, **handler 1** and **handler 0**

### Example

```templ
@doors.Script() {
  <script>
    $d.on("alert", (msg) => { alert(msg); return true })
  </script>
}

@doors.AClick{
  On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
    doors.Call[bool](ctx, "alert", "Hello from Go",
      func(ok bool, err error) { /* handle */ },
      nil,
    )
    return false
  }
}
<button>Alert Me!</button>

```

## 5. Clean Up

You can launch long-running background operations inside a script. `$d` allows the registration of a cleanup handler that will run when the script is removed from the page.

```templ
<script>
		const ctrl = new AbortController()
		// register clean-up function
		$d.clean(() => ctrl.abort())
		
	  /* start long-running process */
</script>
```

