# Scopes

The Scopes API defines **how concurrent events are coordinated**.  
Scopes control queuing, debouncing, and blocking so that UI and backend events execute in a predictable and safe order.

---

## Concept

- The `Scope` field in hook attributes accepts a slice of scopes (a **scope pipeline**).  
- **Simple usage:** one-scope pipeline via helpers like `doors.ScopeOnlyBlocking()`.  
- **Advanced usage:** create reusable scope instances and combine them.  
- A scope can **hold**, **promote**, or **cancel** an event.  
- If all scopes allow the event, it proceeds to the backend.  
- Scopes can be **shared** across multiple handlers to coordinate their concurrency.

---

## Scope Types

### Blocking

Cancels new events while one is processing.  
Prevents double-clicks or duplicate submissions.

**Simple:**

```templ
@doors.AClick{
  Scope: doors.ScopeOnlyBlocking(),
}
```

**Advanced:**

```templ
{{ block := &doors.ScopeBlocking{} }}
@doors.AClick{
  Scope: []doors.Scope{block},
}
```

---

### Serial

Queues events and processes them sequentially in arrival order.

**Simple:**

```templ
@doors.AClick{
  Scope: doors.ScopeOnlySerial(),
}
```

**Advanced:**

```templ
{{ serial := &doors.ScopeSerial{} }}
@doors.AClick{
  Scope: []doors.Scope{serial},
}
```

---

### Debounce

Delays event execution to prevent rapid bursts.  
New events reset the timer; execution is guaranteed after the limit even if activity continues.

**Simple:**

```templ
@doors.AClick{
  Scope: doors.ScopeOnlyDebounce(300*time.Millisecond, time.Second),
}
```

**Advanced:**

```templ
{{ debounce := &doors.ScopeDebounce{} }}
@doors.AClick{
  Scope: []doors.Scope{debounce.Scope(300*time.Millisecond, time.Second)},
}
```

---

### Frame

Separates **immediate** and **frame** events.  
Frame events wait for all prior events to complete, block new ones, and then execute exclusively.

**Example:**

```templ
{{ frame := &doors.ScopeFrame{} }}

@doors.AInput{
  Scope: []doors.Scope{frame.Scope(false)}, // immediate
}
<input>

@doors.AClick{
  Scope: []doors.Scope{frame.Scope(true)}, // frame
}
<button>Submit</button>
```

---

### Concurrent

Allows simultaneous events but isolates them by group ID.  
Events with different groups block each other.

**Example:**

```templ
{{ scope := &doors.ScopeConcurrent{} }}

@doors.AInput{
  Scope: []doors.Scope{scope.Scope(1)}, // group 1
}
<input>

@doors.AClick{
  Scope: []doors.Scope{scope.Scope(0)}, // group 0
}
<button>Submit</button>
```

---

## Scope Pipelining

Combine multiple scopes into a pipeline; each applies sequentially.

```templ
{{ frame := &doors.ScopeFrame{} }}
{{ debounce := &doors.ScopeDebounce{} }}

@doors.AInput {
  Scope: []doors.Scope{
    frame.Scope(false),
    debounce.Scope(300*time.Millisecond, 0),
  }
}
<input>

@doors.AClick {
  Scope: []doors.Scope{frame.Scope(true)}
}
<button>Submit</button>
```

This ensures:

- Input is debounced.  
- Click waits for the input’s frame to complete.  
- Submission happens only after input is processed.

> ⚠️ Overlapping or conflicting scopes can lead to unintended queuing or blocking behavior. Test scope pipelines carefully.
