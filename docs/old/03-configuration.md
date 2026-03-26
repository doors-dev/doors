# System Configuration

Provides system-wide configuration options for configuring sessions, instances, client-server communication, and performance settings. Defaults are automatically initialized.

## Apply

Applied to a router with `doors.UseSystemConf` modifier:

```templ
router.Use(doors.UseSystemConf(doors.SystemConf{
	SessionInstanceLimit: 6,
}))
```

## Fields

```go
type SystemConf struct {
	SessionInstanceLimit     int
	SessionTTL               time.Duration
	InstanceGoroutineLimit   int
	InstanceConnectTimeout   time.Duration
	InstanceTTL              time.Duration
	ServerCacheControl       string
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

> If you’d like certain options configurable per instance or dynamically, please open a GitHub issue.

---

### Resource Control

- **SessionInstanceLimit** — Max page instances per session. Oldest inactive suspended if exceeded.  
  *Default: 12*

- **SessionTTL** — Session lifetime policy.  
  *Default: 0*  
  Behavior at `0`: session ends when **no instances remain**, and the session cookie expires when the **browser closes**.

- **InstanceConnectTimeout** — How long a new instance waits before shutdown for the first client connection.  
  *Default: RequestTimeout*

- **InstanceGoroutineLimit** — Max goroutines per instance for rendering and reactivity.  
  *Default: 16*

- **InstanceTTL** — Lifetime of inactive instances before cleanup.  
  *Default: 40 minutes or ≥ 2× `RequestTimeout`*  
  If `InstanceTTL` < 2× `RequestTimeout`, it is automatically raised to that value.

- **DisconnectHiddenTimer** — Time hidden/background tabs stay connected.  
  *Default: InstanceTTL ÷ 2*

---

### Solitaire Protocol (Synchronization)

#### Control Synchronization Issues

- **SolitaireSyncTimeout** — Max pending duration of a server→client sync call; exceeding kills the instance.  
  *Default: InstanceTTL*

- **SolitaireQueue** — Max queued server→client sync calls; exceeding kills the instance.  
  *Default: 1024*

- **SolitairePending** — Max unresolved server→client sync calls; throttles sending when reached.  
  *Default: 256*

#### Control Network Behavior

- **SolitairePing** — Max idle time before rolling the request.  
  *Default: 15s*

- **SolitaireRollTimeout** — Duration an active sync connection lasts before rolling (affects long queues).  
  *Default: 1s*

- **SolitaireFlushSizeLimit** — Max written bytes before forcing a flush.  
  *Default: 32 KB*

- **SolitaireFlushTimeout** — Max delay before forcing a flush.  
  *Default: 30ms*

---

### Other

- **RequestTimeout** — Max duration of a client-server request (hook).  
  *Default: 30s*

- **ServerCacheControl** — Cache control header for JS and CSS resources prepared by the framework.  
  *Default: `public, max-age=31536000, immutable`*

- **ServerDisableGzip** — Disables gzip compression when true. Applies to HTML, JS, and CSS.  
  *Default: false*

---

### Defaults Overview

| Setting                 | Default                             |
| ----------------------- | ----------------------------------- |
| SessionInstanceLimit    | 12                                  |
| SessionTTL              | 0                                   |
| InstanceGoroutineLimit  | 16                                  |
| InstanceConnectTimeout  | = RequestTimeout                    |
| InstanceTTL             | 40m or ≥ 2× RequestTimeout          |
| DisconnectHiddenTimer   | InstanceTTL ÷ 2                     |
| RequestTimeout          | 30s                                 |
| SolitairePing           | 15s                                 |
| SolitaireRollTimeout    | 1s                                  |
| SolitaireFlushSizeLimit | 32 KB                               |
| SolitaireFlushTimeout   | 30ms                                |
| SolitaireQueue          | 1024                                |
| SolitairePending        | 256                                 |
| ServerCacheControl      | public, max-age=31536000, immutable |
| ServerDisableGzip       | false                               |
