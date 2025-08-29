# Important

Practices you should understand and follow. 

## 1. Use local context

In all `templ` components. handlers and listeners framework provides `context.Context` value. 

❌ Use `context.Background()` (or any external context) instead of `ctx` provided by the framework in framework related operations.  **It panics**.

❌ Try to enable interactivity/reactivity with *doors* in templ-based pages **not served** by *doors*. **It panics**.

✅ Update `Door`, read and mutate `Beam/SourceBeam` values, use `doors.A...` binds using context value in scope inside frameworks space.

✅ Use closest context in scope. *It is used to track component relations, action progress, and more.* 

```templ
templ (f *fragment) Render() {
	@doors.Run(func(pageCtx context.Context) {
		itemId.Sub(pageCtx, func(subCtx context.Context, id int) bool {
			✅  f.door.Update(subCtx, card(id))
			// ❌ f.door.Update(pageCtx, card(id))
			// ❌ f.door.Update(context.Background(), card(id))
			return false
		})
	})
	@f.door
}
```

> Context lifecycle is linked to dynamic components tree. You can use `ctx.Done()` channel in spawned goroutines to trigger clean up on DOM removal.

## 2. Respect blocking free context and data propagation.

Rendering operations and **Beam** handlers are executed in the framework's runtime. 

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
* **There is no need to check cookies/headers in the event handlers**, because they are already protected
* **If user access to certain actions or views can be revoked,** you should 
  * **Verify user view permissions during render** to ensure that the user can't access previously available views with dynamic navigation. 
  * **Verify user write permissions in the transactions** to ensure that even if the permission is revoked after rendering, you are still safe.

## 4. Avoid storing database data in state.

In general, you don't need to store database query results in fields or `Beams`.

✅  Store the ID in the **Fragment** field.

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

❌ Store DB entry in the **Fragment** field like:

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

❌  Store db entry in the **Beam** like:

```templ
itemBeam := doors.NewBeam(pathBeam, func(p Path) db.Item {
  return db.get(p.id)
})
```

If you need data **only to produce render output** - render and forget, so you won't waste server memory for nothing.  However, it's your decision.

## 5. Be conscious with front-end manipulations via JavaScript 

Parts of the DOM are controlled by the framework. Avoid removing or moving dynamic elements via JavaScript.

