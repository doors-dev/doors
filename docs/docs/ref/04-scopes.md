# Scopes

The Scopes API provides concurrency control for event processing. Scopes determine how multiple events are queued, debounced, blocked, or serialized to ensure correctness and predictable user-facing behavior

## Advanced Concept

* `Scope` field on the hook attribute accepts a slice of scopes (scope pipeline). 
* **You can create a simple one-scope pipeline with helper functions, such as `doors.ScopeBlocking()`** or enable complex scope control with `[]doors.Scope{/* scopes */}`
* Each scope in a pipeline can **hold**, **promote** an event to the next scope, and **cancel** it (even after promotion)
* Scope considers event processing is finished when the corresponding request is executed or the event is canceled 
* When the event successfully clears all scopes, it is passed to the backend.
* **Scopes can be shared between different handlers to control group behaviour**

## Scope Types

### Blocking

Prevents concurrent handling of events in the same scope. If an event is already processing, subsequent events are cancelled until the current one is completed. Suitable for preventing double-clicks or rapid repeated submissions.

**Simple usage:**

```templ
@doors.AClick{
  Scope: doors.ScopeBlocking(),
	/* setup */
}
```

**Advanced control:**

```templ
// initialized scope, can be reused across handlers
{{ block := &doors.BlockingScope{} }}

@doors.AClick{
  Scope: []doors.Scope{ block },
  /* setup */
}
```

### Debounce

Delays handling of rapid event bursts using a two-parameter debounce: a *delay* (`duration`) that resets on new events, and a hard *limit* (`limit`) after which execution proceeds regardless of further activity.

**Simple usage:**

```templ
@doors.AClick{
  Scope: doors.ScopeDebouce(300 * time.Millisecond, time.Second),
	/* setup */
}
```

**Advanced control:**

```templ
// initialized scope, can be reused across handlers
{{ debounce := &doors.DebounceScope{} }}

@doors.AClick{
  Scope: []doors.Scope{ debounce.Scope(300 * time.Millisecond, time.Second) },
  /* setup */
}
```

### Latest

Cancels any in-flight or queued work and processes only the most recent event. 

**Simple usage:**

```templ
@doors.AInput{
  Scope: doors.ScopeLatest(),
  /* setup */
}
```

**Advanced control:**

```templ
// initialized scope, can be reused across handlers
{{ latest := &doors.LatestScope{} }}

@doors.AInput{
  Scope: []doors.Scope{ latest },
  /* setup */
}
```

### Serial

Queues events and processes them one at a time in arrival order. Ensures that ordering is preserved. 

**Simple usage:**

```templ
@doors.AClick{
  Scope: doors.ScopeSerial(),
  /* setup */
}
```

**Advanced control:**

```templ
// initialized scope, can be reused across handlers
{{ serial := &doors.SerialScope{} }}

@doors.AClick{
  Scope: []doors.Scope{ serial },
  /* setup */
}

```

### Frame

Manages two types of events: immediate events and frame events. Immediate `events.Scope(false)`  executed normally. Frame events `.Scope(true)` **wait until all previous events (immediate and frame) in the scope complete, while blocking new events**, and then executes normally (non-blocking).

**Simple usage:** -
**Advanced control:**

```temp
// initialized scope, shared across handlers
{{ frame := doors.FrameScope{} }}

// Immediate action
@doors.AInput{
  Scope: []doors.Scope{ frame.Scope(false) },
  /* setup */
}

// Frame (terminalizing) action
@doors.AClick{
  Scope: []doors.Scope{ frame.Scope(true) },
  /* setup */
}
```

## Scope Pipelining

Enables sophisticated concurrency control by combining scopes into a pipeline. You can think of each scope as a gatekeeper that can allow events to enter, manage the queue, or instruct them to "get lost".

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
	Scope: []doors.Scope{ &doors.BlockingScope{}, frame.Scope(true) }
}
<button>Submut</button>
```

In this example, we ensured that a button click event reaches the server after input processing is complete (because framing "true" scope waits until all framing "false" processing is finished). This technique is useful for enabling highly dynamic forms, where each field is processed individually on the server.

> Scope pipelining is an advanced feature that requires special attention. It's very easy to cause unpredictable behaviour with conflicting combinations.

