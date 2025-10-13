# Form & Authentication

Let's make our dashboard protected.

## 1. Database

Basic session storage with sqlite:

`./driver/sessions.db`

```go
package driver

import (
	"database/sql"
	"github.com/doors-dev/doors"
	"time"
)

func newSessionsDb(db *sql.DB) *SessionsDb {
	initQuery := `
		CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			login TEXT NOT NULL,
			expire DATETIME NOT NULL
		);
	`
	if _, err := db.Exec(initQuery); err != nil {
		panic("Failed to create sessions table: " + err.Error())
	}
	s := &SessionsDb{
		db: db,
	}
	go s.cleanup()
	return s
}

type Session struct {
	Token  string    `json:"token"`
	Login  string    `json:"login"`
	Expire time.Time `json:"expire"`
}

type SessionsDb struct {
	db *sql.DB
}

func (d *SessionsDb) cleanup() {
	for {
		<-time.After(10 * time.Minute)
		_, err := d.db.Exec("DELETE FROM sessions WHERE expire <= ?", time.Now())
		if err != nil {
			panic("Failed to cleanup expired sessions: " + err.Error())
		}
	}
}

func (d *SessionsDb) Add(login string, dur time.Duration) Session {
	token := doors.RandId()
	expire := time.Now().Add(dur)

	_, err := d.db.Exec(
		"INSERT INTO sessions (token, login, expire) VALUES (?, ?, ?)",
		token, login, expire,
	)
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}

	return Session{
		Token:  token,
		Login:  login,
		Expire: expire,
	}
}

func (d *SessionsDb) Get(token string) (Session, bool) {
	var session Session
	err := d.db.QueryRow(
		"SELECT token, login, expire FROM sessions WHERE token = ? AND expire > ?",
		token, time.Now(),
	).Scan(&session.Token, &session.Login, &session.Expire)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty session and false if not found or expired
			return Session{}, false
		}
		panic("Failed to get session: " + err.Error())
	}

	return session, true
}

func (d *SessionsDb) Remove(token string) bool {
	result, err := d.db.Exec("DELETE FROM sessions WHERE token = ?", token)
	if err != nil {
		panic("Failed to remove session: " + err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		panic("Failed to get rows affected: " + err.Error())
	}

	return rowsAffected > 0
}
```

Initialization:

`./driver/driver.go`

```go
/* ... */
var Sessions *SessionsDb

func init() {
	/* ... */
	sessions, err := sql.Open("sqlite3", "./sqlite/sessions.sqlite3")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	Sessions = newSessionsDb(sessions)
}

```

## 1. Login Fragment

`./login.templ`

```templ
package main

import (
	"context"
	"github.com/derstruct/doors-dashboard/driver"
	"github.com/doors-dev/doors"
	"net/http"
	"time"
)

func login() templ.Component {
	return doors.F(&loginFragment{
		// create dynamic attribute "class" with value "hide"
        // to use on the login error message
		messageClass: doors.NewADyn("class", "hide", true),
	})
}

// login credentials ("tutorial style")
const userLogin = "admin"
const userPassword = "password123"
const sessionDuration = time.Hour * 24

type loginFragment struct {
	messageClass doors.ADyn
}

// type to decode the form data
type loginData struct {
	Login    string `form:"login"`
	Password string `form:"password"`
}

func (f *loginFragment) submit(ctx context.Context, r doors.RForm[loginData]) bool {
	// check user credentials
	if r.Data().Login != userLogin || r.Data().Password != userPassword {
		// unset dynamic class attribute, show error message
		f.messageClass.Enable(ctx, false)
		return false
	}
	// add session
	session := driver.Sessions.Add(r.Data().Login, sessionDuration)
	// set cookie
	r.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    session.Token,
		Expires:  time.Now().Add(sessionDuration),
		Path:     "/",
		HttpOnly: true,
	})
	// perform actions after the request on the client side
	r.After([]doors.Action{
		// reload the page
		doors.ActionLocationReload{},
		// pro tip: keep showing indication while the page is reloading
		doors.ActionIndicate{
			Duration:  10 * time.Second,
			Indicator: doors.IndicatorOnlyAttrQuery("#login-submit", "aria-busy", "true"),
		},
	})
	doors.SessionExpire(ctx, sessionDuration)
	return true
}

templ (f *loginFragment) Render() {
	<h1>Log In</h1>
    // form submit handler 
	@doors.ASubmit[loginData]{
		// to prevent repeated submission
		Scope: doors.ScopeOnlyBlocking(),
		// indicate on the button
		Indicator: doors.IndicatorOnlyAttrQuery("#login-submit", "aria-busy", "true"),
		// handle
		On: f.submit,
	}
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
			@f.errorMessage()
		</fieldset>
		<button id="login-submit" role="submit">Log In</button>
	</form>
}

// error message component
templ (l *loginFragment) errorMessage() {
	@doors.Style() {
		<style>
            .hide {
                display: none
            }
        </style>
	}
	// attach dynamic attribute
	@l.messageClass
	<p><mark>wrong password or login</mark></p>
}

```

> `doors.SessionExpire` is a safe precaution to ensure that opened pages with access to authorized functionality will not outlive the authorization session. 

## 2. Read the Session in the Model Handler

`./app.templ`

```templ
// read cookies and obtain the session from the database
func getSession(r doors.R) *driver.Session {
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

func Handler(m doors.doors.ModelRouter[Path], r doors.RModel[Path]) doors.ModelRoute {
	return p.App(&app{
		 session: getSession(r),
	})
}
```

## 3. Show the Login Form 

```templ
templ (a *app) Body() {
	if a.session == nil {
		@login()
	} else {
		@doors.Sub(a.id, func(id int) templ.Component {
			if id == -1 {
				return locationSelector(func(ctx context.Context, city driver.Place) {
					a.path.Mutate(ctx, func(p Path) Path {
						p.Selector = false
						p.Dashboard = true
						p.Id = city.Id
						return p
					})
				})
			}
			// render dashboard component
			return dashboard(id, a.path)
		})
	}
}
```

## 4. Adapt the Title

```templ
templ (a *app) Head() {
	if a.session == nil {
		<title>Login</title>
	} else {
		@doors.Head(a.id, func(id int) doors.HeadData {
			if id == -1 {
				return doors.HeadData{
					Title: "Select Location",
				}
			}
		
			city, _ := driver.Cities.Get(id)
			if city.Name == "" {
				return doors.HeadData{
					Title: "Location Not Found",
				}
			}
		
			return doors.HeadData{
				Title: "Weather in " + city.Name + ", " + city.Country.Name,
			}
		})
	}
}

```

## 5. Log Out Button

```templ
templ (a *app) logout() {
	<section>
		@doors.AClick{
			On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
				// remove the coolie
				r.SetCookie(&http.Cookie{
					Name:   "session",
					Path:   "/",
					MaxAge: -1,
				})
				// remove the db entry
				driver.Sessions.Remove(hp.session.Token)
				// end internal framework session
				// to end all authiorized instances (pages)
				doors.SessionEnd(ctx)
				return true
			},
		}
		<button class="secondary">Log Out</button>
	</section>
}
```

**Result:**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/mi2dr7npcqkzoyipbv83.gif)

---

##  Code

`./app.templ`

```templ
package main

import (
	"context"
	"github.com/derstruct/doors-dashboard/driver"
	"github.com/doors-dev/doors"
	"net/http"
)

type Path struct {
	Selector  bool `path:"/"`
	Dashboard bool `path:"/:Id"`
	Id        int
	Units     *driver.Units `query:"units"`
	Days      *int          `query:"days"`
}

// read cookies and obtain the session from the database
func getSession(r doors.R) *driver.Session {
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

func Handler(m doors.doors.ModelRouter[Path], r doors.RModel[Path]) doors.ModelRoute {
	return p.App(&app{
		 session: getSession(r),
	})
}

type app struct {
	// add the session property
	session *driver.Session
	path    doors.SourceBeam[Path]
	id      doors.Beam[int]
}

templ (a *app) Body() {
	if a.session == nil {
		@login()
	} else {
		@doors.Sub(a.id, func(id int) templ.Component {
			if id == -1 {
				return locationSelector(func(ctx context.Context, city driver.Place) {
					a.path.Mutate(ctx, func(p Path) Path {
						p.Selector = false
						p.Dashboard = true
						p.Id = city.Id
						return p
					})
				})
			}
			// render dashboard component
			return dashboard(id, a.path)
		})
	}
}

templ (a *app) Head() {
	if a.session == nil {
		<title>Login</title>
	} else {
		@doors.Head(a.id, func(id int) doors.HeadData {
			if id == -1 {
				return doors.HeadData{
					Title: "Select Location",
				}
			}
		
			city, _ := driver.Cities.Get(id)
			if city.Name == "" {
				return doors.HeadData{
					Title: "Location Not Found",
				}
			}
		
			return doors.HeadData{
				Title: "Weather in " + city.Name + ", " + city.Country.Name,
			}
		})
	}
}

templ (a *app) Body() {
	if a.session == nil {
		@login()
	} else {
		@doors.Sub(a.id, func(id int) templ.Component {
			if id == -1 {
				return locationSelector(func(ctx context.Context, city driver.Place) {
					a.path.Mutate(ctx, func(p Path) Path {
						p.Selector = false
						p.Dashboard = true
						p.Id = city.Id
						return p
					})
				})
			}
			// render dashboard component
			return dashboard(id, a.path)
		})
	}
}

templ (a *app) logout() {
	<section>
		@doors.AClick{
			On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
				// remove the coolie
				r.SetCookie(&http.Cookie{
					Name:   "session",
					Path:   "/",
					MaxAge: -1,
				})
				// remove the db entry
				driver.Sessions.Remove(hp.session.Token)
				// end internal framework session
				// to end all authiorized instances (pages)
				doors.SessionEnd(ctx)
				return true
			},
		}
		<button class="secondary">Log Out</button>
	</section>
}
```

`./login.templ`

```templ
package main

import (
	"context"
	"github.com/derstruct/doors-dashboard/driver"
	"github.com/doors-dev/doors"
	"net/http"
	"time"
)

func login() templ.Component {
	return doors.F(&loginFragment{
		// create dynamic attribute "class" with value "hide"
		// to use on the login error message
		messageClass: doors.NewADyn("class", "hide", true),
	})
}

// login credentials ("tutorial style")
const userLogin = "admin"
const userPassword = "password123"
const sessionDuration = time.Hour * 24

type loginFragment struct {
	messageClass doors.ADyn
}

// type to decode the form data
type loginData struct {
	Login    string `form:"login"`
	Password string `form:"password"`
}

func (f *loginFragment) submit(ctx context.Context, r doors.RForm[loginData]) bool {
	// check user credentials
	if r.Data().Login != userLogin || r.Data().Password != userPassword {
		// unset dynamic class attribute, show error message
		f.messageClass.Enable(ctx, false)
		return false
	}
	// add session
	session := driver.Sessions.Add(r.Data().Login, sessionDuration)
	// set cookie
	r.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    session.Token,
		Expires:  time.Now().Add(sessionDuration),
		Path:     "/",
		HttpOnly: true,
	})
	// perform actions after the request on the client side
	r.After([]doors.Action{
		// reload the page
		doors.ActionLocationReload{},
		// pro tip: keep showing indication while the page is reloading
		doors.ActionIndicate{
			Duration:  10 * time.Second,
			Indicator: doors.IndicatorOnlyAttrQuery("#login-submit", "aria-busy", "true"),
		},
	})
	// set doors session expiration (to not outlive authorization)
	doors.SessionExpire(ctx, sessionDuration)
	return true
}

templ (f *loginFragment) Render() {
	<h1>Log In</h1>
	// form submit handler 
	@doors.ASubmit[loginData]{
		// to prevent repeated submission
		Scope: doors.ScopeOnlyBlocking(),
		// indicate on the button
		Indicator: doors.IndicatorOnlyAttrQuery("#login-submit", "aria-busy", "true"),
		// handle
		On: f.submit,
	}
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
			@f.errorMessage()
		</fieldset>
		<button id="login-submit" role="submit">Log In</button>
	</form>
}

// error message component
templ (l *loginFragment) errorMessage() {
	@doors.Style() {
		<style>
            .hide {
                display: none
            }
        </style>
	}
	// attach dynamic attribute
	@l.messageClass
	<p><mark>wrong password or login</mark></p>
}
```

