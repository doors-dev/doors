# System Configuration 

Provides system-wide configuration options for controlling sessions, instances, client-server communication, and performance settings. Defaults are automatically initialized. 

## Apply

Configuration applied to the `doors.Router` with `doors.SetSystemConf` mod

```templ
router.Use(doors.SetSystemConf(doors.SystemConf{
	SessionInstanceLimit: 12,
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
    DisconnectHiddenTimer   time.Duration
    SolitaireRollSize        int
    SolitaireRequestTimeout  time.Duration
    SolitaireRollPendingTime time.Duration
    SolitaireQueue           int
    SolitairePending         int
}

```

> If you want to have specific options configurable on an instance or session basis, please let me know by opening a GitHub issue.

#### Resource Control

Important settings you need to understand 

- **SessionInstanceLimit**: Maximum number of instances per session. Oldest/least active suspended if exceeded. Default: 6.
- **InstanceGoroutineLimit**: Maximum goroutines per instance for rendering/reactivity. Default: 16. 
- **InstanceTTL**: Lifetime of disconnected instances before termination. Default: 15 min or ≥ 2× `SolitaireRequestTimeout`.
- **DisconnectHiddenTimer**: If the tab is hidden, keep the page connected for the specified time. Default: 10 min. 

### Session Management

If you utilize session storage, you probably want the session to persist.

- **SessionExpiration**: Timeout for session inactivity (no active instances) cleanup. Default: 0 (cleaned when no instances).
- **SessionCookieExpiration**: Expiry for browser session cookie. Default: 0 (expires when browser closes).

### Sync protocol

Usually, you don't need to touch those.

- **SolitaireQueue**: Max queued client calls (sync operations) **before instance termination.** Default: 1024.
- **SolitaireRollSize**: Max response size before rolling to a new request. Default: 8 KB.
- **SolitaireRequestTimeout**: Max duration a request can remain open. Default: 30s. 1/2 of the timeout is used to signal rolling. **Also used as a hook timeout.**
- **SolitaireRollPendingTime**: Wait time before rolling to a new request if pending calls exist.  Default: 100 ms.
- **SolitairePending**: Max unacknowledged calls in flight. Default: 256.

### Other

* **ServerDisableGzip**: Disables gzip compression when true. Default: false. Applied to pages, scripts, and styles.

