# Dynamic Navigation

## 1. Query Params

In the weather API, besides the city, we also have two variables: units (metric/imperial) and forecast days.

Add it to our path model:

```templ
type Path struct {
	Selector  bool `path:"/"`
	Dashboard bool `path:"/:Id"`
	Id        int
	Units     *driver.Units `query:"units"`
	Days      *int          `query:"days"`
}
```

Notice I used reference types. Otherwise, query parameters get a zero value and always appear in the URI.

> Decoding and encoding of query parameters are provided by the [go-playground/form v4](https://github.com/go-playground/form) library.  So refer to its documentation for all features.

## 2. Dashboard Fragment

`./dashboard.templ`

To keep the app simple, let's move the dashboard to a separate fragment.

```templ
func dashboard(id int) templ.Component {
	return doors.F(&dashboardFragment{
		id: id,
	})
}

type dashboardFragment struct {
	id int
}

templ (f *dashboardFragment) Render() {
	<article>
		{{ city, _ := driver.Cities.Get(f.id) }}
		if city.Name == "" {
			@doors.Status(404)
			<h1>Location Not Found</h1>
		} else {
			<h1>{ city.Name }, { city.Country.Name }</h1>
		}
	</article>
}
```

The dashboard depends on the location ID (provided by the app already) and the days and units query parameters:

We derive those from the path:

```templ
func dashboard(id int, path doors.Beam[Path]) templ.Component {
	settings := doors.NewBeam(path, func(p Path) dashboardSettings {
		s := dashboardSettings{}

		// default unit value and some validation
		if p.Units == nil || *p.Units != driver.Imperial {
			s.units = driver.Metric
		} else {
			s.units = driver.Imperial
		}

		// default days value and some validation
		if p.Days == nil || *p.Days <= 1 {
			s.days = 1
		} else {
			s.days = min(*p.Days, 7)
		}

		return s
	})

	return doors.F(&dashboardFragment{
		id:       id,
		settings: settings,
	})
}

type dashboardFragment struct {
	id       int
	settings doors.Beam[dashboardSettings]
}
```

Render the dashboard on the page:

`./app.templ`

```templ
templ (a *app) Body() {
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
		// render the dashboard component
		return dashboard(id, hp.path)
	})
}
```

## 3. Menu

`./dashboard.templ`

### Change City

For the location selector to appear, we need to render a link to `/`. It would also be nice if query parameters persisted, so we generate the link based on the settings beam:

```templ
templ (f *dashboardFragment) Render() {
	<article>
		/* h1 */
		// render the menu
		@f.menu()
	</article>
}

templ (f *dashboardFragment) menu() {
	<nav>
		// subscribe to the settings beam
		@doors.Sub(f.settings, func(s dashboardSettings) templ.Component {
			return f.changeLocation(s)
		})
	</nav>
}

templ (f *dashboardFragment) changeLocation(s dashboardSettings) {
	// generate href attribute and attach click listener
	// for live update
	@doors.AHref{
		// target path model
		Model: Path{
			Selector: true,
			Units:    &s.units,
			Days:     &s.days,
		},
	}
	<a>Change Location</a>
}

```

> `AHref` also supports the **Scopes** and **Indication** APIs

**Switch to the location selector via dynamic link:**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/3lk0hxej1q53ans2h2qg.gif)

After clicking, the query parameters appear with default values. It’s okay, but not ideal.

Provide `nil` for the defaults so the behavior is consistent:

```templ
templ (f *dashboardFragment) changeLocation(s dashboardSettings) {
	// evaluates the provided function during render
	@doors.E(func(ctx context.Context) templ.Component {
		m := Path{
			Selector: true,
		}
	
		// keeps nil for default values
		if s.units != driver.Metric {
			m.Units = &s.units
		}
		if s.days != 1 {
			m.Days = &s.days
		}
	
		return doors.AHref{
			Model: m,
		}
	})
	<a>Change Location</a>
}
```

**Result:**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/kg15lmf5psmrwsjvqo1i.gif)

### Switch Units

Render a link to switch units back and forth:

```templ
templ (f *dashboardFragment) switchUnits(s dashboardSettings) {
    @doors.E(func(ctx context.Context) templ.Component {
        m := Path{
            Dashboard: true,
            Id:        f.id,
        }

        // maintain the days value
        if s.days != 1 {
            m.Days = &s.days
        }

        // switch units
        if s.units == driver.Metric {
            m.Units = driver.Imperial.Ref()
        } else {
            // metric is default
            m.Units = nil
        }

        return doors.AHref{
            Model: m,
        }
    })
    <a class="contrast">
        &#8644; { s.units.Label() }
    </a>
}
```

Add some styles and render the units switcher:

```templ

templ (f *dashboardFragment) menu() {
	// minifies css and converts inline styles to a cacheable <link rel="stylesheet"...>
	@doors.Style() {
		<style>
            nav.dashboard {
                display: flex;
                flex-direction: row;
                justify-content: start;
                gap: var(--pico-spacing);
                white-space: nowrap;
                flex-wrap: wrap;
            }
        </style>
	}
	<nav class="dashboard">
		@doors.Sub(f.settings, func(s dashboardSettings) templ.Component {
			return f.changeLocation(s)
		})
		@doors.Sub(f.settings, func(s dashboardSettings) templ.Component {
			return f.switchUnits(s)
		})
	</nav>
}

```

**Query param switching:**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/q7nds3ntgtd5j3b4yiq8.gif)

### Forecast Days

The forecast-days links must preserve the units query value. To avoid unnecessary updates, derive a beam for the units:

```templ
func dashboard(id int, path doors.Beam[Path]) templ.Component {
	settings := doors.NewBeam(path, func(p Path) dashboardSettings {
		s := dashboardSettings{}
		/* ... */
		return s
	})

	// derive units
	units := doors.NewBeam(settings, func(s dashboardSettings) driver.Units {
		return s.units
	})

	return doors.F(&dashboardFragment{
		id:       id,
		settings: settings,
		units:    units,
	})
}

type dashboardFragment struct {
	id       int
	settings doors.Beam[dashboardSettings]
	units    doors.Beam[driver.Units]
}
```

Subscribe the menu to it:

```templ
templ (f *dashboardFragment) menu() {
    /* ... */
    <nav class="dashboard">
        @doors.Sub(f.settings, func(s dashboardSettings) templ.Component {
            return f.changeLocation(s)
        })

        // subscribe days menu to the units beam
        @doors.Sub(f.units, func(u driver.Units) templ.Component {
            return f.changeDays(u)
        })

        @doors.Sub(f.settings, func(s dashboardSettings) templ.Component {
            return f.switchUnits(s)
        })
    </nav>
}
```

And maintain the units query value in the days menu:

```templ
templ (f *dashboardFragment) changeDays(u driver.Units) {
	for i := range 7 {
		{{ days := i + 1 }}
		@doors.E(func(ctx context.Context) templ.Component {
			m := Path{
				Dashboard: true,
				Id:        f.id,
			}
		
			// if not default
			if days != 1 {
				m.Days = &days
			}
			// if not default
			if u != driver.Metric {
				m.Units = driver.Imperial.Ref()
			}
		
			return doors.AHref{
				Model: m,
			}
		})
		<a class="secondary">
			if i == 0 {
				1 day
			} else {
				{ fmt.Sprintf("%d days", days) }
			}
		</a>
	}
}
```

**Reactive menu:**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/i24pt88zau1h3mjvllgq.gif)

> _doors_ doesn’t keep the whole DOM in memory. With beam derivation, you explicitly tie a specific HTML section to a specific piece of data. **Diff data, not DOM.**

## 4. Active Link Highlighting

The client can automatically apply active-link highlighting if you configure it in `doors.AHref`:

```templ
templ (f *dashboardFragment) changeDays(u driver.Units) {
	for i := range 7 {
		{{ days := i + 1 }}
		@doors.E(func(ctx context.Context) templ.Component {
			m := Path{
				Dashboard: true,
				Id:        f.id,
			}
			/* ... */
			return doors.AHref{
				Model: m,
				// active link highlighting configuration
				Active: doors.Active{
					// use Indication API to set the "aria-current" attribute
					Indicator: doors.IndicatorOnlyAttr("aria-current", "page"),
				},
			}
		})
		<a class="secondary">
			if i == 0 {
				1 day
			} else {
				{ fmt.Sprintf("%d days", days) }
			}
		</a>
	}
}
```

> By default, it checks the whole path and all query values to apply the indication, but you can configure a narrower matching strategy.

**Active link is highlighted:**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/ommxxt8egdwxbcnrwc8a.gif)

Next:  [Charts](./09-charts.md)
---

## Code

```templ
package main

import (
	"context"
	"fmt"
	"github.com/derstruct/doors-dashboard/driver"
	"github.com/doors-dev/doors"
)

type dashboardSettings struct {
	units driver.Units
	days  int
}

func dashboard(id int, path doors.Beam[Path]) templ.Component {
	settings := doors.NewBeam(path, func(p Path) dashboardSettings {
		s := dashboardSettings{}
		// default unit value and some validation
		if p.Units == nil || *p.Units != driver.Imperial {
			s.units = driver.Metric
		} else {
			s.units = driver.Imperial
		}
		// default days value and some validation
		if p.Days == nil || *p.Days <= 1 {
			s.days = 1
		} else {
			s.days = min(*p.Days, 7)
		}
		return s
	})

	// derive units
	units := doors.NewBeam(settings, func(s dashboardSettings) driver.Units {
		return s.units
	})

	return doors.F(&dashboardFragment{
		id:       id,
		settings: settings,
		units:    units,
	})
}

type dashboardFragment struct {
	id       int
	settings doors.Beam[dashboardSettings]
	units    doors.Beam[driver.Units]
}

templ (f *dashboardFragment) Render() {
	<article>
		{{ city, _ := driver.Cities.Get(f.id) }}
		if city.Name == "" {
			@doors.Status(404)
			<h1>Location Not Found</h1>
		} else {
			<h1>{ city.Name }, { city.Country.Name }</h1>
		}
		@f.menu()
	</article>
}

templ (f *dashboardFragment) menu() {
	// minifies css and converts inline styles to a cacheable <link rel="stylesheet"...>
	@doors.Style() {
		<style>
            nav.dashboard {
                display: flex;
                flex-direction: row;
                justify-content: start;
                gap: var(--pico-spacing);
                white-space: nowrap;
                flex-wrap: wrap;
            }
        </style>
	}
	<nav class="dashboard">
		@doors.Sub(f.settings, func(s dashboardSettings) templ.Component {
			return f.changeLocation(s)
		})
		@doors.Sub(f.units, func(u driver.Units) templ.Component {
			return f.changeDays(u)
		})
		@doors.Sub(f.settings, func(s dashboardSettings) templ.Component {
			return f.switchUnits(s)
		})
	</nav>
}

templ (f *dashboardFragment) changeLocation(s dashboardSettings) {
	// evaluates the provided function during render
	@doors.E(func(ctx context.Context) templ.Component {
		m := Path{
			Selector: true,
		}
	
		// keeps nil for default values
		if s.units != driver.Metric {
			m.Units = &s.units
		}
		if s.days != 1 {
			m.Days = &s.days
		}
	
		return doors.AHref{
			Model: m,
		}
	})
	<a>Change Location</a>
}

templ (f *dashboardFragment) changeDays(u driver.Units) {
	for i := range 7 {
		{{ days := i + 1 }}
		@doors.E(func(ctx context.Context) templ.Component {
			m := Path{
				Dashboard: true,
				Id:        f.id,
			}
		
			// if not default
			if days != 1 {
				m.Days = &days
			}
			// if not default
			if u != driver.Metric {
				m.Units = driver.Imperial.Ref()
			}
		
			return doors.AHref{
				Model: m,
				Active: doors.Active{
					// use Indication API to set the "aria-current" attribute
					Indicator: doors.IndicatorOnlyAttr("aria-current", "page"),
				},
			}
		})
		<a class="secondary">
			if i == 0 {
				1 day
			} else {
				{ fmt.Sprintf("%d days", days) }
			}
		</a>
	}
}

templ (f *dashboardFragment) switchUnits(s dashboardSettings) {
	@doors.E(func(ctx context.Context) templ.Component {
		m := Path{
			Dashboard: true,
			Id:        f.id,
		}
	
		// maintain the days value
		if s.days != 1 {
			m.Days = &s.days
		}
	
		// switch units
		if s.units == driver.Metric {
			m.Units = driver.Imperial.Ref()
		} else {
			// metric is default
			m.Units = nil
		}
	
		return doors.AHref{
			Model: m,
		}
	})
	<a class="contrast">
		&#8644; { s.units.Label() }
	</a>
}
```

