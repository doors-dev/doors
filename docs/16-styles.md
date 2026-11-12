# Styles

This page covers class helper and stylesheet resources in **Doors**.

If generic resource syntax is new, start with [Resources](./14-resources.md). This page focuses on what is specific to `<style>` and `<link rel="stylesheet">`.

## Classes

`doors.Class(...)` builds reusable `class` values. It is useful when a project
has repeated utility classes, especially with Tailwind-style CSS.

```gox
var card = doors.Class(
	"rounded-xl border p-4",
	"bg-white shadow-sm",
)

elem ProductCard() {
	<article (card.Add("border-cyan-500").Filter("shadow-sm"))>
		<h2>Product</h2>
	</article>
}
```

`doors.Class` can be used as:

- an attribute value: `<div class=(style)>...</div>`
- an attribute modifier: `<div class="p-4" (doors.Class("rounded-lg"))>...</div>`
- an element proxy: `~>(doors.Class("h-5 w-5")) ~(Icon())`

Each argument is split like `strings.Fields`, so
`doors.Class("px-4", "py-2 rounded")` and `doors.Class("px-4 py-2 rounded")`
produce the same class list.

Use `Add` to append classes, `Remove` to remove classes that are already in the
current value, `Filter` to omit matching classes from final output, and `Join`
to compose reusable groups:

```gox
var button = doors.Class("inline-flex px-4 py-2 rounded-lg")
var primary = doors.Class("bg-cyan-600 text-white")

elem SaveButton(disabled bool) {
	<button (button.Join(primary))>Save</button>
}
```

`Filter` is helpful when a base style is shared but one branch must suppress a
specific utility class. This keeps the reusable style readable without forcing
every caller to duplicate it manually.

> Tailwind note: **Doors** does not run Tailwind generation for you. Keep your
> Tailwind CLI, PostCSS, or other CSS build configured separately, and make sure
> it scans your `.go` files or safelists classes that are assembled in Go.

## CSS

Most pages start with one of these:

- CSS written directly in the template: plain `<style>...</style>`
- CSS kept in a file, bytes, or string: `<link rel="stylesheet" (doors.Resource...)>`
- CSS already hosted somewhere: plain `href="..."`

Examples:

```gox
<style>
	h1 {
		color: red;
	}
</style>
```

```gox
<link
	rel="stylesheet"
	(doors.ResourceLocalFS("web/app.css"))>
```

```gox
<link rel="stylesheet" href="/assets/app.css">
```

## Style Tags

Use plain `<style>...</style>` when the CSS belongs to one page or one component:

```gox
<style>
	h1 {
		color: red;
	}
</style>
```

By default, **Doors** does not keep that literal `<style>` tag. It collects the CSS, creates a stylesheet resource, and emits a stylesheet link in the final HTML.

Use `raw` when you want a literal browser `<style>` tag:

```gox
<style raw>
	h1 {
		color: red;
	}
</style>
```

`name`, `private`, `nocache`, and `portable` also work here:

```gox
<style name="page.css" private>
	h1 {
		color: red;
	}
</style>
```

## Stylesheet Links

Use `<link rel="stylesheet">` when the CSS comes from a file, bytes, a string, a handler, or a proxy:

```gox
<link
	rel="stylesheet"
	(doors.ResourceLocalFS("web/app.css"))>
```

The modifier sets `href` on `<link>`. It takes any [resource](./14-resources.md#two-syntaxes) or `doors.ResourceExternal(...)` ([Not Served by Doors](./14-resources.md#not-served-by-doors)); `href` itself also takes a plain URL string or `[]byte` shorthand. Files, bytes, and strings are minified, cached, and hosted by default ([Hook and Hosted](./14-resources.md#hook-and-hosted)).

```gox
<link rel="stylesheet" href=(appCSS)>
```

`raw` serves the stylesheet unminified. Only on `<style>` does it keep the literal tag and body.

A script that adds the `<link>` itself gets the stylesheet URL from a `data:` attribute, `<script data:css=(doors.ResourceBytes(widgetCSS).Style())>`, read with `$data("css")`; Go code that needs the string uses `doors.ResourceBytes(widgetCSS).Style().URL(ctx)`, see [URL in Code](./14-resources.md#url-in-code).

## Attrs

These attrs control stylesheet behavior. The stage after `.Style()` does the same in a [resource chain](./14-resources.md#two-syntaxes):

- `raw`, `.Raw()`: serve unminified; on `<style>`, keep the literal tag
- `name`, `.Name(...)`: readable output file name

`private`, `nocache`, and `portable` (`.Private()`, `.NoCache()`, `.Portable()`) work as on every resource, see [Combinations](./14-resources.md#combinations). A chain with `ContentType` is an error on a stylesheet link.

Example:

```gox
<link
	rel="stylesheet"
	(doors.ResourceBytes(appCSS))
	name="app.css"
	private>
```
