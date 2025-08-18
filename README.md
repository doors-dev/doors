# doors

Backend UI framework for modern, feature-rich, and secure web apps in Go. 

⚠️ **Early Preview - Not Ready for Production**

## Getting Started

See the [Tutorial](./docs/tutorial) for building your first doors application.
Checkout the [Docs](https://docs.doors.dev) for API reference.

## What is doors?

doors combines the best of Single Page Applications (SPAs) and Multi-Page Applications (MPAs) through a server-driven architecture. Your entire application logic lives in Go - no APIs to build, no client state to synchronize, Go code directly manipulates the HTML.


### 1. **Modern like SPA**
- Composable components 
- Dynamic updates without page refreshes
- Built-in reactivity system

### 2. **Straight like MPA**
- Natural server-side rendering
- Fast initial load and functional `href` links
- Real `FormData` handling
- Execution transparency

### 3. **API-Free Architecture**
- No REST/GraphQL APIs needed
- Static endpoints only for files and pages
- Everything else wrapped and secured by the framework

### 4. **NPM-Free Development**
- No Door.js required to build or run
- Optional JavaScript/TypeScript integration when needed
- Built-in esbuild for processing

### 5. **Type-Safe Throughout**
- From DOM events to routing

## How It Works

### Stateful Server + Ultra-Thin Client

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

## Philosophy 

### Explicid
Build direct connections between events, state and HTML in completely type-safe environment. *It hits different*.

### Lightweight 
Fast loading, efficient syncronization, non-blocking execution environment

### Server Centric
Business logic runs on the server, browser acts like a human interface. 

### Straight
Native expirience of classic MPA with natural reactive UI capabilities.

### JS friendly
If you need - integration, bundling and serving tools included.


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

Unlike **React/Vue**: No business logic on the client side, no hydration, no npm dependencies.
Unlike **htmx**: Full type safety, reactive state, and programmatic control from Go.
Unlike **Phoenix LiveView**: Go instead of Elixir, explicit update model, superior concurrency model and QUIC friendly

## Pricing

doors will be a **paid product** with developer-friendly terms:

- ✅ **Affordable lifetime license** 
- ✅ **Source code available on GitHub** 
- ✅ **Free for development** 
- ✅ **No telemetry**


## Why Paid?

This isn't backed by big tech or VCs. It's a focused effort to build the best possible tool without divided interests. Sustainable funding ensures long-term commitment and continuous improvement.

## Status

⚠️ **Early Preview** - Core functionality complete, API may change. Not recommended for production yet.
