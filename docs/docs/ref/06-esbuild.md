# esbuild

Framework includes [esbuild](https://esbuild.github.io/). It's a highly performant JavaScript build tool. 

## Concept

* `Imports` (checkout **ref/imports**) and inline scripts are processed by esbuild
* Resources are built and cached on the first reference.
* You  can configure multiple build profiles and select one for a specific resource via `Profile` field in `Imports` 
* The default profile is "" (empty string)
* `Imports`  set some esbuild options with required values, so, your values can be overwritten

## Basic Configuration

The **default configuration is sufficient for most cases**. But if you need to add `.jsx` building, **disable minification for easier debugging**; the configuration is simple with `doors.ESBuildOptions`:

```templ
type ESOptions struct {
  // list of external (non-local) dependecies, to avoid building issues
	External []string
	// enable mification
	Minify   bool
	// jsx setup
	JSX      JSX
}
```

You can use the JSX preset `doors.JSXPreact()` or `doors.JSXReact()` for Preact and React, or configure with  `doors.JSX `structure manualy:

```templ
type JSX struct {
	JSX          api.JSX // from github.com/evanw/esbuild/pkg/api
	Factory      string
	ImportSource string
	Fragment     string
	SideEffects  bool
	Dev          bool
}
```

> ⚠️ Basic (and default) configuration does not depend on profile value

## Advanced Configuration

If you need control over the whole esbuild option set, implement this interface:

```templ
type ESConf interface {
	Options(profile string) api.BuildOptions
}
```

Where api.BuildOptions coming from `github.com/evanw/esbuild/pkg/api`

**You must support the default profile (an empty string), which the framework uses to build the front-end part.**

## Apply configuration

To apply your configuration to the framework, use `doors.SetESConf(conf doors.ESConf)` mod:

```templ
// pass doors.SetESConf or custom doors.ESConf implementation
router.Use(doors.SetESConf(
		doors.ESOptions{
			JSX: doors.JSXPreact(),
		},
))
```

> 