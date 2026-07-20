# Printer Middleware

`doors.WithPrinterMiddleware` lets you observe or adjust the HTML **Doors** emits, at the point where render jobs are serialized.

Use it for cross-cutting output concerns: stamping elements with extra attributes, collecting render metrics, or auditing produced markup.

## Option

```go
app := doors.NewApp(page,
	doors.WithPrinterMiddleware(func(next gox.Printer) gox.Printer {
		return &stamp{next: next}
	}),
)
```

The middleware is a `func(next gox.Printer) gox.Printer`. It applies to every page render and Door render cycle in the app. Passing `nil` disables wrapping.

## Contract

**Doors** calls the middleware once per *drain*:

- the initial page render — the whole document, including fully static pages
- a Door render cycle — the updated Door plus all descendant Doors rendered in that cycle, delivered as one payload

The returned printer must be non-nil. It receives that drain's jobs as sequential `Send` calls on a single goroutine, in document order, and is never reused across drains. The middleware function itself can be called concurrently — separate pages and Door updates can print at the same time — so keep per-drain state in the returned printer, not in shared variables.

Rules:

- forward each job to `next` exactly once; jobs are pooled, so never retain a job or its `Attrs` after `Send` returns
- read `job.Context()` for scope information: each job carries the context of the render scope that produced it, so helpers like `doors.SessionContext` work on it
- `gox.JobHeadOpen` attrs can be inspected or changed until the job is serialized; they may be `nil` on container heads that emit no HTML
- do not modify **Doors**-managed output: `d0*` elements and attributes, and rewritten resource URLs
- do not inject head-relevant elements into a page render stream before the document head is printed
- do not block: the middleware runs on the render hot path

## Example

Count elements and stamp each with its render index:

```go
type stamp struct {
	next  gox.Printer
	count int
}

func (s *stamp) Send(j gox.Job) error {
	if open, ok := j.(*gox.JobHeadOpen); ok && open.Tag != "d0-r" && open.Attrs != nil {
		s.count++
		open.Attrs.Get("data-render-index").Set(strconv.Itoa(s.count))
	}
	return s.next.Send(j)
}

app := doors.NewApp(page, doors.WithPrinterMiddleware(
	func(next gox.Printer) gox.Printer {
		return &stamp{next: next}
	},
))
```

## Not Wrapped

The middleware only sees **Doors** render output. It does not wrap:

- error pages rendered through `doors.WithErrorPage`
- the error replacement HTML of a failed Door render cycle
- static files and resources served through `app.Use(...)` middleware or resource URLs

## Related

- [App](./04-app.md) for `NewApp` and app middleware.
- [Configuration](./21-configuration.md) for the full option list.
