# Charts

`./dashboard.templ`

## 1. Temperature 

Instead of `doors.Sub`, I’ll use the `doors.Inject` helper. It essentially does the same, but instead of evaluating a function, it renders children with context that contains the beam value.

To serve the generated SVG, I’ll use `doors.ARawSrc`, which creates a `src` attribute with a custom request handler:

```templ
templ (f *dashboardFragment) temperatureChart(city driver.City) {
	<article>
		@doors.Inject("settings", f.settings) {
			<header>
				Temperature
			</header>
			// inject the beam with settings into the context
			@doors.E(func(ctx context.Context) templ.Component {
				// read the injected value
				s := ctx.Value("settings").(dashboardSettings)
			
				// request temperature data
				values, _ := driver.Weather.Temperature(ctx, city, s.units, s.days)
			
				// generate SVG
				svg, _ := driver.ChartLine(values.Values, values.Labels, s.units.Temperature())
			
				// src with the custom request handler attached
				return doors.ARawSrc{
					// remove the hook when served to allow svg garbage collection
					Once: true,
					Handler: func(w http.ResponseWriter, r *http.Request) {
						// proper content type for svg
						w.Header().Set("Content-Type", "image/svg+xml")
						// svg is text and compresses well with gzip
						w.Header().Set("Content-Encoding", "gzip")
						gz := gzip.NewWriter(w)
						gz.Write(svg)
						gz.Close()
					},
				}
			})
			// the source will be attached to this image
			<img height="auto" width="100%"/>
		}
	</article>
}

```

Render it:

```templ
templ (f *dashboardFragment) Render() {
	{{ city, _ := driver.Cities.Get(f.id) }}
	<article>
		if city.Name == "" {
			@doors.Status(404)
			<h1>Location Not Found</h1>
		} else {
			<h1>{ city.Name }, { city.Country.Name }</h1>
		}
		@f.menu()
	</article>
	// charts component
	@f.charts(city)
}

templ (f *dashboardFragment) charts(city driver.City) {
  // charts layout
	<div class="grid">
		<div>
			@f.temperatureChart(city)
		</div>
		<div></div>
	</div>
}
```

**Temperature line chart with dynamic SVG:**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/euzeo1g28297ygalr24j.gif)

## 2. All Charts

Abstract the chart component so it can be reused for all charts:

```templ
templ (f *dashboardFragment) chart(title string, generateSVG func(dashboardSettings) []byte) {
	<article>
		@doors.Inject("settings", f.settings) {
			<header>
				{ title }
			</header>
			@doors.E(func(ctx context.Context) templ.Component {
				s := ctx.Value("settings").(dashboardSettings)
				svg := generateSVG(s)
				return doors.ARawSrc{
					Once: true,
					Handler: func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "image/svg+xml")
						w.Header().Set("Content-Encoding", "gzip")
						gz := gzip.NewWriter(w)
						gz.Write(svg)
						gz.Close()
					},
				}
			})
			<img height="auto" width="100%"/>
		}
	</article>
}
```

All charts:

```templ
templ (f *dashboardFragment) charts(city driver.City) {
	<div class="grid">
		<div>
			@f.chart("Temperature", func(s dashboardSettings) []byte {
				values, _ := driver.Weather.Temperature(ctx, city, s.units, s.days)
				svg, _ := driver.ChartLine(values.Values, values.Labels, s.units.Temperature())
				return svg
			})
			@f.chart("Humidity", func(s dashboardSettings) []byte {
				values, _ := driver.Weather.Humidity(ctx, city, s.days)
				svg, _ := driver.ChartLine(values.Values, values.Labels, s.units.Humidity())
				return svg
			})
		</div>
		<div>
			@f.chart("Weather", func(s dashboardSettings) []byte {
				values, _ := driver.Weather.Code(ctx, city, s.days)
				svg, _ := driver.ChartPie(values.Values)
				return svg
			})
			@f.chart("Wind", func(s dashboardSettings) []byte {
				values, _ := driver.Weather.WindSpeed(ctx, city, s.units, s.days)
				svg, _ := driver.ChartLine(values.Values, values.Labels, s.units.WindSpeed())
				return svg
			})
		</div>
	</div>
}
```

## 3. UX improvements

Image preloader + parameter-switch indication:

```templ
templ (f *dashboardFragment) chart(title string, generateSVG func(dashboardSettings) []byte) {
	<article>
		@doors.Inject("settings", f.settings) {
			<header>
				// loader to indicate on param switching
				{ title } &emsp;<span class="chart-loader"></span>
			</header>
			// wrapper for image loader positioning
			<div class="img-wrapper">
				@doors.E(func(ctx context.Context) templ.Component {
					s := ctx.Value("settings").(dashboardSettings)
					svg := generateSVG(s)
					return doors.ARawSrc{
						Once: true,
						Handler: func(w http.ResponseWriter, r *http.Request) {
							w.Header().Set("Content-Type", "image/svg+xml")
							w.Header().Set("Content-Encoding", "gzip")
							gz := gzip.NewWriter(w)
							gz.Write(svg)
							gz.Close()
						},
					}
				})
				<img height="auto" width="100%"/>
				// loader underneath the image
				<div class="img-loader" aria-busy="true"></div>
			</div>
		}
	</article>
}
```

Update dashboard styles:

```templ
<style>
    nav.dashboard {
        display: flex;
        flex-direction: row;
        justify-content: start;
        gap: var(--pico-spacing);
        white-space: nowrap;
        flex-wrap: wrap;
    }
    .img-wrapper {
        position: relative;
        aspect-ratio: 3 / 2;
        width: 100%;
    }
    .img-loader {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        z-index: 0;
    }
    .img-wrapper img {
        position: relative;
        z-index: 1;
    }
</style>
```

Include this indication on all menu links:

```templ
func (f *dashboardFragment) chartLoader() []doors.Indicator {
	return doors.IndicatorOnlyAttrQueryAll(".chart-loader", "aria-busy", "true")
}
```

**Charts with preloaders (slow internet):**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/czbrrzagzqst4ntm6nu6.gif)

## 2. Optimization

You may have noticed that weather and humidity don’t depend on the `units` value.

Apprach as always - derive the beam that does not depend on units:

```templ
func dashboard(id int, path doors.Beam[Path]) templ.Component {
	settings := doors.NewBeam(path, func(p Path) dashboardSettings {
		s := dashboardSettings{}
		/* ... */
		return s
	})

	units := doors.NewBeam(settings, func(s dashboardSettings) driver.Units {
		return s.units
	})

	// settings, that only change on days change
	daysSettings := doors.NewBeam(settings, func(s dashboardSettings) dashboardSettings {
		return dashboardSettings{
			days: s.days,
		}
	})

	return doors.F(&dashboardFragment{
		id:           id,
		settings:     settings,
		units:        units,
		daysSettings: daysSettings,
	})
}

type dashboardFragment struct {
	id       int
	settings doors.Beam[dashboardSettings]
	units    doors.Beam[driver.Units]

	// new beam
	daysSettings doors.Beam[dashboardSettings]
}
```

Additionally, we don’t need that indication triggered, so make it more specific:

```templ
templ (f *dashboardFragment) chart(days bool, title string, generateSVG func(dashboardSettings) []byte) {
	{{
        beam := f.settings
        marker := ""
        if days {
            beam = f.daysSettings
            marker = " days"
        }
	}}
	<article>
		@doors.Inject("settings", beam) {
			<header>
				{ title } &emsp;<span class={ "chart-loader" + marker }></span>
			</header>
			<div class="img-wrapper">
				@doors.E(func(ctx context.Context) templ.Component {
					s := ctx.Value("settings").(dashboardSettings)
					svg := generateSVG(s)
					return doors.ARawSrc{
						Once: true,
						Handler: func(w http.ResponseWriter, r *http.Request) {
							w.Header().Set("Content-Type", "image/svg+xml")
							w.Header().Set("Content-Encoding", "gzip")
							gz := gzip.NewWriter(w)
							gz.Write(svg)
							gz.Close()
						},
					}
				})
				<img height="auto" width="100%"/>
				<div class="img-loader" aria-busy="true"></div>
			</div>
		}
	</article>
}
```

Additionally, we don’t need that indication triggered all the time, so make it more specific:

```templ
func (f *dashboardFragment) chartLoader(days bool) []doors.Indicator {
	selector := ".chart-loader"
	if !days {
		selector = selector + ":not(.days)"
	}
	return doors.IndicatorOnlyAttrQueryAll(selector, "aria-busy", "true")
}
```

**Final result (slow internet simulation):**

![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/a8crw99mwdm245wp3z98.gif)

> Page size:
> ![Image description](https://dev-to-uploads.s3.amazonaws.com/uploads/articles/a863wm92vgjeozc7vugw.png)
> where ~13 KB is PicoCSS and ~10 KB is the doors client.

---

Next: [Authentication](./10-authentification.md)

---

## Code

`./dashboard.templ`

```templ
package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"github.com/derstruct/doors-dashboard/driver"
	"github.com/doors-dev/doors"
	"net/http"
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

	// settings, that only change on days change
	daysSettings := doors.NewBeam(settings, func(s dashboardSettings) dashboardSettings {
		return dashboardSettings{
			days: s.days,
		}
	})

	return doors.F(&dashboardFragment{
		id:           id,
		settings:     settings,
		units:        units,
		daysSettings: daysSettings,
	})
}

type dashboardFragment struct {
	id           int
	settings     doors.Beam[dashboardSettings]
	units        doors.Beam[driver.Units]
	daysSettings doors.Beam[dashboardSettings]
}

templ (f *dashboardFragment) Render() {
	@doors.Style() {
		@f.style()
	}
	{{ city, _ := driver.Cities.Get(f.id) }}
	<article>
		if city.Name == "" {
			@doors.Status(404)
			<h1>Location Not Found</h1>
		} else {
			<h1>{ city.Name }, { city.Country.Name }</h1>
		}
		@f.menu()
	</article>
	@f.charts(city)
}

templ (f *dashboardFragment) menu() {
	// minifies css and converts inline styles to a cacheable <link rel="stylesheet"...>
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
			if days != 1 {
				m.Days = &days
			}
			if u != driver.Metric {
				m.Units = driver.Imperial.Ref()
			}
			return doors.AHref{
				Indicator: f.chartLoader(true),
				Model:     m,
				Active: doors.Active{
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
		if s.days != 1 {
			m.Days = &s.days
		}
		if s.units == driver.Metric {
			m.Units = driver.Imperial.Ref()
		} else {
			m.Units = nil
		}
		return doors.AHref{
			Indicator: f.chartLoader(false),
			Model:     m,
		}
	})
	<a class="contrast">
		&#8644; { s.units.Label() }
	</a>
}

templ (f *dashboardFragment) charts(city driver.City) {
	<div class="grid">
		<div>
			@f.chart(false, "Temerature", func(s dashboardSettings) []byte {
				values, _ := driver.Weather.Temperature(ctx, city, s.units, s.days)
				svg, _ := driver.ChartLine(values.Values, values.Labels, s.units.Temperature())
				return svg
			})
			@f.chart(true, "Humidity", func(s dashboardSettings) []byte {
				values, _ := driver.Weather.Humidity(ctx, city, s.days)
				svg, _ := driver.ChartLine(values.Values, values.Labels, s.units.Humidity())
				return svg
			})
		</div>
		<div>
			@f.chart(true, "Weather", func(s dashboardSettings) []byte {
				values, _ := driver.Weather.Code(ctx, city, s.days)
				svg, _ := driver.ChartPie(values.Values)
				return svg
			})
			@f.chart(false, "Wind", func(s dashboardSettings) []byte {
				values, _ := driver.Weather.WindSpeed(ctx, city, s.units, s.days)
				svg, _ := driver.ChartLine(values.Values, values.Labels, s.units.WindSpeed())
				return svg
			})
		</div>
	</div>
}

templ (f *dashboardFragment) chart(days bool, title string, generateSVG func(dashboardSettings) []byte) {
	{{
	beam := f.settings
	marker := ""
	if days {
		beam = f.daysSettings
		marker = " days"
	}
	}}
	<article>
		@doors.Inject("settings", beam) {
			<header>
				{ title } &emsp;<span class={ "chart-loader" + marker }></span>
			</header>
			<div class="img-wrapper">
				@doors.E(func(ctx context.Context) templ.Component {
					s := ctx.Value("settings").(dashboardSettings)
					svg := generateSVG(s)
					return doors.ARawSrc{
						Once: true,
						Handler: func(w http.ResponseWriter, r *http.Request) {
							w.Header().Set("Content-Type", "image/svg+xml")
							w.Header().Set("Content-Encoding", "gzip")
							gz := gzip.NewWriter(w)
							gz.Write(svg)
							gz.Close()
						},
					}
				})
				<img height="auto" width="100%"/>
				<div class="img-loader" aria-busy="true"></div>
			</div>
		}
	</article>
}

func (f *dashboardFragment) chartLoader(days bool) []doors.Indicator {
	selector := ".chart-loader"
	if !days {
		selector = selector + ":not(.days)"
	}
	return doors.IndicatorOnlyAttrQueryAll(selector, "aria-busy", "true")
}

templ (f *dashboardFragment) style() {
	<style>
        nav.dashboard {
            display: flex;
            flex-direction: row;
            justify-content: start;
            gap: var(--pico-spacing);
            white-space: nowrap;
            flex-wrap: wrap;
        }
        .img-wrapper {
            position: relative;
            aspect-ratio: 3 / 2;
            width: 100%;
        }
        .img-loader {
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            z-index: 0;
        }
        .img-wrapper img {
            position: relative;
            z-index: 1;
        }
    </style>
}
```

