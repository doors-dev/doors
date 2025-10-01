# Integrations

> Using a database, calling an external API, or generating SVG isn’t the main focus of this tutorial. I recommend just copying and pasting those snippets.

## 1. Database

To search for countries and cities, I used SQLite databases from [here](github.com/dr5hn/countries-states-cities-database) 

```bash
./sqlite
├── cities.sqlite3
└── countries.sqlite3
```

To use them, install the SQLite library:

```bash
 go get github.com/mattn/go-sqlite3
```

`./driver/countries.go`

```go
package driver

import (
	"database/sql"
	"errors"
)

type CountriesDb struct {
	db *sql.DB
}

func (d *CountriesDb) Get(id int) (Place, error) {
	var p Place
	query := `
		SELECT id, name
		FROM countries
		WHERE id = ?
		LIMIT 1
	`
	row := d.db.QueryRow(query, id)
	if err := row.Scan(&p.Id, &p.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Place{}, nil
		}
		return Place{}, err
	}
	return p, nil
}

func (d *CountriesDb) Search(term string) ([]Place, error) {
	query := `
		SELECT id, name
		FROM countries
		WHERE LOWER(name) LIKE LOWER(?)
		LIMIT 7
	`
	rows, err := d.db.Query(query, term+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Place
	for rows.Next() {
		var c Place
		if err := rows.Scan(&c.Id, &c.Name); err != nil {
			return nil, err
		}
		results = append(results, c)
	}
	return results, nil
}

```

`./driver/cities.go`

```go
package driver

import (
	"database/sql"
	"errors"
)

type CitiesDb struct {
	db *sql.DB
}

type City struct {
	Name    string
	Country Place
	Id      int
	Lat     float64
	Long    float64
}

func (d *CitiesDb) Get(city int) (City, error) {
	var c City
	row := d.db.QueryRow(`
		SELECT id, name, latitude, longitude, country_id
		FROM cities
		WHERE id = ?
		LIMIT 1
	`, city)

	var countryId int
	var err error
	if err = row.Scan(&c.Id, &c.Name, &c.Lat, &c.Long, &countryId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return City{}, nil
		}
		return City{}, err
	}
	c.Country, err = Countries.Get(countryId)
	if err != nil {
		return City{}, err
	}
	return c, nil
}

func (d *CitiesDb) Search(country int, term string) ([]Place, error) {
	query := `
		SELECT id, name
		FROM cities
		WHERE country_id = ?
		AND LOWER(name) LIKE LOWER(?)
		LIMIT 7
	`
	rows, err := d.db.Query(query, country, term+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Place
	for rows.Next() {
		var c Place
		if err := rows.Scan(&c.Id, &c.Name); err != nil {
			return nil, err
		}
		results = append(results, c)
	}
	return results, nil
}
```

## 3. Weather API

Connect a weather API to retrieve time-series data for the dashboard charts.

`./driver/weather.go`

```go
package driver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"fmt"
)

type WeatherAPI struct {
	endpoint string
	timeout  time.Duration
}


func (w *WeatherAPI) parseTime(str string) (time.Time, error) {
	layout := "2006-01-02T15:04"
	return time.Parse(layout, str)
}

func (w *WeatherAPI) request(ctx context.Context, city City, parameter parameter, units Units, days int) (Response, error) {
	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()
	url := fmt.Sprintf(
		"%s?latitude=%.2f&longitude=%.2f%s%s&forecast_days=%d",
		w.endpoint,
		city.Lat, city.Long,
		parameter.param(), units.param(),
		days,
	)
	var r Response
	res, err := http.Get(url)
	if err != nil {
		return r, err
	}
	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		return r, err
	}
	for i, v := range r.Hourly.Time {
		t, err := w.parseTime(v)
		if err != nil {
			return r, err
		}
		if days < 3 {
			r.Hourly.Time[i] = t.Format("15:04") + " " 
		} else {
			r.Hourly.Time[i] = t.Format("02.01")
		}
	}
	return r, err
}

type FloatSamples struct {
	Labels []string  `json:"labels"`
	Values []float64 `json:"values"`
}

type StringSamples struct {
	Labels []string `json:"labels"`
	Values []string `json:"values"`
}

func (w *WeatherAPI) Humidity(ctx context.Context, city City, days int) (FloatSamples, error) {
	r, err := w.request(ctx, city, humidity, noUnits, days)
	if err != nil {
		return FloatSamples{}, err
	}
	samples := FloatSamples{
		Labels: make([]string, len(r.Hourly.Time)),
		Values: make([]float64, len(r.Hourly.Time)),
	}
	for i := range r.Hourly.Time {
		samples.Labels[i] = r.Hourly.Time[i]
		samples.Values[i] = r.Hourly.RelativeHumidity2m[i]
	}
	return samples, nil
}

func (w *WeatherAPI) Temperature(ctx context.Context, city City, units Units, days int) (FloatSamples, error) {
	r, err := w.request(ctx, city, temperature, units, days)
	if err != nil {
		return FloatSamples{}, err
	}
	samples := FloatSamples{
		Labels: make([]string, len(r.Hourly.Time)),
		Values: make([]float64, len(r.Hourly.Time)),
	}
	for i := range r.Hourly.Time {
		samples.Labels[i] = r.Hourly.Time[i]
		samples.Values[i] = r.Hourly.Temperature2m[i]
	}
	return samples, nil
}

func (w *WeatherAPI) WindSpeed(ctx context.Context, city City, units Units, days int) (FloatSamples, error) {
	r, err := w.request(ctx, city, windSpeed, noUnits, days)
	if err != nil {
		return FloatSamples{}, err
	}
	samples := FloatSamples{
		Labels: make([]string, len(r.Hourly.Time)),
		Values: make([]float64, len(r.Hourly.Time)),
	}
	for i := range r.Hourly.Time {
		samples.Labels[i] = r.Hourly.Time[i]
		samples.Values[i] = r.Hourly.WindSpeed10m[i]
	}
	return samples, nil
}

func (w *WeatherAPI) Code(ctx context.Context, city City, days int) (StringSamples, error) {
	r, err := w.request(ctx, city, weatherCode, Metric, days)
	if err != nil {
		return StringSamples{}, err
	}
	samples := StringSamples{
		Labels: make([]string, len(r.Hourly.Time)),
		Values: make([]string, len(r.Hourly.Time)),
	}
	for i := range r.Hourly.Time {
		samples.Labels[i] = r.Hourly.Time[i]
		str, ok := weatherCodeShort[r.Hourly.WeatherCode[i]]
		if !ok {
			str = "unknown"
		}
		samples.Values[i] = str
	}
	return samples, nil
}

var weatherCodeShort = map[int]string{
	0:  "Clear",
	1:  "Mainly clear",
	2:  "Partly cloudy",
	3:  "Overcast",
	45: "Fog",
	48: "Rime fog",
	51: "Drizzle light",
	53: "Drizzle mod",
	55: "Drizzle dense",
	56: "Frzg drizzle lgt",
	57: "Frzg drizzle hvy",
	61: "Rain light",
	63: "Rain mod",
	65: "Rain heavy",
	66: "Frzg rain lgt",
	67: "Frzg rain hvy",
	71: "Snow light",
	73: "Snow mod",
	75: "Snow heavy",
	77: "Snow grains",
	80: "Shower rain lgt",
	81: "Shower rain mod",
	82: "Shower rain hvy",
	85: "Snow shower lgt",
	86: "Snow shower hvy",
	95: "Thunderstorm",
	96: "Storm + small hail",
	99: "Storm + heavy hail",
}

type Units int

const (
	Metric Units = iota
	Imperial
	noUnits
)

func (u Units) Label() string {
	if u == Imperial {
		return "Imperial"
	}
	if u == Metric {
		return "Metric"
	}
	return "unknown"
}


func (u Units) WindSpeed() string {
	if u == Imperial {
		return "KMH"
	}
	if u == Metric {
		return "MPH"
	}
	return "unknown"
}

func (u Units) Temperature() string {
	if u == Imperial {
		return " °F"
	}
	if u == Metric {
		return " °C"
	}
	return "unknown"
}

func (u Units) Humidity() string {
	return "%"
}

func (u Units) Ref() *Units {
	return &u
}

func (u Units) param() string {
	if u == Metric {
		return "&wind_speed_unit=kmh&temperature_unit=celsius&precipitation_unit=mm"
	}
	if u == Imperial {
		return "&wind_speed_unit=mph&temperature_unit=fahrenheit&precipitation_unit=inch"
	}
	return ""
}

type parameter string

const (
	temperature parameter = "temperature_2m"
	humidity    parameter = "relative_humidity_2m"
	windSpeed   parameter = "wind_speed_10m"
	weatherCode parameter = "weather_code"
)

func (u parameter) param() string {
	return "&hourly=" + string(u)
}

type Response struct {
	Hourly struct {
		Time               []string  `json:"time"`
		Temperature2m      []float64 `json:"temperature_2m"`
		RelativeHumidity2m []float64 `json:"relative_humidity_2m"`
		WindSpeed10m       []float64 `json:"wind_speed_10m"`
		Rain               []float64 `json:"rain"`
		WeatherCode        []int     `json:"weather_code"`
	} `json:"hourly"`
}

```

## 3. Charts SVG

I used [github.com/vicanso/go-charts](https://github.com/vicanso/go-charts/) to generate SVG based on the weather data.

```bash
 go get github.com/vicanso/go-charts
```

`./driver/charts.go`

```go
package driver

import (
	"sort"

	"github.com/vicanso/go-charts/v2"
)

func ChartLine(values []float64, labels []string, unit string) ([]byte, error) {
	p, err := charts.LineRender(
		[][]float64{values},
		charts.SVGTypeOption(),
		func(opt *charts.ChartOption) {
			chartDefaults(opt)
			opt.XAxis = charts.NewXAxisOption(labels)
			opt.SymbolShow = charts.FalseFlag()
			opt.Legend = charts.LegendOption{
				Data: []string{unit},
			}
			opt.LineStrokeWidth = 2
		},
	)
	if err != nil {
		return nil, err
	}
	return p.Bytes()

}

func ChartPie(values []string) ([]byte, error) {
	m := make(map[string]float64)
	for _, v := range values {
		c, _ := m[v]
		m[v] = c + 1
	}
	counts := make([]float64, len(m))
	labels := make([]string, len(m))
	i := 0
	for k := range m {
		labels[i] = k
		i += 1
	}
	sort.Strings(labels)
	for i, label := range labels {
		counts[i] = m[label]
	}
	p, err := charts.PieRender(
		counts,
		charts.SVGTypeOption(),
		charts.PieSeriesShowLabel(),
		func(opt *charts.ChartOption) {
			chartDefaults(opt)
			f := false
			opt.Legend = charts.LegendOption{
				Orient: charts.OrientVertical,
				Data:   labels,
				Show:   &f,
			}
		},
	)
	if err != nil {
		return nil, err
	}
	return p.Bytes()
}

func chartDefaults(opt *charts.ChartOption) {
	opt.Theme = "dark"
	opt.Height = 400
	opt.BackgroundColor = charts.Color{
		R: 24,
		G: 28,
		B: 37,
		A: 255,
	}
}

```

> github.com/vicanso/go-charts is good enough for tutorial, but it’s buggy and no longer maintained.

## 4. Init

Initialize wrappers for external use.

`./driver/driver.go`

```go
package driver

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Place struct {
	Id   int
	Name string
}

var Countries *CountriesDb
var Cities *CitiesDb
var Weather *WeatherAPI

func init() {
	countries, err := sql.Open("sqlite3", "./sqlite/countries.sqlite3")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	Countries = &CountriesDb{db: countries}
	cities, err := sql.Open("sqlite3", "./sqlite/cities.sqlite3")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	Cities = &CitiesDb{db: cities}
	Weather = &WeatherAPI{
		endpoint: "https://api.open-meteo.com/v1/forecast",
		timeout:  10 * time.Second,
	}
}

```

---

Next: [Location Selector](./05-location-selector.md)
