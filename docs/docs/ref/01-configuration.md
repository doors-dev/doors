# System Configuration 

Provides system-wide configuration options for controlling sessions, instances, client-server communication, and performance settings. Defaults are automatically initialized. 

## Apply

Configuration applied to the `doors.Router` with `doors.UseSystemConf` mod

```templ
router.Use(doors.UseSystemConf(doors.SystemConf{
	SessionInstanceLimit: 6,
}))
```

## Fields

```templ
type SystemConf struct {
	SessionInstanceLimit     int
	SessionExpiration        time.Duration
	SessionCookieExpiration  time.Duration
	InstanceGoroutineLimit   int
	InstanceTTL              time.Duration
	ServerDisableGzip        bool
	DisconnectHiddenTimer    time.Duration
	RequestTimeout           time.Duration
	SolitairePing            time.Duration
	SolitaireSyncTimeout     time.Duration
	SolitaireRollTimeout     time.Duration
	SolitaireFlushSizeLimit  int
	SolitaireFlushTimeout    time.Duration
	SolitaireQueue           int
	SolitairePending         int
}

```
> If you’d like certain options configurable per instance or dynamicaly, please open a GitHub issue.

#### Resource Control

Key settings to understand:

- **SessionInstanceLimit** — Max page instances per session. Oldest inactive suspended if exceeded.
   *Default: 12*
- **InstanceGoroutineLimit** — Max goroutines per instance for rendering and reactivity.
   *Default: 16*
- **InstanceTTL** — Lifetime of inactive instances before cleanup.
   *Default: 40minutes or ≥ 2× `RequestTimeout`*
- **DisconnectHiddenTimer** — Time hidden/background tabs stay connected.
   *Default: InstanceTTL ÷ 2*

### Session Management

- **SessionExpiration** — Lifetime of a session. If >0, session expires even if has active instances.
   *Default: 0 (expires when no instances remain)*

  **SessionCookieExpiration** — Browser session cookie lifetime.
   *Default: 0 (expires when browser closes)*

### Solitaire Protocol (syncronization)

#### Control Syncronization Issues 

- **SolitaireSyncTimeout** — Max pending duration of a server→client sync calls; **exceeding kills instance**.
   *Default: InstanceTTL*
- **SolitaireQueue** —  Max queued server→client sync calls; **exceeding kills instance**.
   *Default: 1024*
- **SolitairePending** —  Max unresolved server→client sync calls; throttles sending when reached.
   *Default: 256*

#### Control Network Behavior

- **SolitairePing** — Max idle time before rolling the request. *Default: 15s*
- **SolitaireRollTimeout** — How long an active sync connection lasts before rolling (affects if sync queue is long).
   *Default: 1s*
- **SolitaireFlushSizeLimit** — Max written bytes before forcing a flush.
   *Default: 32 KB*
- **SolitaireFlushTimeout** — Max delay before forcing a flush.
   *Default: 30ms*

### Other

* **RequestTimeout** — Max duration of a client-server request (hook).
   *Default: 30s*
* **ServerDisableGzip** — Disables gzip compression when true. Applies to HTML, JS, and CSS.
   *Default: false*.

