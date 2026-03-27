# Head & Status

In **Doors**, `<title>`, `<meta>`, and `doors.Status(...)` are page-level tools.

They are special because they do not behave like ordinary local DOM content:

- `<title>` and `<meta>` can be rendered anywhere in your page tree
- **Doors** always moves them into the real document `<head>`
- later updates are synchronized on the client
- `doors.Status(...)` can also be rendered anywhere, but it affects only the initial HTTP page response

It is a little unusual structurally, but convenient when the page content depends on state or path and the matching title/meta should be updated in the same rendering branch.

## Head

You can render `<title>` and `<meta>` tags anywhere in the HTML tree, including inside content that is otherwise in the body.

For example:

```gox
elem Page() {
	<>
		<title>Docs</title>
		<meta name="description" content="doors dev desc">
	</>

	<main>
		<h1>Docs</h1>
	</main>
}
```

Even though those tags are written inside the body here, **Doors** outputs them in the real `<head>`.

If the page does not already contain a `<head>`, **Doors** creates one.

Keep `<title>` simple: it should contain text, not nested tags.

When `<title>` and `<meta>` are rendered outside a literal `<head>`, it is often clearer to wrap them in their own `<>...</>` block so readers can immediately see that this is intentional.

## Meta

For `<meta>`, **Doors** uses either:

- `name`
- `property`

to identify which head tag should be updated.

Everything else on the tag is treated as normal attributes to sync, usually `content`.

`<meta>` should be written as a normal void tag.

```gox
<>
	<meta name="description" content="doors dev desc">
	<meta property="og:title" content="Docs">
</>
```

The practical rule is to render one canonical tag for each `name` or `property` you care about.

## Sync

These tags are synchronized with the frontend.

That means when the current live page rerenders and produces a different `<title>` or `<meta>`, **Doors** updates `document.title` and the matching head tags in the browser.

This is why same-instance page switches can update the browser title without a full page reload.

## Status

Use `doors.Status(code)` when the initial page response should use a specific HTTP status.

```gox
elem NotFound() {
	<html>
		<body>
			~(doors.Status(404))
			<h1>Not found</h1>
		</body>
	</html>
}
```

Like `<title>` and `<meta>`, it can be rendered anywhere in the page tree.

But unlike them, it only affects the initial page response. Once the HTML response has been sent, later reactive updates cannot change that HTTP status code.

So `doors.Status(404)` is useful for the first render of a not-found page, but not for trying to change the browser response code later during the life of an already-open instance.

## Rules

- Render plain `<title>` and `<meta>` tags directly. **Doors** will place them in `<head>`.
- Use `name` or `property` on `<meta>` so **Doors** knows which tag to sync.
- Expect title and meta to update on the client when the live page rerenders.
- Use `doors.Status(...)` only for the initial page response.
