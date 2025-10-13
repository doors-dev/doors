# Imports

Declarative components for importing and registering JavaScript and CSS assets.  
Each import automatically integrates with [Content Security Policy (CSP)](./23-csp.md) and the **import map**.

---

## JavaScript Imports

### `ImportCommon`

Imports a plain JavaScript file and inserts it as a regular `<script>` tag.

```templ
@doors.ImportCommon{
  Path: "scripts/app.js",
}
```

- **Path** — required; local file path.
- **Name** — optional; overrides output file name.
- **Attrs** — optional; additional attributes for the `<script>` tag.

---

### `ImportCommonExternal`

Imports an external JS file and registers it with CSP.

```templ
@doors.ImportCommonExternal{
  Src: "https://cdn.example.com/lib.js",
}
```

---

### `ImportCommonHosted`

Links an already hosted JS file from your static directory.

```templ
@doors.ImportCommonHosted{
  Src: "/static/app.js",
}
```

---

### `ImportModule`

Imports a JS or TS module, builds it with **esbuild**, and exposes it as an ES module.

```templ
@doors.ImportModule{
  Path:      "web/main.ts",
  Specifier: "main",
  Load:      true,
}
```

- **Specifier** — optional; key used in the import map.  
- **Path** — required; module file path.  
- **Profile** — optional; esbuild profile.  
- **Load** — if true, loads immediately in the page.  
- **Name** — custom file name.  
- **Attrs** — optional; additional `<script>` attributes.

---

### `ImportModuleRaw`

Imports a raw JS file without processing or bundling.

```templ
@doors.ImportModuleRaw{
  Path:      "scripts/plain.js",
  Specifier: "plain",
  Load:      true,
}
```

---

### `ImportModuleBundle`

Bundles an entry JS/TS file and its dependencies using **esbuild** into a single module.

```templ
@doors.ImportModuleBundle{
  Entry:     "web/app/index.ts",
  Specifier: "app",
  Load:      true,
}
```

---

### `ImportModuleBundleFS`

Bundles from an in-memory filesystem (`fs.FS`).

```templ
@doors.ImportModuleBundleFS{
  CacheKey:  "app",
  FS:        assets,
  Entry:     "index.ts",
  Specifier: "app",
  Load:      true,
}
```

---

### `ImportModuleHosted`

Registers a locally hosted JS module without processing.

```templ
@doors.ImportModuleHosted{
  Src: "/static/app.js",
  Load: true,
}
```

---

### `ImportModuleExternal`

Registers an external JS module and adds it to CSP.

```templ
@doors.ImportModuleExternal{
  Src: "https://cdn.example.com/mod.js",
  Load: true,
}
```

---

## CSS Imports

### `ImportStyle`

Processes and links a CSS file (minification and fingerprinting included).

```templ
@doors.ImportStyle{
  Path: "styles/main.css",
}
```

---

### `ImportStyleHosted`

Links a locally hosted CSS file without processing.

```templ
@doors.ImportStyleHosted{
  Href: "/static/style.css",
}
```

---

### `ImportStyleExternal`

Links an external CSS file and registers it with CSP.

```templ
@doors.ImportStyleExternal{
  Href: "https://cdn.example.com/style.css",
}
```

---

## Notes

- If both `Specifier` and `Load` are empty or false, the import is ignored.  
- All imports are automatically included in the application’s **CSP** and **import map**.  
- Prefer `ImportModule` for ES modules.  
