# Important

Practices you should understand and follow. 

## 1. Use *doors* entities inside *doors* context.

✅ Render `Door`, read/mutate `Beam/SourceBeam` values, use `doors.A...` binds and handlers in **pages served by doors.ServePage**

❌ Try to enable interactivity/reactivity with *doors* in self-served templ-based pages

**You will get a panic if you try.**

## 2. Respect blocking free context and data propagation.

Rendering operations and Beam handlers are executed in the framework's runtime. 

* Query data and make requests to external systems during rendering or, less preferably, in `Beam` subscription handlers.
* Use manually spawned (or via `@doors.Go`) goroutines, or, less preferably, hook handlers to wait on channels, asynchronous events, or perform any long-running operations.

✅  Best: query data during render: 

```templ
templ (c *card) Render() {*
  // predictable wait time
	{{ item := db.get(c.id) }}
	/* use item in render */
}
```

⚠️ Acceptable: query data in subscription

```templ
pathBeam.Sub(ctx, func(ctx context.Context, p Path) bool {
	item := db.get(c.id)
	door.Update(ctx, itemInfo(item))
}
```

> Querying data in the beam handler will delay the propagation of **beam** data to nested elements; it's acceptable.

✅  Spawn goroutine for long-running tasks:

```templ
templ (f *fragment) Render() (
    // spawned goroutine
		@doors.Go(func(ctx context.Context) {
        select {
        // after 1 minute update
        case <-time.After(1 * time.Minute):
          f.door.Update(ctx, doors.Text("Updated after 1 minute!"))
        // or cancel if unmounted
        case <-ctx.Done():
          return
        }
		})
		@f.door {
				Initial State
		}
)
```

❌ Bad 

*  Wait for external systems events during render or in Beam subscription handlers
*  Try to synchronize rendering operations with each other by blocking in render functions

❌  Block framework runtime to wait for other runtime-dependent operations to complete, like :

```temp
pathBeam.Sub(ctx, func(ctx context.Context, p Path) bool {
  // XUpdate returns channel to track when frontend applies update
	err, ok := <- door.XUpdate(ctx, itemInfo(item))  // can cause a deadlock!
	/* ... */
	return false
}
```

✅   Instead, do this:

```temlp
pathBeam.Sub(ctx, func(ctx context.Context, p Path) bool {
// spawn independent runtime goroutine
	go func() {
			// tell the framework that you are safe
			blockingCtx := doors.AllowBlocking(ctx)
			err, ok := <- door.XUpdate(blockingCtx, itemInfo(item))
			/* ... */
	}()
	return false
}
```

❌  Block the render like:

```templ
templ (f *fragment) Render() {
 		// blocking receive from pubsub topic
		{{ msg, _ := <- f.pubsub.Channel() }}
		// print payload
		{ msg.Payload }
}
```

✅  Instead, do this:

```templ 
templ (f *fragment) Render() {
	// initialize door
	{{ door := doors.Door{} }}
	// spawn goroutine
	@doors.Go(func(ctx context.Context) {
	   select {
        // blocking recieve from pubsub topic 
        case msg, _ := <- f.pubsub.Channel():
           n.update(doors.Text(msg.Payload))
        // or cancel if unmounted
        case <-ctx.Done():
          return
     }
	})
	// render door
	@door
}
```

## 3. Understand the security model.

* For protected pages, **verify cookie authentication in the `ServePage` handler**
* **Don't forget to call `doors.SessionEnd(ctx)` when the user logs out**, and manage framework session expiration with `doors.SessionExpire(ctx, duration)`. Otherwise, you might leave private page instances active after authentication has ended.
* There is no need to check cookies/headers in event handlers, because they are already protected
* **If user access to certain actions or views can be revoked,** you should 
  * **Verify user view permissions during render** (not only in the ServePage handler) to ensure that the user can't access previously available views with dynamic navigation. 
  * **Verify user write permissions in the hook handler functions**, to ensure that even after permission is revoked after render, you are safe 


## 4. Use the closest `context.Context` value.

Context is used to track component relations, action progress, and many more.  Always use the closest context you have in scope. 

```templ
templ (f *fragment) Render() {
	@doors.Run(func(pageCtx context.Context) {
		itemId.Sub(pageCtx, func(ctx context.Context, id int) bool {
			✅  f.door.Update(ctx, card(id))
			// ❌ f.door.Update(pageCtx, card(id))
			return false
		})
	})
	@f.door
}

```

Most of the time, nothing serious will happen if you mess it up, 

## 5. Avoid storing database data in state.

In general, you don't need to store database query results in fields or `Beams`.

✅  Store the ID in the fragment field.

```templ
func newCard(id string) *card {
	return &card {
	  // store id
		id: id,
	}
}

type card struct {
	id string
}

templ (c *card) Render() {
  // retrieve db data
	{{ item := db.get(c.id) }}
	/* use item in render */
}
```

✅   Store ID in **Beam**

```templ
idBeam := doors.NewBeam(pathBeam, func(p Path) string {
  return p.id
})
```

❌ Store DB entry in fragment field like:

```templ
func newCard(id string) *card {
  // query
  item := db.get(c.id)
	return &card {
	  // store item
		item: item,
	}
}

type card struct {
	item db.Item
}

templ (c *card) Render() {
	/* use c.item in render */
}
```

❌  Store db entry in Beam like:

```templ
itemBeam := doors.NewBeam(pathBeam, func(p Path) db.Item {
  return db.get(p.id)
})
```

If you need data **only to produce render output** - fire and forget, so you won't waste server memory for nothing.  However, it's your decision.

## 6. Be conscious with front-end manipulations via JavaScript 

Parts of the DOM are controlled by the framework. Avoid removing or moving dynamic elements via JavaScript.

