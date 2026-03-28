package doors

import (
	"fmt"
	"net/http"
	"net/url"
)

type examplePath struct {
	Home bool `path:"/"`
	Post bool `path:"/posts/:ID"`
	ID   int
}

func ExampleNewRouter() {
	router := NewRouter()

	UseModel(router, func(r RequestModel, s Source[examplePath]) Response {
		return ResponseRedirect(examplePath{Post: true, ID: 42}, http.StatusFound)
	})
}

func ExampleNewSource() {
	count := NewSource(1)
	label := NewBeam(count, func(v int) string {
		return fmt.Sprintf("count:%d", v)
	})

	fmt.Println(count.Get())
	_ = label
	// Output:
	// 1
}

func ExampleLocation() {
	loc := Location{
		Segments: []string{"posts", "42"},
		Query: url.Values{
			"tag":  []string{"go"},
			"page": []string{"2"},
		},
	}

	fmt.Println(loc.String())
	// Output:
	// /posts/42?page=2&tag=go
}
