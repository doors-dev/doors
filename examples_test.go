// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
