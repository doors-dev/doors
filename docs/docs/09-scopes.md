# Scopes

Scopes shape how events behave on the client before backend work begins.

They are part of the event attribute pipeline and are primarily a client-side coordination mechanism.

## Scopes

Scopes decide how event requests are allowed to proceed.

They run in the browser before the request is sent. So they do more than describe server ordering:

- they can cancel an event immediately
- they can hold it
- they can queue it
- they can replace it with a newer one

This is why scopes are the right tool for interaction policy such as double-click prevention, debouncing, and coordination between related controls.

## Helpers

Simple helpers:

- `doors.ScopeOnlyBlocking()`
- `doors.ScopeOnlySerial()`
- `doors.ScopeOnlyDebounce(duration, limit)`
- `doors.ScopeOnlyLatest()`

Reusable scope types:

- `doors.ScopeBlocking`
- `doors.ScopeSerial`
- `doors.ScopeDebounce`
- `doors.ScopeFrame`
- `doors.ScopeConcurrent`
- `doors.ScopeLatest`

Use the helper form for one simple scope.  
Use the reusable struct form when several handlers must share the same coordination state.

## Sharing

Sharing a scope instance means several event handlers participate in the same client-side coordination rule.

```gox
<>
	~{
		block := &doors.ScopeBlocking{}
	}

	<button
		(doors.AClick{
			Scope: []doors.Scope{block},
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				return false
			},
		})>
		One
	</button>

	<button
		(doors.AClick{
			Scope: []doors.Scope{block},
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				return false
			},
		})>
		Two
	</button>
</>
```

Here both buttons share the same blocking behavior.

## Blocking

`ScopeBlocking` cancels new events while one is already in progress.

Use it for:

- submit buttons
- destructive actions
- anything that should not run twice in parallel

## Serial

`ScopeSerial` queues events and runs them in arrival order.

Use it for:

- ordered mutations
- append-style workflows
- cases where every accepted event should run eventually

## Debounce

`ScopeDebounce` keeps the latest event in a burst and delays execution.

```gox
<>
	~{
		debounce := &doors.ScopeDebounce{}
	}

	<input
		(doors.AInput{
			Scope: []doors.Scope{
				debounce.Scope(300 * time.Millisecond, 600 * time.Millisecond),
			},
			On: func(ctx context.Context, r doors.RequestEvent[doors.InputEvent]) bool {
				return false
			},
		})/>
</>
```

Parameters:

- `duration`: resettable delay
- `limit`: maximum total wait; `0` means no limit

Use it for:

- search boxes
- live filters
- expensive input-driven updates

With no limit, only the final burst event runs. With a limit, execution still happens even if new events keep arriving.

## Frame

`ScopeFrame` separates normal events from frame events.

```gox
<>
	~{
		frame := &doors.ScopeFrame{}
	}
</>
```

- `frame.Scope(false)` means normal event
- `frame.Scope(true)` means exclusive frame event

A frame event waits until earlier events in that shared scope finish. While a frame event is pending or running, new events in that same frame scope are blocked.

Use it when one event should act as a barrier or synchronization point for a group of related interactions.

## Concurrent

`ScopeConcurrent` allows concurrent processing only for the same group id.

```gox
<>
	~{
		scope := &doors.ScopeConcurrent{}
	}

	<button (doors.AClick{Scope: []doors.Scope{scope.Scope(1)}, On: on1})>A</button>
	<button (doors.AClick{Scope: []doors.Scope{scope.Scope(1)}, On: on1})>B</button>
	<button (doors.AClick{Scope: []doors.Scope{scope.Scope(2)}, On: on2})>C</button>
</>
```

Handlers in group `1` can overlap with each other. A different group is blocked while that group is active.

## Latest

`ScopeLatest` cancels the previous running event and keeps only the newest one.

Use it when only the latest interaction matters.

Typical cases:

- pointer-driven movement
- preview requests
- selection changes where stale work should be dropped

## Pipelines

Scopes can be combined in order.

```gox
<>
	~{
		frame := &doors.ScopeFrame{}
		debounce := &doors.ScopeDebounce{}
		serial := &doors.ScopeSerial{}
	}

	<button
		(doors.AClick{
			Scope: []doors.Scope{
				frame.Scope(false),
				debounce.Scope(150 * time.Millisecond, 0),
				serial,
			},
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				return false
			},
		})>
		Run
	</button>
</>
```

Each scope sees the event after the previous scope has accepted it.

This lets you build coordination pipelines such as:

- debounce, then serialize
- normal events plus an exclusive frame event
- shared blocking across several controls

## Notes

- Scopes are evaluated on the client before the request runs.
- Indication has its own doc: [10-indication.md](/Users/alex/Lib/doors/docs/docs/10-indication.md).
- Scopes and indication work well together: scopes decide whether the request proceeds, indication shows the current interaction state.
