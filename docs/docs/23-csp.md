# Content Security Policy

Implements Content Security Policy (CSP) generation and automatic hash collection for inline scripts and styles.  
Enable via:

```go
router.Use(doors.EnableCSP(doors.CSP{/* configuration */}))
```

---

## Configuration Fields

### 1. Always Emitted (you can append)

These directives are **always present** even if unset.  
You can append additional sources.

- **ScriptSources** (`[]string`)  
  `script-src` always includes `'self'` plus any collected hashes and sources.  
  If set, user sources are appended.

- **StyleSources** (`[]string`)  
  `style-src` always includes `'self'` plus any collected hashes and sources.  
  If set, user sources are appended.

- **ConnectSources** (`[]string`)  
  `connect-src` always includes `'self'`.  
  If set, user sources are appended.

---

### 2. Has Default Value (you can overwrite or omit)

These have built-in defaults when `nil`.  
An empty slice omits the directive.

| Field | nil → default | [] → omit | values → custom |
|--------|---------------|-----------|-----------------|
| **DefaultSources** | `'self'` | omitted | `<values>` |
| **FormActions** | `'none'` | omitted | `<values>` |
| **ObjectSources** | `'none'` | omitted | `<values>` |
| **FrameSources** | `'none'` | omitted | `<values>` |
| **FrameAcestors** | `'none'` | omitted | `<values>` |
| **BaseURIAllow** | `'none'` | omitted | `<values>` |

---

### 3. Omitted by Default (you can emit)

These are not emitted unless explicitly set.

| Field | nil/[] → omitted | values → emitted |
|--------|------------------|------------------|
| **ImgSources** | omitted | `img-src <values>` |
| **FontSources** | omitted | `font-src <values>` |
| **MediaSources** | omitted | `media-src <values>` |
| **Sandbox** | omitted | `sandbox <flags>` |
| **WorkerSources** | omitted | `worker-src <values>` |

---

### 4. Optional Flags

- **ScriptStrictDynamic** (`bool`)  
  When `true`, appends `'strict-dynamic'` to `script-src`.  
  When `false` or unset, `'strict-dynamic'` is **not** emitted.

- **ReportTo** (`string`)  
  When non-empty, emits `report-to <value>`.  
  Requires a matching `Report-To` HTTP header defining the reporting group.

---


## Example

```go
doors.CSP{
	ScriptStrictDynamic: true,
	ConnectSources: []string{"https://api.example.com"},
	ReportTo: "default",
}
```

Outputs header like:

```
default-src 'self'; connect-src 'self' https://api.example.com; script-src 'self' 'strict-dynamic'; style-src 'self'; form-action 'none'; object-src 'none'; frame-src 'none'; frame-ancestors 'none'; base-uri 'none'; report-to default
```