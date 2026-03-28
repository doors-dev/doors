# Template Syntax

**GoX** is a **Go** syntax extension designed primarily for **Doors**, but it can also be used standalone.

It is compatible with `templ` and can write the resulting HTML directly to an `io.Writer`.

**GoX** support is provided by its own language server, which enables a nearly seamless experience across `.gox` and `.go` files. It also compiles `.x.go` files to `.go` automatically as you type, and saves the generated output when you save the `.gox` file.

> It is worth keeping `gox` on your `PATH`, so you, your editor tooling, and code agents can run commands like `gox fmt` and `gox gen` when needed.

## HTML as Expression

In `.gox` files, HTML compiles to a value of type `gox.Elem`.

```gox
var hello gox.Elem = <h1>Hello World!</h1>
```

It behaves like any other Go expression. For example, you can assign it to a field:

```gox
Component{
    Content: <>
        <h1>Header</h1>
        <p>Text</p>
    </>,
}
```

> `<>...</>` acts as a container and does not output any tags.

You can also pass it as an argument or return it from a function:

```gox
func header() gox.Elem {
    return <h1>Header</h1>
}
```

## Elem Primitive

**GoX** adds the `elem` keyword, which lets you write HTML directly in the function body:

```gox
elem header() {
    <h1>Header</h1>
}

/* is 100% equivalent to:
func header() gox.Elem {
    return <h1>Header</h1>
}
*/
```

Anonymous functions are supported, as are methods:

```gox
var f = elem() {
    <h1>Header</h1>
}

type App struct{}

elem (a App) Main() {
    <!doctype html>
    <html lang="en">
        ~/* ... */
    </html>
}
```

## Component

A component is anything that implements `gox.Comp`:

```go
type Comp interface {
    Main() gox.Elem
}
```

Example:

```gox
type Item struct {
    Name string
    Desc any
}

elem (it Item) Main() {
    <card>
        <header>~(it.Name)</header>
        <div>
            ~(it.Desc)
        </div>
    </card>
}
```

> `gox.Elem` also implements `gox.Comp`. So if a function accepts `gox.Comp`, it also accepts `gox.Elem`.

## Placeholder

Render a value of any type using a placeholder: `~(expression)`.

```gox
elem Card(name string, price int) {
    <card>
        <header>~(name)</header>
        <span>~(price)</span>
    </card>
}
```

> `~(...)` accepts any expression of any type.

Sometimes it is convenient to insert multiple values at once. Pass them as multiple arguments:

```gox
elem Hello(name string, surname string) {
    Hello ~(name, " ", surname)!
}
```

You can omit the parentheses for strings, numbers, and composite literals (`struct`, `array`, `slice`, `map`):

```gox
elem Items() {
    <div>
        ~item{
            Name: "Product A",
            Desc: <>
                <b>Bold</b> product!
            </>
        }
        ~item{
            Name: "Product B",
            Desc: <>
                <i>Italic</i> product!
            </>
        }
    </div>
}
```

Values are rendered using the default formatter, with special handling for:

- `gox.Comp` and `gox.Elem`
- [templ.Component](https://github.com/a-h/templ)
- `gox.Job` and `gox.Editor` — advanced GoX primitives
- `[]string`, `[]gox.Comp`, `[]gox.Elem`, `[]any`, `[]gox.Job` — rendered item by item

## Conditions and Loops

`if / else-if / else` are available in the form `~(if ... { })`:

```gox
<div>
    ~(if user != nil {
        Hello ~(user.name)!
    } else if loggedOut {
        Bye!
    } else {
        Please log in
    })
</div>
```

A `for` loop is written in a similar way:

```gox
<table>
    ~(for _, user := range users {
        <tr>
            <td>~(user.name)</td>
            <td>~(user.email)</td>
        </tr>
    })
</table>
```

## Attributes

Use parentheses to provide a Go expression as an attribute value:

```gox
elem block(id: string) {
    <div id=(id)>Content</div>
}
```

- Any value is accepted.
- If the value is `false` or `nil`, the attribute is omitted.
- Attribute names are case-sensitive (`class` and `Class` are different attributes). Use consistent casing.

### Modifiers

To inspect and/or modify multiple attributes at once, implement `gox.Modify`:

```go
type Modify interface {
    Modify(ctx context.Context, tag string, attrs gox.Attrs) error
}
```

Attribute modifiers are applied at render time and can mutate the full attribute set. To attach one, place it in parentheses inside the opening tag:

```gox
<button (LandingAction)>
    Request Demo!
</button>
```

### Advanced attribute behavior

During rendering, the default formatter is used unless the value implements `gox.Output`:

```go
type Output interface {
    Output(w io.Writer) error
}
```

To compute a new attribute value from the previous one, or to take the attribute name into account during assignment, implement `gox.Mutate`:

```go
type Mutate interface {
    Mutate(attributeName string, attributeValue any) (new any)
}
```

## Code

### Inline expression

An inline expression is a function literal evaluated immediately during rendering. Its return value is inserted exactly where it appears.

```gox
<div>
    ~func {
        user, err := Users.get(id)
        if err != nil {
            return <span>DB error</span>
        }
        return Card(user)
    }
</div>
```

It can also be used in attribute values:

```gox
<input type="checkbox" checked=func {
    user := Users.get(id)
    return user.Agreed // false or nil omits the attribute
}>
```

### Go snippets

To switch to **Go** mode, use `~{ /* statements */ }`:

```gox
<card>
    ~{
        // write regular Go code here
        user := Users.Get(id)
    }
    <header>~(user.name)</header>
</card>
```

### Comments

To write comments inside templates, use `~//` or `~/* ... */`:

```gox
<div>
    ~// <div>Commented Content</div>

    ~/*
        multiline
        comment
    */
</div>
```

> HTML comments are also supported.

## Element Proxy

A proxy captures the following element subtree and can transform it before rendering.

Apply a proxy by prefixing the target expression with `~>`:

```gox
<>
    ~>(proxy) <div>
        Proxy can apply transformations to this HTML.
    </div>
</>
```

You can apply multiple proxies as a comma-separated list:

```gox
<>
    ~>(Track, Transform) <div>
        Proxy can apply transformations to this HTML.
    </div>
</>
```

A proxy must implement:

```gox
type Proxy interface {
    Proxy(cur gox.Cursor, elem gox.Elem) error
}
```

> A proxy is a powerful tool that lets you extend tooling with custom capabilities.

## Raw

To output HTML verbatim, without escaping or template processing, wrap it in the special tag: `<:>...</:>`.

```gox
<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
    <:>
        <path d="..." />
        <path d="..." />
    </:>
</svg>
```

This is recommended for large static fragments, especially inline SVG, to reduce rendering overhead.

Alternatively, you can implement the `gox.Editor` interface over the underlying `gox.Cursor`, which controls the printing process:

```go
type Editor interface {
    Edit(cur gox.Cursor) error
}
```

Or use the `gox.EditorFunc` helper:

```gox
<div>
    ~(gox.EditorFunc(func(cur gox.Cursor) error {
        return cur.Raw("<span>Unescaped HTML</span>")
    }))
</div>
```

> For a deeper dive into **GoX** and its capabilities, refer to [https://github.com/doors-dev/gox](https://github.com/doors-dev/gox).
