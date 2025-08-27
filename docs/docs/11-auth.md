# Authentification

For protected resources, you need to control authentication & authorization in the page handler.

## Access Control

**Example:**

```templ

router.Use(doors.ServePage(
  // page handler function
  func (p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
  	c, err := r.GetCookie("session")
    if err != nil {
      // show unauthorized page in detached (without path sync) mode 
    	return p.Reroute(UnauthorizedPath{}, true)
    }
    session, ok := db.Get(c.Value)
    if !ok  {
    	return p.Reroute(UnauthorizedPath{}, true)
    }
  	return p.Page(&Admin{})
	},
))
```

`doors.RPage` gives you access to cookies, headers, and the requested **path model**.

There is no need to check cookies/headers in event handlers, because they are already scoped to the session and page instance.

> **❗ Don't rely only on the cookie value; always implement session storage** 

### ⚠️ When to check authorization besides the page handler

If user access to certain actions or views can be revoked, you should 

* **Verify user view permissions during render** (not only in the ServePage handler) to ensure that the user can't access previously available views with dynamic navigation. 
* **Verify user write permissions in the hook handler functions** to ensure that even if the permission is revoked after rendering, you are still safe.

## Session Management

The framework has internal session mechanics to serve pages, content, and handle events. By default, an internal session lasts until at least one page instance remains alive. 

**However, when you implement your own authentication, ensure that the framework's session does not outlive yours.** Otherwise, you may leave the users in a situation where they are logged out, but opened tabs still have access to the private space.

### 1. Control expiration in login handler

```templ
templ (l *loginFragment) Render() {
  // login form handler
	@doors.ASubmit[loginData]{
	  /* setup */
		On: func(ctx context.Context, r doors.RForm[loginData]) bool {
		  /*  check login data */
		  
		  sessionDuration := time.Hour * 24
			session := db.CreateSession(r.Data, sessionDuration)
			r.SetCookie(&http.Cookie{
				Name:     "session",
				Value:    session.Token,
				// set expiration
				Expires:  time.Now().Add(sessionDuration),
				Path:     "/",
				HttpOnly: true,
			})
	
			// ✅ set internal session expiration to not outlive cookies
			doors.SessionExpire(ctx, sessionDuration)
			
			// reload after request to initiate instance with private space
			r.After(doors.AfterLocationReload())
			return true
		},
	}
	<form>
		/* ... */
	</form>
}
```

### 2. End session on log-out

When the user logs out, destroy all active instances by ending the session.

```templ
templ logout() {
		@doors.AClick{
          On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
             // ✅ end doors session to ensure no active pages left
            defer doors.SessionEnd(ctx)
            
            // clean cookies
            r.SetCookie(&http.Cookie{
              Name:   "session",
              Path:   "/",
              MaxAge: -1,
            })
            // remove session entry
            db.Sessions.Remove(h.session.Token)
           
            return true
          },
        }
		<button>Log Out</button>

}
```

> ⚠️ Ending the session causes pages to reload. In theory, it could happen before the browser receives a response and clears the cookies. To be on the safe side, **rely on session storage (remove on logout, check in page handler)**. 

## After Actions

Use [After Actions](./ref/07-after-actions.md) in login handlers to "redirect" the user to the authorized page.



