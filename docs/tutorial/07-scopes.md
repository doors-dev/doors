# Advanced Scope Usage

In practice, the submit handler won’t respond instantly.  What if we interact with the UI during processing?

Let's simulate conflicting actions:

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/ehntbc7v1q1f7qr5mf9f.gif)

## Concurrent Scope

This issue can be easily mitigated with **Concurrent Scope** from the **Scopes API**.

> Concurrent Scope can be “occupied” only by events with the same group id.

Add a concurrent scope to the location selector:

```templ
type locationSelectorFragment struct {
	location doors.SourceBeam[selectedLocation]
	country  doors.Beam[driver.Place]
	city     doors.Beam[driver.Place]
	apply    func(context.Context, driver.Place)

	// add a concurrent scope
	scope doors.ScopeConcurrent
}
```

Add a parent scope property to the place selector:

```templ
type placeSelector struct {
	label       string
	query       func(string) ([]driver.Place, error)
	update      func(context.Context, driver.Place)
	value       doors.Beam[driver.Place]
	optionsDoor doors.Door

	// new field
	parentScope doors.Scope
}
```

Assign group Id 1 to both place selectors:

```templ
templ (f *locationSelectorFragment) Render() {
	<article>
		@doors.F(&placeSelector{
			/* ... */
			// provide the scope value
			parentScope: f.scope.Scope(1),
		})
		@doors.Sub(f.country, func(country driver.Place) templ.Component {
			// no country selected, no need to render the city selection
			if country.Name == "" {
				return nil
			}
			return doors.F(&placeSelector{
				/* ... */
				// provide the scope value
				parentScope: f.scope.Scope(1),
			})
		})
		@doors.Sub(f.city, func(p driver.Place) templ.Component {
			return f.submit(p)
		})
	</article>
}
```

Use it on the ‘change place’ button:

```templ
templ (f *placeSelector) showSelectedPlace(p driver.Place) {
	<h3>{ f.label }: <b>{ p.Name }</b></h3>
	@doors.AClick{
		Indicator: doors.IndicatorOnlyAttr("aria-busy", "true"),
		// combine blocking scope with the scope provided by location selector
		Scope: []doors.Scope{&doors.ScopeBlocking{}, f.parentScope},
		On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
			f.update(ctx, driver.Place{})
			return true
		},
	}
	<button class="secondary">Change</button>
}
```

Finally, apply it to the submit button with a different group Id:

```templ
templ (f *locationSelectorFragment) submit(p driver.Place) {
	if p.Name == "" {
		<button disabled>Confirm Location</button>
	} else {
		@doors.AClick{
			Indicator: doors.IndicatorOnlyAttr("aria-busy", "true"),
			// combine blocking scope with the concurrent scope instance
			Scope: []doors.Scope{&doors.ScopeBlocking{}, f.scope.Scope(0)},
			On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
				f.apply(ctx, p)
				return true
			},
		}
		<button id="submit-location">Confirm Location</button>
		@focus("submit-location")
	}
}
```

## Result

**This setup ensures that either the submit or the change-place event can run, not both at the same time:**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/esr06wgviglqn6xffom0.gif)

**While changes to the city and country won't affect each other, since they share the same group:**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/8vzen8wrcxgfrho0rm9t.gif)

> Concurrency control is necessary due to the framework’s non-blocking event model. This is a major advantage of _doors_ over Phoenix LiveView or Blazor Server, enabling highly interactive UIs without UX compromises.

Next: [Menu](./08-menu.md)
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

	// add a concurrent scope
	scope doors.ScopeConcurrent
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
			// provide the scope value
			parentScope: f.scope.Scope(1),
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
				// provide the scope value
				parentScope: f.scope.Scope(1),
			})
		})
		// submit depends on the city beam
		@doors.Sub(f.city, func(p driver.Place) templ.Component {
			return f.submit(p)
		})
	</article>
}

templ (f *locationSelectorFragment) submit(p driver.Place) {
	if p.Name == "" {
		<button disabled>Confirm Location</button>
	} else {
		@doors.AClick{
			Indicator: doors.IndicatorOnlyAttr("aria-busy", "true"),
			// combine blocking scope with the concurrent scope instance
			Scope: []doors.Scope{&doors.ScopeBlocking{}, f.scope.Scope(0)},
			On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
				f.apply(ctx, p)
				return true
			},
		}
		<button id="submit-location">Confirm Location</button>
		@focus("submit-location")
	}
}

type placeSelector struct {
	label       string
	query       func(string) ([]driver.Place, error)
	update      func(context.Context, driver.Place)
	value       doors.Beam[driver.Place]
	optionsDoor doors.Door

	// new field
	parentScope doors.Scope
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
	<h3>{ f.label }: <b>{ p.Name }</b></h3>
	@doors.AClick{
		Indicator: doors.IndicatorOnlyAttr("aria-busy", "true"),
		// combine blocking scope with the scope provided by location selector
		Scope: []doors.Scope{&doors.ScopeBlocking{}, f.parentScope},
		On: func(ctx context.Context, r doors.REvent[doors.PointerEvent]) bool {
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
            const id = $data("id")
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
