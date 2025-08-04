# doors

⚠️ **Early Preview - Not Ready for Production**

A next-gen web framework that combines the best of SPAs and MPAs, built in Go with a server-driven stateful architecture.

## The Solution

### 1. **Modern like SPA**
- Composable components
- Shared and local state management
- Built-in reactivity

### 2. **Straight like MPA**
- Natural SSR (not running frontend code on server)
- Fast loading, functional `href`
- Real `FormData` 
- Execution transparency

### 3. **API-Free Architecture**
- Static endpoints only for files and pages
- Everything else wrapped and secured by the framework

### 4. **NPM-Free**
- Not required to write or run
- Optional integration when needed

### 5. **Server-Driven**
- Business logic stays on the server
- User interactions handled in browser
- Clean separation of concerns

### 6. **Self-Host Friendly**
- Batteries included
- No vendor lock-in

### 7. **UX First**
- Genuine lightweight SPA feel
- No compromise on user experience

## How It Works

### Stateful Server + Ultra-Thin Client

1. **User loads page** → Server assigns session cookie and initializes page instance
2. **Template rendering** → Server collects handler functions and prepares event bindings
3. **Browser connection** → Lightweight client connects for real-time sync
4. **Event handling** → Server routes requests via session to specific handlers
5. **State updates** → Changes trigger layout updates 

### Dynamic Component Architecture

Clean separation of concerns:

- **Node (Container)**: Dynamic HTML placeholders managed in a tree structure
- **Beam (State)**: Reactive state primitives that can be subscribed to and combined
- **Fragment (Component)**: Composable entities with render methods


### Real-Time Sync Without WebSockets

doors uses "rolling handover request" system:
- Client sends regular HTTP requests
- Server streams HTML fragments and JS calls with metadata
- Client opens new request with batched results
- Server seamlessly switches streams
- Cycle continues with minimal latency

### Type-Safe Routing
- Declarative, reactive routing system
- Path parameters deserialized into annotated structs
- Automatic parameter parsing and propagation
- Each page is a mini-app with dynamic updates

### Event Handling
- Pointer, Keyboard, Form, and Input events capture support
- Secure, session-scoped execution of hooks

### Built-In Tools
- Static asset serving (public or session-scoped)
- JS/TS integration via embedded esbuild
- Resource and session management
- CSP header automation
- Frontend state indicators and concurrency control

## Pricing Philosophy

doors will be a **paid product** with developer-friendly terms:

- ✅ **Affordable lifetime license** (no subscription)
- ✅ **Source code available** on GitHub
- ✅ **Free for development** - pay only for production
- ✅ **No telemetry** or data collection

## Why Paid?

This isn't backed by big companies or VC funding. It's a focused effort to build the best possible tool without serving other interests. Quality development requires dedicated professionals, and sustainable funding ensures long-term commitment to the project.


