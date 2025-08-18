# Imports

The framework provides the `doors.Imports` component for including JavaScript and CSS resources in a page.  
It supports multiple import types—ranging from locally built ES modules to externally hosted scripts and stylesheets—and automatically generates an **import map** along with the necessary `<script>` and `<link>` tags.

When a page is rendered, the `doors.Imports` component processes all declared imports, updates the Content Security Policy (CSP) with required hashes and sources, and ensures that resources are properly loaded or mapped for later usage.

> Refer to **ref/esbuild** article for esbult configuration 

---

## Overview

`doors.Imports` accepts one or more import entries implementing the `Import` interface.  
Each entry describes how a resource should be built or referenced and whether it should be loaded immediately.

Resources can be:

- ES modules, created from source files, raw files, byte slices, or bundles.
- Stylesheets, created from files, byte slices, or hosted/external URLs.

---

## Page-level integration

`doors.Imports` should be placed once per page, typically inside the `<head>` element.  
It collects all provided imports, generates the import map, writes it to HTML, and renders any required `<script>` or `<link>` tags.

If `doors.Imports` is called more than once per page render, the extra calls are ignored and an error is logged.

```templ
// Example: in your page head
@doors.Imports(
    // one or more Import entries …
)
```

When invoked, `doors.Imports`:

1. Writes a `<script type="importmap">…</script>` block.
2. Prepares, bundles, and caches
3. Records the script’s hash for CSP.
4. Renders any queued loader scripts or stylesheets.

## Common rules

- **Specifier vs. Load:** If an entry is created with no `Specifier` (required in the import map) and `Load === false (true means include as HTML element), it is skipped and a warning is logged. At least one must be set to include the resource.
- **CSP updates:** Import map hashes are added to CSP (if it's enabled). External JS/CSS sources are whitelisted using CSP `script-src` / `style-src`.
- **Naming and paths:** Generated asset file names are derived from the source path or  `Name` if provided. This can help you identify the resource in the web debugger.
- For bundling, `Profile` value will be used to determine esbuild profile. The default profile is an empty string.

## Import types

### `ImportModule`

Builds (not bundeles) a JS/TS file into an ES module via the framework’s builder.

**Fields:** `Specifier`, `Path`, `Profile`, `Load`, `Name`.

- Adds to the import map when `Specifier` is set.
- Loads immediately with a `<script>` tag when `Load` is true.

```templ
@doors.Imports(doors.ImportModule{
    Specifier: "module",
    Path:      "module/index.ts",
    Load:      false,
})
```

> Then it can be imported into the script via the specifier
>
> ```templ
> // script helper component
> @doors.Script() {
> 	<script>
> 		const module = await import("module")
> 	</script>
> }
> ```
>
> 

### `ImportModuleBytes`

Same as `ImportModule` but with module source provided as `[]byte`.

**Fields:** `Specifier`, `Content`, `Profile`, `Load`, `Name`.

------

### `ImportModuleRaw`

Serves a JavaScript file as-is without processing.

**Fields:** `Specifier`, `Path`, `Load`, `Name`.

------

### `ImportModuleRawBytes`

Serves raw JavaScript from `[]byte` without processing.

**Fields:** `Specifier`, `Content`, `Load`, `Name`.

------

### `ImportModuleBundle`

Creates a bundled ES module from an entry point using the builder.

**Fields:** `Specifier`, `Entry`, `Profile`, `Load`, `Name`.

------

### `ImportModuleBundleFS`

Bundles from an `fs.FS` (such as `embed.FS`). CacheKey is used to identify a cache entry. 

**Fields:** `CacheKey`, `Specifier`, `FS`, `Entry`, `Profile`, `Load`, `Name`.

------

### `ImportModuleHosted`

References a locally hosted JavaScript module by absolute path from the application root.

**Fields:** `Specifier`, `Load`, `Src`.

------

### `ImportModuleExternal`

References an external JavaScript module by URL.

**Fields:** `Specifier`, `Load`, `Src`.

Also adds the URL to CSP `script-src`.

------

### `ImportStyle`

Processes a CSS file (minifying if applicable) and emits a `<link>` tag.

**Fields:** `Path`, `Name`.

------

### `ImportStyleBytes`

Processes CSS content from `[]byte` and emits a `<link>` tag.

**Fields:** `Content`, `Name`.

------

### `ImportStyleHosted`

References a locally hosted CSS file without processing.

**Fields:** `Href`.

------

### `ImportStyleExternal`

References an external CSS file and adds its URL to CSP `style-src`.

**Fields:** `Href`.

## Usage example

 Bundle and import module and stylesheet from local sources

```templ
@doors.Imports(
    doors.ImportModuleBundle{
        Specifier: "module",
        Entry:      modulePath + "/index.ts",
    },
    doors.ImportStyle{
        Path: modulePath + "/style.css",
    },
)

```

