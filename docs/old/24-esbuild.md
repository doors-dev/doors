# ESBuild

Integrates [esbuild](https://esbuild.github.io/), a high-performance JavaScript/TypeScript build tool.

---

## Concept

- [Imports](./20-imports.md) and inline scripts are processed through **esbuild**.  
- Resources are **built and cached** on first reference.  
- Multiple **build profiles** can be configured and selected per resource via the `Profile` field in `Imports`.  
- The **default profile** is an empty string `""`.  
- `Imports` apply required esbuild options automatically — custom values may be overridden.

---

## Basic Configuration

The **default configuration** covers most common use cases.  
To add JSX building or disable minification for easier debugging, use `doors.ESBuildOptions`:

```go
type ESOptions struct {
	// List of external (non-local) dependencies to skip during build
	External []string
	// Enable minification
	Minify   bool
	// JSX setup
	JSX      JSX
}
```

You can use presets:

```go
doors.JSXPreact() // for Preact
doors.JSXReact()  // for React
```

Or configure manually with `doors.JSX`:

```go
type JSX struct {
	JSX          api.JSX // from github.com/evanw/esbuild/pkg/api
	Factory      string
	ImportSource string
	Fragment     string
	SideEffects  bool
	Dev          bool
}
```

> ⚠️ The **basic configuration** provides the same profile regardless of the provided profile value.

---

## Advanced Configuration

To fully control esbuild behavior, implement the `doors.ESConf` interface:

```go
type ESConf interface {
	Options(profile string) api.BuildOptions
}
```

`api.BuildOptions` comes from `github.com/evanw/esbuild/pkg/api`.

You **must support the default profile** (`""`), which the framework uses for the main bundle.

---

## Apply Configuration

Apply your configuration using the middleware function `doors.UseESConf`.  
Different profiles can be used for **development** vs **production** builds.

```go
// UseESConf configures esbuild profiles for JavaScript/TypeScript processing.
// Different profiles can be used for development vs production builds.
func UseESConf(conf ESConf) Use {
	return router.UseESConf(conf)
}
```

Example:

```go
router.Use(doors.UseESConf(
	doors.ESOptions{
		JSX: doors.JSXPreact(),
	},
))
```
