## Path and Title

## 1. Path Variants

`./page.templ`

Here’s the page code again:

```templ
package main

import "github.com/doors-dev/doors"

type Path struct {
	Selector  bool `path:"/"`    
	Dashboard bool `path:"/:Id"`
	Id        int  
}

func Handler(p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
	return p.Page(&page{})
}

type page struct{}

func (hp *page) Render(path doors.SourceBeam[Path]) templ.Component {
	return Template(hp)
}

templ (hp *page) Head() {
	<title>Dashboard App</title>
}

templ (hp *page) Body() {
	@locationSelector()
}

```

It always displays the location selector. However, we have two path variants: `/` and `/:Id`. The second is used when a location is selected.

Look closely at the render function argument: `path doors.SourceBeam[Path]`. To have multiple page variants based on the path, we just need to subscribe to the provided beam. 

But remember, we’ll also have query parameters, but we don’t want the whole page to rerender on each query change, so _derive_:

```templ
type page struct {
	path doors.SourceBeam[Path]
	id   doors.Beam[int]
}

func (hp *page) Render(path doors.SourceBeam[Path]) templ.Component {
	// store path beam
	hp.path = path
	// derive beam with id
	hp.id = doors.NewBeam(path, func(p Path) int {
		// if the dashboard variant is active
		if p.Dashboard {
			return p.Id
		}
		// means location is not selected
		return -1
	})
	return Template(hp)
}
```

And then subscribe:

```templ
templ (hp *page) Body() {
    @doors.Sub(hp.id, func(id int) templ.Component {
        // no location selected
        if id == -1 {
            return locationSelector()
        }
        // display selected location
        return hp.showLocation(id)
    })
}

templ (hp *page) showLocation(id int) {
    <article>
        {{ city, _ := driver.Cities.Get(id) }}
        if city.Name == "" {
            // set the response status code
            @doors.Status(404)
            <h1>Location Not Found</h1>
        } else {
            <h1>{ city.Name }, { city.Country.Name }</h1>
        }
    </article>
}
```

**Display the selected city or 404:**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/qqpbkk11fk662sm94mu8.gif)

## 2. Apply Location Selection

`./location_selector.templ`

The location selector must change the path. Add an apply dependency to the location selector:

```templ
// add func to apply the selection
func locationSelector(apply func(context.Context, driver.Place)) templ.Component {
	/* .... */
	return doors.F(&locationSelectorFragment{
		// save to the new property
		apply: apply,
		/* ... */
	})
}

type locationSelectorFragment struct {
	apply func(context.Context, driver.Place)
	/* .... */
}
```

Implement submit functionality:

```templ
templ (f *locationSelectorFragment) submit(p driver.Place) {
	// nothing selected
	if p.Name == "" {
		<button disabled>Confirm Location</button>
	} else {
		@doors.AClick{
			// block repeated clicks
			Scope: doors.ScopeOnlyBlocking(),
			// indicate on the button (target) itself
			Indicator: doors.IndicatorOnlyAttr("aria-busy", "true"),
			On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
				// call the provided apply function
				f.apply(ctx, p)
				return true
			},
		}
		<button id="submit-location">Confirm Location</button>
		// focus on the button when rendered
		@focus("submit-location")
	}
}
```

Provide the apply function to the selector:

`./page.templ`

```templ
templ (hp *page) Body() {
	@doors.Sub(hp.id, func(id int) templ.Component {
		if id == -1 {
			return locationSelector(func(ctx context.Context, city driver.Place) {
				// mutate the path beam
				hp.path.Mutate(ctx, func(p Path) Path {
					// switch from selector to dashboard
					p.Selector = false
					p.Dashboard = true
					// set id value
					p.Id = city.Id
					return p
				})
			})
		}
		return hp.showLocation(id)
	})
}
```

**Location selected via path mutation:**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/ijcty8ucbq39p0b00mgg.gif)

> Instead of passing a path mutation function, we could just render a link. This example shows how to navigate programmatically.

## 3. Dynamic Title

In *doors* you can register a JS handler on the front end with `$d.on(...)` and invoke it from Go with `doors.Call(...)`. Implementing a dynamic title is straightforward. 

However, you can just use premade `doors.Head` component:

```templ
templ (hp *page) Head() {
	// title depends on the beam with id
	@doors.Head(hp.id, func(id int) doors.HeadData {
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
```

> Besides `title`, it also supports `<meta>` tags.

---

Next: [Advanced Scopes Usage](./07-scopes.md)

---

## Code

`./location_selector.templ`

```templ
package main

import (
	"context"
	"github.com/derstruct/doors-dashboard/driver"
	"github.com/doors-dev/doors"
	"time"
)

type selectedLocation struct {
	country driver.Place
	city    driver.Place
}

func locationSelector(apply func(context.Context, driver.Place)) templ.Component {
	location := doors.NewSourceBeam(selectedLocation{})
	// derive the country beam
	country := doors.NewBeam(location, func(p selectedLocation) driver.Place {
		return p.country
	})
	// derive the city beam
	city := doors.NewBeam(location, func(p selectedLocation) driver.Place {
		return p.city
	})
	return doors.F(&locationSelectorFragment{
		location: location,
		country:  country,
		city:     city,
		apply:    apply,
	})
}

type locationSelectorFragment struct {
	location doors.SourceBeam[selectedLocation]
	country  doors.Beam[driver.Place]
	city     doors.Beam[driver.Place]
	apply    func(context.Context, driver.Place)
}

templ (f *locationSelectorFragment) Render() {
	<article>
		// country selector fragment
		@doors.F(&placeSelector{
			label: "Country",
			query: driver.Countries.Search,
			update: func(ctx context.Context, p driver.Place) {
				// update location with the selected country
				// and no city selected
				f.location.Update(ctx, selectedLocation{country: p})
			},
			// provide the country beam
			value: f.country,
		})
		// city depends on the country
		@doors.Sub(f.country, func(country driver.Place) templ.Component {
			// no country selected, no need to render the city selection
			if country.Name == "" {
				return nil
			}
			return doors.F(&placeSelector{
				label: "City",
				query: func(s string) ([]driver.Place, error) {
					// search for cities in the provided country
					return driver.Cities.Search(country.Id, s)
				},
				update: func(ctx context.Context, p driver.Place) {
					// mutate location with the new city
					f.location.Mutate(ctx, func(sl selectedLocation) selectedLocation {
						sl.city = p
						return sl
					})
				},
				// provide the city beam
				value: f.city,
			})
		})
		// submit depends on the city beam
		@doors.Sub(f.city, func(p driver.Place) templ.Component {
			return f.submit(p)
		})
	</article>
}

templ (f *locationSelectorFragment) submit(p driver.Place) {
	// nothing selected
	if p.Name == "" {
		<button disabled>Confirm Location</button>
	} else {
		@doors.AClick{
			// block repeated clicks
			Scope: doors.ScopeOnlyBlocking(),
			// indicate on the button (target) itself
			Indicator: doors.IndicatorOnlyAttr("aria-busy", "true"),
			On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
				// call the provided apply function
				f.apply(ctx, p)
				return true
			},
		}
		<button id="submit-location">Confirm Location</button>
		// focus on the button when rendered
		@focus("submit-location")
	}
}

type placeSelector struct {
	// label for headers
	label string
	// function to search
	query func(string) ([]driver.Place, error)
	// function to update selected value
	update func(context.Context, driver.Place)
	// beam that holds selected value
	value       doors.Beam[driver.Place]
	optionsDoor doors.Door
}

templ (f *placeSelector) Render() {
	// some layout
	<section>
		// subscribe to the provided beam
		@doors.Sub(f.value, func(p driver.Place) templ.Component {
			if p.Name == "" {
				return f.selectPlace()
			}
			return f.showSelectedPlace(p)
		})
	</section>
}

templ (f *placeSelector) showSelectedPlace(p driver.Place) {
	// use label
	<h3>{ f.label }: <b>{ p.Name }</b></h3>
	@doors.AClick{
		Indicator: doors.IndicatorOnlyAttr("aria-busy", "true"),
		Scope:     doors.ScopeOnlyBlocking(),
		On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
			// update via provided function
			f.update(ctx, driver.Place{})
			return true
		},
	}
	<button class="secondary">Change</button>
}

templ (f *placeSelector) selectPlace() {
	// use label and construct unique id for the loader
	<h3>Select { f.label } &emsp;<span id={ "search-loader-" + f.label }></span></h3>
	@f.input()
	{{ f.optionsDoor.Clear(ctx) }}
	@f.optionsDoor
}

templ (f *placeSelector) input() {
	{{ prevValue := "" }}
	@doors.AInput{
		// use unique id to apply indication
		Indicator: doors.IndicatorOnlyAttrQuery("#search-loader-"+f.label, "aria-busy", "true"),
		Scope:     doors.ScopeOnlyDebounce(300*time.Millisecond, 600*time.Millisecond),
		On: func(ctx context.Context, r doors.REvent[doors.InputEvent]) bool {
			term := r.Event().Value
			term = term[:min(len(term), 16)]
			if term == prevValue {
				return false
			}
			prevValue = term
			if len(term) == 0 {
				f.optionsDoor.Clear(ctx)
				return false
			}
			f.optionsDoor.Update(ctx, f.options(term))
			return false
		},
	}
	{{ inputId := "search-input-" + f.label }}
	<input id={ inputId } type="search" placeholder={ f.label } aria-label={ f.label } autocomplete="off"/>
	@focus(inputId)
}

templ focus(id string) {
	@doors.AData{
		Name:  "id",
		Value: id,
	}
	@doors.Script() {
		<script>
            const id = $d.data("id")
            const el = document.getElementById(id)
            el.focus()
        </script>
	}
}

templ (f *placeSelector) options(term string) {
	if len(term) < 2 {
		<p>
			<mark>Type at least two letters to search</mark>
		</p>
	} else {
		{{ places, _ := f.query(term) }}
		if len(places) == 0 {
			<i>Nothing found</i>
		} else {
			{{ scope := doors.ScopeOnlyBlocking() }}
			for _, place := range places {
				// attach keydown event hook
				@doors.AKeyDown{
					Scope: scope,
					// filter by "Enter" key
					Filter: []string{"Enter"},
					On: func(ctx context.Context, r doors.REvent[doors.KeyboardEvent]) bool {
						f.update(ctx, place)
						return true
					},
				}
				@doors.AClick{
					Scope: scope,
					On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
						f.update(ctx, place)
						return true
					},
				}
				// add tabindex attribute
				<p tabindex="0" role="link" class="secondary">
					{ place.Name }
				</p>
			}
		}
	}
}
```

`./page.templ`

```templ
package main

import (
	"context"
	"github.com/derstruct/doors-dashboard/driver"
	"github.com/doors-dev/doors"
)

type Path struct {
	Selector  bool `path:"/"`
	Dashboard bool `path:"/:Id"`
	Id        int
}

func Handler(p doors.PageRouter[Path], r doors.RPage[Path]) doors.PageRoute {
	return p.Page(&page{})
}

type page struct {
	path doors.SourceBeam[Path]
	id   doors.Beam[int]
}

func (hp *page) Render(path doors.SourceBeam[Path]) templ.Component {
	// store path beam
	hp.path = path
	// derive beam with id
	hp.id = doors.NewBeam(path, func(p Path) int {
		// if the dashboard variant is active
		if p.Dashboard {
			return p.Id
		}
		// means location is not selected
		return -1
	})
	return Template(hp)
}

templ (hp *page) Head() {
	// title depends on the beam with id
	@doors.Head(hp.id, func(id int) doors.HeadData {
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

templ (hp *page) Body() {
	@doors.Sub(hp.id, func(id int) templ.Component {
		if id == -1 {
			return locationSelector(func(ctx context.Context, city driver.Place) {
				// mutate the path beam
				hp.path.Mutate(ctx, func(p Path) Path {
					// switch from selector to dashboard
					p.Selector = false
					p.Dashboard = true
					// set id value
					p.Id = city.Id
					return p
				})
			})
		}
		return hp.showLocation(id)
	})
}

templ (hp *page) showLocation(id int) {
	<article>
		{{ city, _ := driver.Cities.Get(id) }}
		if city.Name == "" {
			// set the response status code
			@doors.Status(404)
			<h1>Location Not Found</h1>
		} else {
			<h1>{ city.Name }, { city.Country.Name }</h1>
		}
	</article>
}
```

