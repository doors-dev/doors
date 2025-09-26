# Scopes

The Scopes API provides concurrency control for event processing. Scopes determine how multiple events are queued, debounced, blocked, or serialized to ensure correctness and predictable user-facing behavior.

## Concept

* The `Scope` field on a hook attribute accepts a slice of scopes (a **scope pipeline**).  
* **Simple usage**: one-scope pipeline via helpers like `doors.ScopeOnlyBlocking()`.  
* **Advanced usage**: construct reusable scope objects (e.g. `&doors.ScopeBlocking{}`) and combine them into a pipeline.  
* Each scope can **hold**, **promote**, or **cancel** an event.  
* Event processing is considered complete when the request is executed or canceled.  
* If an event clears all scopes, it is sent to the backend.  
* **Scopes can be shared** between handlers to coordinate behavior across them.

## Scope Types

### Blocking

Cancels new events while one is processing. Prevents double-clicks or repeated submissions.

**Simple usage:**

```templ
@doors.AClick{
  Scope: doors.ScopeOnlyBlocking(),
}
```

**Advanced usage:**

```templ
{{ block := &doors.ScopeBlocking{} }}

@doors.AClick{
  Scope: []doors.Scope{block},
}
```

### Serial

Queues events and processes them sequentially in arrival order.

**Simple usage:**

```templ
@doors.AClick{
  Scope: doors.ScopeOnlySerial(),
}
```

**Advanced usage:**

```templ
{{ serial := &doors.ScopeSerial{} }}

@doors.AClick{
  Scope: []doors.Scope{serial},
}
```

### Debounce

Delays handling of rapid bursts of events.  
New events reset the delay timer; execution is guaranteed after the *limit* even if activity continues.

**Simple usage:**

```templ
@doors.AClick{
  Scope: doors.ScopeOnlyDebounce(300*time.Millisecond, time.Second),
}
```

**Advanced usage:**

```templ
{{ debounce := &doors.ScopeDebounce{} }}

@doors.AClick{
  Scope: []doors.Scope{debounce.Scope(300*time.Millisecond, time.Second)},
}
```

### Frame

Separates immediate and frame events.  

* `frame=false`: event runs immediately.  
* `frame=true`: waits for all previous events, blocks new ones, then runs exclusively.

**Advanced usage:**

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

### Concurrent

Can be "occupied" only by events with the same group id.

```templ
{{ scope := &doors.ScopeConcurrent{} }}

@doors.AInput{
  Scope: []doors.Scope{scope.Scope(1)}, // group 1
}
<input>

@doors.AClick{
  Scope: []doors.Scope{frame.Scope(0)},  // group 0
}
<button>Submit</button>

```



## Scope Pipelining

Combine multiple scopes to form a pipeline. Each scope is applied in sequence.

```templ
{{ frame := doors.FrameScope{} }}
{{ debounce = doors.DebounceScope{} }}

@doors.AInput {
  // pass throught frame scope, then apply debounce (300 milliseconds, no limit)
	Scope: []doors.Scope{ frame.Scope(false), debounce.Scope(300 * time.Millisecond, 0) }
/* setup */
}
<input type="text" name="name">
@doors.AClick {
  // Termination frame, to ensure that input value will be passed to server
	Scope: []doors.Scope{ frame.Scope(true) }
}
<button>Submut</button>
```

In this example:

* Input is debounced.  
* Click event waits for input to finish (frame termination) and then blocks everything until completion.  

This guarantees that input is processed before submission.  

> ⚠️ Scope pipelining is powerful but easy to misuse. Conflicting scopes may lead to unexpected behavior.
