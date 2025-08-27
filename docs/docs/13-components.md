# Helper Components

## Fragment Render

`func doors.F(doors.Fragment) templ.Component`

Renders fragment

```templ
type Fragment interface {
    Render() templ.Component
}
```

```templ
templ Demo() {
	@doors.F(&Counter{})
}
```

## Script

`func Script() templ.Component`
`func ScriptPrivate() templ.Component`
`func ScriptDisposable() templ.Component`

Prepares an inline script. Please look in the [JavaScript](./10-javascript.md) article for details.

## Style

`func Style() templ.Component`
Creates a publicly accessible, cacheable, static resource from an inline style sheet.

`func StylePrivate() templ.Component`
Creates a session-protected resource and internally caches style content.

`func StyleDisposable() templ.Component`
Creates a session-protected resource and does not cache content to prevent memory leaks with dynamic styles.

```templ
@doors.Style(){
	<style>
		body {
      background-color: powderblue;
    }
    h1 {
      color: blue;
    }
    p {
      color: red;
    }
	</style>
}
```

## Text

`func Text(any) templ.Component`

Converts any value to a component with an escaped string using default formats.

```templ
@doors.Text(3) 
```

## Evaluate

`func E(func(context.Context) templ.Component) templ.Component`

Evaluates a function at render time and returns its component.

```templ
@doors.E(func(ctx context.Context) templ.Component {
		if str = "" {
				return doors.Text("No value")
		}
		return doors.Text(str)
})
```

## Run

`func Run(func(context.Context)) templ.Component`

Runs the function during rendering (synchronously).

```templ
templ (f *Fragment) Render() {
    @doors.Run(func(ctx context.Context) {
    		f.init()
    })
    <h1>{ f.header }</h1>
}

```

## Components

Renders multiple components sequentially.

`func Components(content ...templ.Component) templ.Component`

## Attributes

`func Attributes([]Attr) templ.Component`

Prepares magic attributes from an array, refer to [Attributes](./07-attributes.md)

## Any

`func Any(any) templ.Component {`

Depending on the value type

* renders `templ.Component`  directly
* renders `doors.Fragment` with `doors.F`
* renders `[]templ.Component` with `doors.Components`
* prepares magic attributes from `[]Attr` with `doors.Attributes`
* rvaluates and renders `func(context.Context) templ.Component `with `doors.E`
* runs `func(context.Context) `with `doors.Run`
* tries to format and render as text with `doors.Text`

## Goroutine Spawn

`func Go(func(context.Context)) templ.Component`

* Starts a goroutine at render time, tied to the component lifecycle.
* Context is canceled when the component unmounts.
* Supports blocking operations safely.

```templ
@doors.Go(func(ctx context.Context) {
    for {
        select {
        case <-time.After(time.Second):
            door.Update(ctx, currentTime())
        case <-ctx.Done():
            return
        }
    }
})
```

> ℹ️ Alternatively, you can spawn with `go func()`  and obtain a blocking allowed context with `doors.AllowBlocking(ctx)` 

