# doors

Back-end UI Framework for feature-rich, secure, and fast web apps in Go.

⚠️ **Beta - Not Ready for Production**

## Getting Started

* See the [Tutorial](./docs/tutorial) for building your first doors application.
* Read the [Docs](./docs/docs) to dive into details.
* Check out the [API Reference](https://docs.doors.dev).

## Philosophy 

### Explicid
Build direct connections between events, state, and HTML in a completely type-safe environment. *It hits different*.

### Lightweight 
Fast loading, non-blocking execution environment, minimal memory footprint

### Server Centric
Business logic runs on the server, browser acts like a human I/0. 

### Straight
Native experience of classic MPA with natural reactive UI capabilities.

### JS friendly
If you need - integration, bundling, and serving tools included.

## How It Works

### Stateful Server + Ultra-Thin Client

> API-free architecture

1. **User loads page** → Server creates instance and sets session cookie
2. **Server maintains state** → Live representation of each user's page
3. **Persistent connection** → Lightweight client syncs with server
4. **Events flow up** → User interactions sent to Go handlers
5. **Updates flow down** → Server sends specific DOM changes


### Core Components

**Doors** - Dynamic placeholders in HTML where content can change:
- Update, Replace, Remove, Clear, Reload operations
- Explicit control over what changes and when

**Beams** - Reactive state primitives on the server:
- SourceBeam for mutable state
- Derived beams for computed values
- Automatic propagation of changes

**Path Models** - Type-safe routing through Go structs:
- URL patterns as struct tags
- Variants within single model for related routes
- Compile-time parameter validation

### Instance and Session Model

- Each browser tab creates an **instance** (live server connection)
- Multiple instances share a **session** (common state)
- Navigation within same Path Model: reactive updates
- Navigation to different Path Model: new instance created

### Real-Time Sync Protocol

- Client maintains connection for syncronization via short-living, handover HTTP requests
- Works through proxies and firewalls
- Takes advatage of QUIC

### Event Handling
- Secure session-scoped DOM event handeling in Go 
- Events as separate HTTP requests
- Advanced concurrency control (blocking, debounce and more)


## When to Use doors

**Excellent for:**
- SaaS products
- Business process automation (ERP, CRM, etc)
- Administrative interfaces
- Customer portals
- Internal tools 
- Other form-heavy applications

**Not ideal for:**
- Public marketing websites with minmal interactivity
- Offline-first applications
- Static content sites

## Comparison

Unlike **React/Vue**: No business logic on the client side, no hydration, NPM-free.

Unlike **htmx**: Full type safety, reactive state, and programmatic control from Go.

Unlike **Phoenix LiveView**: Go instead of Elixir, explicit update model, superior concurrency model and QUIC friendly

## License

doors is source-available under the **Business Source License 1.1 (BUSL-1.1)** from **doors dev LLC**.

- **Free for development** (non-production)  
- **Free for non-commercial production** (personal, education, research, non-profit) — optional pay-what-you-want support  
- **Commercial production** requires a paid license (Startup, Business, or Enterprise)  

Each version of doors automatically converts to **AGPL-3.0** after 4 years.

### Commercial Licensing

Commercial licenses are delivered as signed **License Certificates**:

- **Startup License** (per production domain) — strict startup criteria, internal use only  
- **Business License** (per production domain / per client domain) — internal + client work  
- **Enterprise License** — custom scope (domain-linked, company-linked, or enterprise-wide)  

Unless otherwise stated, License Certificates are **permanent (no expiration)**.  
Certificates may be shared, but they only enable Production use for the scope encoded in the certificate.  

To activate a commercial license, the License Certificate must be provided to the framework via the API described in the official documentation.

To purchase a license, visit [https://doors.dev](https://doors.dev).  
For Enterprise terms, you may also contact [sales@doors.dev](mailto:sales@doors.dev).

### Full License Texts

- [LICENSE.txt](./LICENSE.txt) — BUSL with parameters  
- [LICENSE-COMMERCIAL.txt](./LICENSE-COMMERCIAL.txt) — Commercial license summary  
- [COMMERCIAL-EULA.md](./COMMERCIAL-EULA.md) — Full commercial terms  

---

**Note:** This README section is a summary. The binding terms are in the license files listed above.

