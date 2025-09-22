# doors

Back-end UI Framework for feature-rich, secure, and fast web apps in Go.

⚠️ **Beta - Not Ready for Production**

## Getting Started

* See the [Tutorial](https://doors.dev/tutorial/) for building your first doors application.
* Read the [Docs](https://doors.dev/docs/) to dive into details.
* Check out the [API Reference](https://docs.doors.dev).

## Philosophy 

### Explicid
Build direct connections between events, state, and HTML in a completely type-safe environment. *It hits different*.

### Lightweight 
Fast loading, non-blocking execution environment, minimal memory footprint

### Server Centric
Business logic runs on the server, and the browser acts like a human I/O. 

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

**Door** - Dynamic container in HTML where content can change:

- Update, Replace, Remove, Clear, Reload operations
- Form a tree structure, where each branch has its own lifecycle
- Provides local [context](https://pkg.go.dev/context) that can be used as an unmount hook.

**Beam** - Reactive state primitive on the server:

- SourceBeam for mutable state
- Derived beams for computed values
- Respects the dynamic container tree, guaranteeing render consistency

**Path Models** - Type-safe routing through Go structs:

- Declare multiple path variants (the matched field becomes true)
* Use type-safe parameter capturing 
* Use splat parameter to capture the remaining path tail
* Use almost any types for query parameters ([go-playground/form](https://github.com/go-playground/form) under the hood)

### Instance and Session Model

- Each browser tab creates an **instance** (live server connection)
- Multiple instances share a **session** (common state)
- Navigation within same Path Model: reactive updates
- Navigation to different Path Model: new instance created

### Real-Time Sync Protocol

- Client maintains a connection for synchronization via short-lived, handover HTTP requests
- Works through proxies and firewalls
- Takes advantage of QUIC

### Event Handling
- Secure session-scoped DOM event handeling in Go 
- Events as separate HTTP requests
- Advanced concurrency control (blocking, debounce, and more)


## When to use *doors*

**Excellent for:**
- SaaS products
- Business process automation (ERP, CRM, etc)
- Administrative interfaces
- Customer portals
- Internal tools 
- Other form-heavy applications

**Not ideal for:**
- Public marketing websites with minimal interactivity
- Offline-first applications
- Static content sites

## Comparison

Unlike **React/Vue**: No business logic on the client side, no hydration, NPM-free.

Unlike **htmx**: Full type safety, reactive state, and programmatic control from Go.

Unlike **Phoenix LiveView**: Explicit update model, parallel rendering & non-blocking event handling and QUIC friendly

## License

*doors* is source-available under the **Business Source License 1.1 (BUSL-1.1)** from **doors dev LLC**.

- **Free for development** (non-production)  
- **Free for non-commercial production** (personal, education, research, non-profit) — optional pay-what-you-want support  
- **Commercial production** requires a paid license (lifetime, no subscription)

Each version of doors automatically converts to **AGPL-3.0** after 4 years.

To purchase a license, visit [https://doors.dev](https://doors.dev).  
For Enterprise terms, you may also contact [sales@doors.dev](mailto:sales@doors.dev).

### Full License Texts

- [LICENSE.txt](./LICENSE.txt) — BUSL with parameters  
- [LICENSE-COMMERCIAL.txt](./LICENSE-COMMERCIAL.txt) — Commercial license summary  
- [COMMERCIAL-EULA.md](./COMMERCIAL-EULA.md) — Full commercial terms  



