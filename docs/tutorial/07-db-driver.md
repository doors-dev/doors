# Database Driver

Before we proceed, we need a DB to work with. Let's use SQLite.

## 1. Install SQLite

```bash
$ go get github.com/mattn/go-sqlite3 
```

## 2. Driver Package

Basic CRUD for categories+items and session management. You can just copy paste.

> It panics on any error,  that's ok for our tutorial

### Category Management

`./driver/cats.go`

```templ
package driver

import (
	"database/sql"
	"log"
	"strings"
)

type Cat struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func newCatsDriver(db *sql.DB) *CatsDriver {
	initQuery := `
		CREATE TABLE IF NOT EXISTS cats (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			desc TEXT
		);
	`
	if _, err := db.Exec(initQuery); err != nil {
		log.Fatal("Failed to create cats table:", err)
	}
	return &CatsDriver{
		db: db,
	}
}

type CatsDriver struct {
	db *sql.DB
}

func (c *CatsDriver) populate() {
	var count int
	err := c.db.QueryRow("SELECT COUNT(*) FROM cats").Scan(&count)
	if err != nil {
		panic(err)
	}
	if count != 0 {
		return
	}
	sampleCats := []Cat{
		{Id: "electronics", Name: "Electronics", Desc: "Phones, laptops, and gadgets"},
		{Id: "books", Name: "Books", Desc: "Fiction, non-fiction, and educational books"},
		{Id: "clothing", Name: "Clothing", Desc: "Shirts, pants, shoes, and accessories"},
		{Id: "home", Name: "Home & Garden", Desc: "Furniture, decor, and gardening supplies"},
		{Id: "sports", Name: "Sports & Outdoors", Desc: "Equipment for fitness and outdoor activities"},
		{Id: "food", Name: "Food & Beverages", Desc: "Snacks, drinks, and cooking ingredients"},
	}
	for _, cat := range sampleCats {
		c.Create(cat)
	}
}

func (c *CatsDriver) Get(id string) (Cat, bool) {
	var cat Cat
	err := c.db.QueryRow("SELECT id, name, desc FROM cats WHERE id = ?", id).
		Scan(&cat.Id, &cat.Name, &cat.Desc)

	if err != nil {
		if err == sql.ErrNoRows {
			return Cat{}, false
		}
		panic(err)
	}

	return cat, true
}

func (c *CatsDriver) List() []Cat {
	rows, err := c.db.Query("SELECT id, name, desc FROM cats")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var cats []Cat
	for rows.Next() {
		var cat Cat
		err := rows.Scan(&cat.Id, &cat.Name, &cat.Desc)
		if err != nil {
			panic(err)
		}
		cats = append(cats, cat)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	return cats
}

func (c *CatsDriver) Create(cat Cat) bool {
	_, err := c.db.Exec("INSERT INTO cats(id, name, desc) VALUES(?, ?, ?)", cat.Id, cat.Name, cat.Desc)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return false // Id already exists
		}
		panic(err)
	}
	return true
}

func (c *CatsDriver) Remove(id string) bool {
	Items.removeItemsByCat(id)
	result, err := c.db.Exec("DELETE FROM cats WHERE id = ?", id)
	if err != nil {
		panic(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		panic(err)
	}
	return rowsAffected > 0
}

func (c *CatsDriver) Update(cat Cat) bool {
	result, err := c.db.Exec("UPDATE cats SET name = ?, desc = ? WHERE id = ?",
		cat.Name, cat.Desc, cat.Id)
	if err != nil {
		panic(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		panic(err)
	}
	return rowsAffected > 0
}

```

### Item Management

`./driver/items.go`

```templ
package driver

import (
	"database/sql"
	"log"
)

type Item struct {
	Id     int    `json:"id"`
	Cat    string `json:"cat"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Rating int    `json:"rating"`
}

func newItemsDriver(db *sql.DB) *ItemsDriver {
	initQuery := `
		CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cat TEXT NOT NULL,
			name TEXT NOT NULL,
			desc TEXT,
			rating INTEGER
		);
	`
	if _, err := db.Exec(initQuery); err != nil {
		log.Fatal("Failed to create items table:", err) // Fixed
	}
	return &ItemsDriver{
		db: db,
	}
}

type ItemsDriver struct {
	db *sql.DB
}

const onPage = 6

func (d *ItemsDriver) CountPages(catId string) int {
	var total int
	err := d.db.QueryRow("SELECT COUNT(*) FROM items WHERE cat = ?", catId).Scan(&total)
	if err != nil {
		panic(err)
	}
	pages := total / onPage
	if total%onPage > 0 {
		pages += 1
	}
	return pages
}

func (d *ItemsDriver) List(catId string, page int) []Item {
	offset := page * onPage
	rows, err := d.db.Query("SELECT id, cat, name, desc, rating FROM items WHERE cat = ? LIMIT ? OFFSET ?",
		catId, onPage, offset)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.Id, &item.Cat, &item.Name, &item.Desc, &item.Rating)
		if err != nil {
			panic(err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}
	return items
}

func (d *ItemsDriver) Get(id int) (Item, bool) {
	var item Item
	err := d.db.QueryRow("SELECT id, cat, name, desc, rating FROM items WHERE id = ?", id).
		Scan(&item.Id, &item.Cat, &item.Name, &item.Desc, &item.Rating)
	if err != nil {
		if err == sql.ErrNoRows {
			return Item{}, false
		}
		panic(err)
	}
	return item, true
}

func (d *ItemsDriver) Create(item Item) {
	_, err := d.db.Exec("INSERT INTO items(cat, name, desc, rating) VALUES(?, ?, ?, ?)",
		item.Cat, item.Name, item.Desc, item.Rating)
	if err != nil {
		panic(err)
	}
}

func (d *ItemsDriver) Remove(id int) bool {
	result, err := d.db.Exec("DELETE FROM items WHERE id = ?", id)
	if err != nil {
		panic(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		panic(err)
	}
	return rowsAffected > 0
}

func (d *ItemsDriver) Update(item Item) bool {
	result, err := d.db.Exec("UPDATE items SET cat = ?, name = ?, desc = ?, rating = ? WHERE id = ?",
		item.Cat, item.Name, item.Desc, item.Rating, item.Id)
	if err != nil {
		panic(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		panic(err)
	}
	return rowsAffected > 0
}

func (d *ItemsDriver) removeItemsByCat(catId string) {
	_, err := d.db.Exec("DELETE FROM items WHERE cat = ?", catId)
	if err != nil {
		panic(err)
	}
}

```

### Session Management

`./driver/sessions.go`

```templ
package driver

import (
	"database/sql"
	"github.com/doors-dev/doors"
	"time"
)

func newSessionsDriver(db *sql.DB) *SessionsDriver {
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
	s := &SessionsDriver{
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

type SessionsDriver struct {
	db *sql.DB
}

func (d *SessionsDriver) cleanup() {
	for {
		<-time.After(10 * time.Minute)
		_, err := d.db.Exec("DELETE FROM sessions WHERE expire <= ?", time.Now())
		if err != nil {
			panic("Failed to cleanup expired sessions: " + err.Error())
		}
	}
}

func (d *SessionsDriver) Add(login string, dur time.Duration) Session {
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

func (d *SessionsDriver) Get(token string) (Session, bool) {
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

func (d *SessionsDriver) Remove(token string) bool {
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

### Initialization 

`./driver/init.go`

```templ
package driver

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var Cats *CatsDriver
var Items *ItemsDriver
var Sessions *SessionsDriver

func init() {
	db, err := sql.Open("sqlite3", "./tutorial.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	Cats = newCatsDriver(db)
	Items = newItemsDriver(db)
	Cats.populate()
	Sessions = newSessionsDriver(db)
}
```

## 3. Categories Listing

>  DB driver populates the categories table with sample data if it's empty.

`./catalog/main.templ`

```templ
package catalog

import (
	"github.com/derstruct/doors-tutorial/driver"
	"github.com/doors-dev/doors"
)

templ main() {
	<h1>Catalog</h1>
	// query and iterate categories list
	for _, cat := range driver.Cats.List() {
		<article>
			<header>
			  // attach href
				@doors.AHref{
				  // category path model
					Model: Path{
						IsCat: true,
						CatId: cat.Id, 
					},
				}
				<a>{ cat.Name }</a>
			</header>
			{ cat.Desc }
		</article>
	}
}

```

You can visit https://localhost:8443/catalog/ to see a list of categories from the database

> The first build may take longer than expected because of SQLite



---

Next: [Fragment](./08-fragment.md)

