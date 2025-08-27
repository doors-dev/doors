# CSP

Use `doors.EnableCSP(doors.CSP{/* configuration fields */})`  router mod to enable CSP.  

##  Configuration Fields 

### 1. Always Emitted (you can append)

These directives are **always present**. Even if you don’t configure them, they emit with at least `'self'` or collected sources/hashes.  
You can append additional sources.

- **ScriptSources** ([]string)  
  `script-src` is always emitted with `'self'` plus any collected hashes/sources.  
  If set, user sources are appended.

- **ScriptStrictDynamic** (bool)  
  When true, appends `'strict-dynamic'` to `script-src`.

- **StyleSources** ([]string)  
  `style-src` is always emitted with `'self'` plus any collected hashes/sources.  
  If set, user sources are appended.

- **ConnectSources** ([]string)  
  `connect-src` is always emitted with `'self'`.  
  If set, user sources are appended.

###  2. Has Default Value (you can overwrite or omit)

These directives fall back to a default when `nil`.  
If set to a zero-length array, the directive is **omitted**.  
If set to values, your values are used.

- **DefaultSources** ([]string)  
  - `nil` → `default-src 'self'`  
  - `[]` → omitted  
  - values → `default-src <values>`

- **FormActions** ([]string)  
  - `nil` → `form-action 'none'`  
  - `[]` → omitted  
  - values → `form-action <values>`

- **ObjectSources** ([]string)  
  - `nil` → `object-src 'none'`  
  - `[]` → omitted  
  - values → `object-src <values>`

- **FrameSources** ([]string)  
  - `nil` → `frame-src 'none'`  
  - `[]` → omitted  
  - values → `frame-src <values>`

- **FrameAcestors** ([]string)  
  - `nil` → `frame-ancestors 'none'`  
  - `[]` → omitted  
  - values → `frame-ancestors <values>`

- **BaseURIAllow** ([]string)  
  - `nil` → `base-uri 'none'`  
  - `[]` → omitted  
  - values → `base-uri <values>`

### 3. Omitted by Default (you can emit)

These directives are not emitted unless you explicitly configure them.  
If set to values, they are emitted.

- **ImgSources** ([]string)  
  - `nil` or `[]` → omitted  
  - values → `img-src <values>`

- **FontSources** ([]string)  
  - `nil` or `[]` → omitted  
  - values → `font-src <values>`

- **MediaSources** ([]string)  
  - `nil` or `[]` → omitted  
  - values → `media-src <values>`

- **Sandbox** ([]string)  
  - `nil` or `[]` → omitted  
  - values → `sandbox <flags>`

- **WorkerSources** ([]string)  
  - `nil` or `[]` → omitted  
  - values → `worker-src <values>`
  
- **ReportTo** (string)  
  - `""` → omitted  
  - non-empty → `report-to <value>`  
  
    > ⚠️ Requires sending a matching **`Report-To` HTTP response header**  to define the reporting group.
