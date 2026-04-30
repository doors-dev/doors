# Scopes

Scopes control what happens when events overlap.

They run on the client, before the backend request begins. That is why scopes are the right tool for interaction policy such as:

- preventing double-clicks
- queuing repeated actions
- debouncing typing
- making one action wait for a related group of actions
- keeping only the newest interaction

Without scopes, every event is free to proceed as soon as it fires.

## Why

Think about a few common UI problems:

- a submit button should not run twice
- three rapid clicks should run one after another
- a search box should wait until typing settles
- a final "apply" action should wait for earlier edits to finish
- a preview request should drop stale in-flight work and keep only the newest request

Scopes exist to express those rules directly in the event attribute instead of rebuilding them by hand in handlers.

## Basics

Scopes are attached through the `Scope` field on event attributes. The field type is `doors.Scopes` — a single value that may carry one or several scopes.

The reusable scope types are:

- `doors.ScopeBlocking`
- `doors.ScopeSerial`
- `doors.ScopeDebounce`
- `doors.ScopeLatest`
- `doors.ScopeFrame`      (use `.Scope(frame bool)` to get a `Scopes`)
- `doors.ScopeConcurrent` (use `.Scope(groupID int)` to get a `Scopes`)

Each one is a `Scopes` value on its own — you can pass `&doors.ScopeBlocking{}` directly to `Scope`. Combine several with `.And(...)` or `doors.JoinScopes(...)`:

```go
Scope: blocking.And(debounce).And(serial)

// or
Scope: doors.JoinScopes(blocking, debounce, serial)
```

Use a fresh instance when one handler just needs one rule. Reuse a scope instance across handlers when they should coordinate with each other.

## Sharing

Sharing a scope instance means the handlers coordinate with each other instead of each acting independently.

```gox
<>
	~{
		block := &doors.ScopeBlocking{}
	}

	<button
		(doors.AClick{
			Scope: block,
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				return false
			},
		})>
		One
	</button>

	<button
		(doors.AClick{
			Scope: block,
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				return false
			},
		})>
		Two
	</button>
</>
```

Here both buttons share one blocking rule, so only one of their clicks can proceed at a time.

## Blocking

`ScopeBlocking` cancels a new event if another event in that shared scope is already running.

Use it for:

- submit buttons
- destructive actions
- anything that should not run twice in parallel

This is the simplest "prevent double-submit" scope.

`ScopeBlocking` is client-side interaction policy, not a backend permission or one-shot guarantee. Calls to the same hook instance are already serialized on the backend. If a hook should run once and then disappear, return `true` from its handler.

## Serial

`ScopeSerial` queues accepted events and starts them in arrival order. The first hook starts immediately; each later hook starts only after the previous one reports completion.

Use it when every accepted event should still run, just not at the same time.

Typical case - repeated actions that must preserve order.

Unlike blocking, serial does not drop later events. It holds them and runs them one by one.

## Debounce

`ScopeDebounce` keeps the latest pending event in a burst and delays execution.

```gox
<>
	~{
		debounce := &doors.ScopeDebounce{
			Duration: 300 * time.Millisecond,
			Limit:    600 * time.Millisecond,
		}
	}

	<input
		(doors.AInput{
			Scope: debounce,
			On: func(ctx context.Context, r doors.RequestEvent[doors.InputEvent]) bool {
				return false
			},
		})/>
</>
```

Fields:

- `Duration`: the resettable wait time
- `Limit`: the maximum total wait; `0` means no limit

Use it for:

- search boxes
- live filters
- expensive input-driven updates

If a new event arrives before the debounce fires, the previous pending one is canceled and the new one takes its place.

Without a limit, only the final burst event runs. With a limit, execution still happens even if new events keep arriving.

## Frame

`ScopeFrame` lets you separate normal events from a barrier event.

- `frame.Scope(false)` is a normal event in that frame scope
- `frame.Scope(true)` is a frame event

A frame event waits until earlier events in the same shared frame scope finish. Once that frame event is pending or running, new events in that same frame scope are blocked.

This is useful when one action should act like "stop here, then do this exclusively."

Example:

```gox
<>
	~{
		frame := &doors.ScopeFrame{}
		debounce := &doors.ScopeDebounce{
			Duration: 300 * time.Millisecond,
			Limit:    600 * time.Millisecond,
		}
	}

	<input
		(doors.AInput{
			Scope: frame.Scope(false).And(debounce),
			On: func(ctx context.Context, r doors.RequestEvent[doors.InputEvent]) bool {
				return false
			},
		})/>

	<button
		(doors.AClick{
			Scope: frame.Scope(true),
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				return false
			},
		})>
		Apply
	</button>
</>
```

Here the input events are normal frame members, and `Apply` is the barrier event.

## Concurrent

`ScopeConcurrent` allows overlap only inside the same group id.

If the scope is already occupied by one group, an event from a different group is canceled.

That makes it useful when several related controls may work together, but another action must not overlap with them.

Example:

```gox
<>
	~{
		scope := &doors.ScopeConcurrent{}
	}

	<button
		(doors.AClick{
			Scope: scope.Scope(1),
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				return false
			},
		})>
		Country
	</button>

	<button
		(doors.AClick{
			Scope: scope.Scope(1),
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				return false
			},
		})>
		City
	</button>

	<button
		(doors.AClick{
			Scope: scope.Scope(0),
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				return false
			},
		})>
		Confirm
	</button>
</>
```

Here `Country` and `City` belong to the same group, so they can overlap with each other. `Confirm` is in a different group, so it is blocked while group `1` is active, and vice versa.

This pattern comes up in multi-step selectors and other compound controls.

## Latest

`ScopeLatest` cancels the previous event and keeps only the newest one.

Use it when only the current interaction matters.

Typical cases:

- preview requests
- selection changes
- pointer-driven movement
- interactions where stale work should be dropped immediately

Compared to debounce:

- debounce waits before sending anything
- latest can replace work that is already in progress

Canceling an in-flight hook is a client-side effect. The request may already have reached the server, so do not rely on `ScopeLatest` to prevent handler execution or protect writes. It is most useful for ending previous indication, ignoring stale results, and keeping the newest interaction in control of the UI.

## Pipelines

Scopes are joinable, so they can be chained into a pipeline with `.And(...)` (or `doors.JoinScopes(...)`).

```gox
<>
	~{
		frame := &doors.ScopeFrame{}
		debounce := &doors.ScopeDebounce{
			Duration: 150 * time.Millisecond,
		}
		serial := &doors.ScopeSerial{}
	}

	<button
		(doors.AClick{
			Scope: frame.Scope(false).And(debounce).And(serial),
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				return false
			},
		})>
		Run
	</button>
</>
```

The order matters.

Each scope sees the event only after the previous scope accepted it. That lets you build pipelines like:

- debounce, then serialize
- normal frame members plus one barrier frame event
- blocking shared across several controls

## Rules

- Scopes are client-side. They shape whether and when a request is sent.
- Use a helper for one simple scope.
- Reuse a scope instance when several handlers should coordinate with each other.
- Use blocking to drop overlap, serial to queue overlap, and debounce to delay bursts.
- Use frame when one action should wait for earlier related actions and then run exclusively.
- Use concurrent when overlap is allowed only inside one group.
- Use latest when stale work should be canceled in favor of the newest event.
- Scopes pair naturally with indication: scopes decide whether the request proceeds, indication shows the interaction state.
