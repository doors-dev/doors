# Form & Authentification

We cannot simply grant anyone the ability to populate our catalog. Let's add a login form to the home page.

## 1. Form Fragment

`./home/login.templ`

```templ
package home

type loginFragment struct {
}

templ (l *loginFragment) Render() {
	<form>
		<fieldset>
			<label>
				Login
				<input
					name="login"
					autocomplete="login"
				/>
			</label>
			<label>
				Password
				<input
					type="password"
					name="password"
					autocomplete="password"
				/>
			</label>
		</fieldset>
		<button role="submit">Log In</button>
	</form>
}

```

`./home/page.teml` 

```templ
templ (h *homePage) Body() {
	// render fragment
	@doors.F(&loginFragment{})
}

```

Okay, now we can see the form on the home page, but it doesn't do anything. 

## 2. Submit Event Handler

### Prepare Form Data Structure 

`./home/login.templ`

```templ
type loginData struct {
	Login    string `form:"login"`
	Password string `form:"password"`
}
```

> It uses https://github.com/go-playground/form under the hood. 

### Event Handler Attribute

`./home/login.teml`

```templ

templ (l *loginFragment) Render() {
	// attach submit attribute with loginData as form data
	@doors.ASubmit[loginData]{
		On: func(ctx context.Context, r doors.RForm[loginData]) bool {
			// debug print
			fmt.Printf("%+v\n", r.Data())

			// imitation of something happening
			<-time.After(time.Second)

			// not done, keep active
			return false
		},
	}
	<form>
		<fieldset>
			<label>
				Login
				<input
					name="login"
					autocomplete="login"
				/>
			</label>
			<label>
				Password
				<input
					type="password"
					name="password"
					autocomplete="password"
				/>
			</label>
			@l.messageDoor
		</fieldset>
		<button role="submit">Log In</button>
	</form>
}
```

We added a one-second delay in the form handler function to examine the *doors*' concurrency control. 

First, you can notice that regardless of how frequently you submit the form, events are queued and processed sequentially — that's intentional concurrency protection at **a single hook** level. 

> **Inside the hook function, you can be sure that it could be invoked next time only after the current execution completes. **
>
> That does not affect other hooks and operations, so you need to use synchronization techniques when accessing shared data.

But we don't want the user to submit a new form when the previous one is still processing. 

That's what **scopes** are here for. 

#### Scope

```templ
func (l *loginFragment) submit() doors.Attr {
	return doors.ASubmit[loginData]{
		// blocks new submission, until the previous one is processed
		Scope: doors.ScopeBlocking(),
		On: func(ctx context.Context, r doors.RForm[loginData]) bool {
			fmt.Printf("%+v\n", r.Data())
			<-time.After(time.Second)
			return false
		},
	}
}
```

> There are multiple types of scopes; additionally, they can be shared between hooks and combined in a pipeline. `doors.Scope{Type}` is a helper for creating a scope pipeline of one scope of a specific type.

#### Indication

Let's tell the user that something is happening during our form processing. [PicoCSS](https://picocss.com/docs/loading) provides a special attribute we can use on a button to display a loading state.

##### Specify Id for the submit button

```templ
<button id="login-submit" role="submit">Log In</button>
```

##### Set up hook pending indication

```templ
func (l *loginFragment) submit() doors.Attr {
	return doors.ASubmit[loginData]{
		// query element with id login-submit and set attr area-busy to true during hook execution
		Indicator: doors.IndicatorAttrQuery("#login-submit", "aria-busy", "true"),
		Scope:     doors.ScopeBlocking(),
		On: func(ctx context.Context, r doors.RForm[loginData]) bool {
			fmt.Printf("%+v\n", r.Data())
			<-time.After(time.Second)
			return false
		},
	}
}

```

> `doors.Indicator{Type}{Selector}` is a helper function to define a single indicator of a specific type and selector.  There are multiple indication types  (attribute, class, content) and selectors (target, query, parent query).  
>
> You can specify multiple indicators:
>
> ```go
> func (l *loginFragment) submit() doors.Attr {
> 	return doors.ASubmit[loginData]{
>       // advanced indicator control
> 		Indicator: []doors.Indicator{
> 			doors.AttrIndicator{
> 				Selector: doors.SelectorQuery("#login-submit"),
> 				Name:     "aria-busy",
> 				Value:    "true",
> 			},
> 			doors.ContentIndicator{
> 				Selector: doors.SelectorQuery("#login-submit"),
> 				Content:  "Wait...",
> 			},
> 		},
>     /* ... */
>   }
> }
> ```
>
> 

## 3. Form Error Message

`./home/login.teml`

```templ
package home

import (
	"context"
	"github.com/doors-dev/doors"
	"time"
)

// hardcoded login data, only for tutorial purposes!
const userLogin = "admin"
const userPassword = "password123"


type loginFragment struct {
  // dynamic door to display message
	messageDoor doors.Door
}


templ (l *loginFragment) Render() {
  // attach submit
    @doors.ASubmit[loginData]{
		Indicator: doors.IndicatorAttrQuery("#login-submit", "aria-busy", "true"),
		Scope:     doors.ScopeBlocking(),
		On: func(ctx context.Context, r doors.RForm[loginData]) bool {
            if r.Data().Login != userLogin || r.Data().Password != userPassword {
              //display errror
              l.messageDoor.Update(ctx, l.errorMessage())
              return false
            } 
            // display ok, just for testing
            l.messageDoor.Update(ctx, doors.Text("ok"))
            return false
		},
	}
	<form>
		<fieldset>
			<label>
				Login
				<input
					name="login"
					autocomplete="login"
				/>
			</label>
			<label>
				Password
				<input
					type="password"
					name="password"
					autocomplete="password"
				/>
			</label>
			@l.messageDoor
		</fieldset>
		<button id="login-submit" role="submit">Log In</button>
	</form>
}

// error message template
templ (l *loginFragment) errorMessage() {
	<p><mark>wrong password or login</mark></p>
}

type loginData struct {
	Login    string `form:"login"`
	Password string `form:"password"`
}

```

> Now, the form should tell if the login and password are correct or not.

## 4.  Session Cookies 

### Set Cookie

In the form handler function

`./home/login.templ`

```templ
const sessionDuration = time.Hour * 24

templ (l *loginFragment) Render() {
	@doors.ASubmit[loginData]{
		Indicator: doors.IndicatorAttrQuery("#login-submit", "aria-busy", "true"),
		Scope:     doors.ScopeBlocking(),
		On: func(ctx context.Context, r doors.RForm[loginData]) bool {
			if r.Data().Login != userLogin || r.Data().Password != userPassword {
				l.messageDoor.Update(ctx, l.errorMessage())
				return false
			}
			session := driver.Sessions.Add(r.Data().Login, sessionDuration)
			r.SetCookie(&http.Cookie{
				Name:     "session",
				Value:    session.Token,
				Expires:  time.Now().Add(sessionDuration),
				Path:     "/",
				HttpOnly: true,
			})
			r.After(doors.AfterLocationReload())
			doors.SessionExpire(ctx, sessionDuration)
			return true
		},
	}
	<form>
		/* ... */
	</form>
}
```

> `r.After(doors.After)` allows you to specify an action to execute on the front-end after the hook request is finished. `doors.AfterLocationReload()` is useful for situations when you need to reinitialize the page after hook execution.
>
> `doors.SessionExpire` is a safe precaution to ensure that opened pages with access to authorized functionality will not outlive the authorization session. 

### Check Authorization

`./home/page.templ`

```templ
type homePage struct {
  // add session property
	session *driver.Session
}

templ (h *homePage) Body() {
    // display login form if there is no session
    if h.session == nil {
        @doors.F(&login{})
    } else {
        <h1>Welcome <strong>{ h.session.Login }</strong>!</h1>
    }
}
```



`./home/handler.go`

```templ
package home

import (
	"github.com/derstruct/doors-starter/driver"
	"github.com/doors-dev/doors"
)


func Handler(p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
	c, err := r.GetCookie("session")
	if err != nil {
		return p.Page(&homePage{})
	}
	// get session by cookie value
	s, found := driver.Sessions.Get(c.Value)
	if !found {
		return p.Page(&homePage{})
	}
	// if there is a session, set it to home page properties
	return p.Page(&homePage{
		session: &s,
	})
}
```

> The handler function is the page's entry point. **It's the only place where you must care about authorization**.  

Let's refactor the session extraction to an external function, so we can reuse it on **catalog** page.

`./common/utils.go`

```templ
package common

import (
	"github.com/derstruct/doors-starter/driver"
	"github.com/doors-dev/doors"
)

// doors.R - base request interface
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

`./home/handler.go`

```templ
package home

import (
	"github.com/derstruct/doors-starter/common"
	"github.com/doors-dev/doors"
)

func Handler(p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
	return p.Page(&homePage{
		session: common.GetSession(r),
	})
}
```



### 5. Logout

Add a logout button and handler to the home page for simplicity.

`./home/page.templ`

```templ
templ (h *homePage) Body() {
	if h.session == nil {
		@doors.F(&login{})
	} else {
		<h1>Welcome <strong>{ h.session.Login }</strong>!</h1>
		// attach click attribute with logout handler
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
		<button>Log Out</button>
	}
}
```

> **It's very important to call SessionEnd(ctx) on logout to ensure that no pages are left running under the authorized user.** 

