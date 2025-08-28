# Form & Authentification

We cannot simply grant anyone the ability to populate our catalog. Let's add a login form to the home page.

## 1. Login Fragment

Prepare the login fragment with a form

`./home/login.templ`

```templ
package home

import "github.com/doors-dev/doors"

func login() templ.Component {
	return doors.F(&loginFragment{})
}

type loginFragment struct {}

templ (f *loginFragment) Render() {
	<h1>Log In</h1>
	<form>
		<fieldset>
			<label>
				Login
				<input
					name="login"
					required="true"
				/>
			</label>
			<label>
				Password
				<input
					type="password"
					name="password"
					required="true"
				/>
			</label>
		</fieldset>
		<button role="submit">Log In</button>
	</form>
}
```

`./home/page.teml` 

```templ
/* .. */
templ (h *homePage) Body() {
	@login()
}
/* .. */
```

Okay, now we can see the form on the home page, but it doesn't do anything. 

## 2. Submit Event Handler

### Prepare form data and handler 

`./home/login.templ`

```templ
type loginData struct {
	Login    string `form:"login"`
	Password string `form:"password"`
}

func (f *loginFragment) submit(ctx context.Context, r doors.RForm[loginData]) bool {
	// debug print
	fmt.Printf("%+v\n", r.Data())

	// imitation of something happening
	<-time.After(time.Second)

	// not done, keep active
	return false
}
```

> It uses https://github.com/go-playground/form under the hood. 

### Attach handler to form

`./home/login.templ`

```templ
templ (f *loginFragment) Render() {
	<h1>Log In</h1>
	// onsubmit 
	@doors.ASubmit[loginData]{
		On: f.submit,
	}
	<form>
		/* .. */
	</form>
}

```

We added a one-second delay in the form handler function to examine the *doors*' concurrency control. 

First, you can notice that regardless of how frequently you submit the form, events are queued and processed sequentially — that's intentional concurrency protection at **a single hook** level. 

> **Inside the hook function, you can be sure that it can be invoked next time only after the current execution completes. **
>
> That does not affect other hooks and operations, so you need to use synchronization techniques when accessing shared data.

But we don't want the user to submit a new form when the previous one is still processing. 

That's what **scopes** are here for. 

#### Scope

```templ
templ (f *loginFragment) Render() {
	/* ... */
	@doors.ASubmit[loginData]{
		Scope: doors.ScopeBlocking(),
		On:    f.submit,
	}
	<form>
		/* ... */
	</form>
}

```

> There are multiple types of scopes; additionally, they can be shared between hooks and combined in a pipeline. `doors.Scope{Type}` is a helper for creating a scope pipeline of one scope of a specific type. Please refer to [Scopes](../docs/ref/04-scopes.md) for details.

#### Indication

Let's tell the user that something is happening during our form processing. [PicoCSS](https://picocss.com/docs/loading) provides a special attribute we can use on a button to display a loading state.

##### Specify Id for the submit button

```templ
<button id="login-submit" role="submit">Log In</button>
```

##### Set up hook pending indication

```templ
templ (f *loginFragment) Render() {
	/* ... */
	@doors.ASubmit[loginData]{
		// query element with id login-submit and set attr area-busy to true during hook execution
		Indicator: doors.IndicatorAttrQuery("#login-submit", "aria-busy", "true"),
		// blocks new submission (on front-end), until the previous one is processed
		Scope: doors.ScopeBlocking(),
		On:    f.submit,
	}
	<form>
		/* ... */
		<button id="login-submit" role="submit">Log In</button>
	</form>
}

```

> `doors.Indicator{Type}{Selector}` is a helper function to define a single indicator of a specific type and selector.  There are multiple indication types  (attribute, class, content) and selectors (target, query, parent query).  You can also specify multiple indicators. Please refer to [Indication](../docs/ref/03-indication.md) for details.

## 3. Form Error Message

Just add a **door** for the error message.

`./home/login.teml`

```templ
/* ... */

type loginFragment struct {
	// door to display message
	message doors.Door
}

// error message template
templ (l *loginFragment) errorMessage() {
	<p><mark>wrong password or login</mark></p>
}

func (f *loginFragment) submit(ctx context.Context, r doors.RForm[loginData]) bool {
	if r.Data().Login != userLogin || r.Data().Password != userPassword {
		//display errror
		f.message.Update(ctx, f.errorMessage())
		return false
	}
	// display ok, just for testing
	f.message.Update(ctx, doors.Text("ok"))
	return false
}

templ (f *loginFragment) Render() {
	/* ... */
	<form>
		<fieldset>
			/* ... */
			// render th message door
			@f.message
		</fieldset>
		/* ... */
	</form>
}

```

> Now, the form should tell if the login and password are correct or not.

## 4.  Session Cookies 

### Set Cookie

In the form handler function

`./home/login.templ`

```templ
const sessionDuration = time.Hour * 24

func (f *loginFragment) submit(ctx context.Context, r doors.RForm[loginData]) bool {
	if r.Data().Login != userLogin || r.Data().Password != userPassword {
		f.message.Update(ctx, f.errorMessage())
		return false
	}
    // add session to the storage
	session := driver.Sessions.Add(r.Data().Login, sessionDuration)

    // set cookie in the form submit response
	r.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    session.Token,
		Expires:  time.Now().Add(sessionDuration),
		Path:     "/",
		HttpOnly: true,
	})

    // tell frontent to reload the page after the form submit request
	r.After(doors.AfterLocationReload())
    // limit doors internal session to cookie session duration
    // to ensure that pages won't outlive authenitifacation
	doors.SessionExpire(ctx, sessionDuration)
	return true
}
```

> `r.After(doors.After)` allows you to specify an action to execute on the front-end after the hook request is finished. `doors.AfterLocationReload()` is useful for situations when you need to reinitialize the page after hook execution.
>
> `doors.SessionExpire` is a safe precaution to ensure that opened pages with access to authorized functionality will not outlive the authorization session. 

## 5. Check Authorization 

`./home/page.templ`

```templ
type homePage struct {
	// add session property
	session *driver.Session
}

/* ... */

templ (h *homePage) Body() {
	// display login form if there is no session
	if h.session == nil {
		@login()
	} else {
		<h1>Welcome <strong>{ h.session.Login }</strong>!</h1>
	}
}

/* ... */

func Handler(p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
	// read cookie for the page request
	c, err := r.GetCookie("session")
	if err != nil {
		return p.Page(&homePage{})
	}
	// get session entry by cookie value
	s, found := driver.Sessions.Get(c.Value)
	if !found {
		return p.Page(&homePage{})
	}
	// provide session to page
	return p.Page(&homePage{
		session: &s,
	})
}

```

> The handler function is the page's entry point. **It's the only place where you need to worry about checking cookie authorization**.   

## 6. Logout

Add a logout button and handler to the home page for simplicity.

`./home/page.templ`

```templ
templ (h *homePage) Body() {
	if h.session == nil {
		@login()
	} else {
		<h1>Welcome <strong>{ h.session.Login }</strong>!</h1>
		// button to log out
		@doors.AClick{
			On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
				// clean cookies
				r.SetCookie(&http.Cookie{
					Name:   "session",
					Path:   "/",
					MaxAge: -1,
				})
				// remove session entry
				driver.Sessions.Remove(h.session.Token)
				// end doors session to ensure no active pages left
				doors.SessionEnd(ctx)
				return true
			},
		}
		<button class="secondary">Log Out</button>
	}
}
```

> **It's very important to call SessionEnd(ctx) on logout to ensure that no pages are left running under the authorized user.** 

## 7. Refactor 

Let's refactor the session extraction to an external function, so we can reuse it on the **catalog** page.

`./common/utils.go`

```templ
package common

import (
	"github.com/derstruct/doors-tutorial/driver"
	"github.com/doors-dev/doors"
)

// doors.R - base request interface to deal with cookies
func GetSession(r doors.R) *driver.Session {
	c, err := r.GetCookie("session")
	if err != nil {
		return nil
	}
	s, found := driver.Sessions.Get(c.Value)
	if !found {
		return nil
	}
	return &s
}

```

`./home/page.templ`

```templ
/* ... */
func Handler(p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
	return p.Page(&homePage{
		session: common.GetSession(r),
	})
}
```

## 8.  Add authorization to the catalog page

### Page Handler

`./catalog/page.templ`

```templ
/* ... */

type catalogPage struct {
    // session property
	session *driver.Session
	path    doors.SourceBeam[Path]
}

/* ... */

func Handler(p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
	return p.Page(&catalogPage{
		session: common.GetSession(r),
	})
}

```

### Instance Storage

Now, instead of propagating the session object as a property, let's utilize **instance context storage**.

Prepare utils:

`./common/utils.go`

```templ
/* ... */

type sessionKey struct{}

func StoreSession(ctx context.Context, session *driver.Session) {
  // save to the pages global "thread safe" storage
	doors.InstanceSave(ctx, sessionKey{}, session)
}

func LoadSession(ctx context.Context) *driver.Session {
	session, ok := doors.InstanceLoad(ctx, sessionKey{}).(*driver.Session)
	if !ok {
		return nil
	}
	return session
}

// helper just to check
func IsAuthorized(ctx context.Context) bool {
	return LoadSession(ctx) != nil
}


/* ... */
```

### Auth check in category fragment

```templ
templ (f *categoryFragment) content(catId string) {
	{{ cat, ok := driver.Cats.Get(catId) }}
	if ok {
		<hgroup>
			<h1>{ cat.Name }</h1>
			<p>{ cat.Desc } </p>
		</hgroup>
		// check directly on the context in scope
		if common.IsAuthorized(ctx) {
			<p>
				<button class="contrast">Add Item</button>
			</p>
		}
	} else {
		@doors.Status(http.StatusNotFound)
		<div>
			<mark>Not Found</mark>
		</div>
	}
}
```

> Check any category page, you should see button "Add Item" button if authorized



---

Next: [Create Item](./10-create-item.md)
