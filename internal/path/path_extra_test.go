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

package path

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/evanw/esbuild/pkg/api"
)

type stubProfiles struct{}

func (stubProfiles) Options(string) api.BuildOptions {
	return api.BuildOptions{}
}

type stubSettings struct{}

func (stubSettings) Conf() *common.Conf {
	return &common.Conf{ServerCacheControl: common.DefaultCacheControl}
}

func (stubSettings) ESProfile(string) api.BuildOptions {
	return api.BuildOptions{}
}

func TestEncodeLocationValues(t *testing.T) {
	loc := Location{
		Segments: []string{"a b", "c"},
		Query:    url.Values{"tag": {"x"}},
	}
	got, err := Encode(loc)
	if err != nil {
		t.Fatal(err)
	}
	if !EqualLocation(got, loc) {
		t.Fatalf("unexpected encoded location: %#v", got)
	}

	if _, err = Encode(&loc); err == nil {
		t.Fatal("expected pointer location to use struct adapter path and fail")
	}
}

func TestGenericAdapterHelpers(t *testing.T) {
	type page struct {
		Home bool `path:"/"`
		Post bool `path:"/posts/:ID"`
		ID   int
		Tag  *string `query:"tag"`
	}

	adapter, err := GetModelAdapter[page]()
	if err != nil {
		t.Fatal(err)
	}
	tag := "go"
	model := page{Post: true, ID: 42, Tag: &tag}

	loc, err := adapter.Encode(&model)
	if err != nil || loc.String() != "/posts/42?tag=go" {
		t.Fatalf("unexpected Encode result: %#v %v", loc, err)
	}
	decoded, ok := adapter.Decode(loc)
	if !ok || decoded.ID != 42 {
		t.Fatalf("expected adapter to decode locations, got %#v %v", decoded, ok)
	}
}

func TestURLValuesField(t *testing.T) {
	type page struct {
		View  bool `path:"/docs/:ID"`
		ID    string
		Query url.Values
	}
	adapter, err := GetModelAdapter[page]()
	if err != nil {
		t.Fatal(err)
	}
	loc := NewLocationFromEscapedURI("/docs/intro?tag=go&tag=doors&page=1")
	decoded, ok := adapter.Decode(loc)
	if !ok {
		t.Fatal("expected adapter to decode raw query values")
	}
	if !decoded.View || decoded.ID != "intro" || !EqualLocation(Location{Segments: loc.Segments, Query: decoded.Query}, loc) {
		t.Fatalf("unexpected decoded model: %#v", decoded)
	}
	encoded, err := adapter.Encode(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !EqualLocation(encoded, loc) {
		t.Fatalf("unexpected encoded location: %#v", encoded)
	}

	type mixed struct {
		View  bool `path:"/"`
		Query url.Values
		Tag   string `query:"tag"`
	}
	if _, err := GetModelAdapter[mixed](); err == nil {
		t.Fatal("expected raw query values and query tags to conflict")
	}
}

func TestLocationHelpers(t *testing.T) {
	loc := Location{
		Segments: []string{"docs", "hello world"},
		Query:    url.Values{"tag": {"x", "y"}},
	}
	if loc.Path() != "/docs/hello%20world" {
		t.Fatalf("unexpected path: %q", loc.Path())
	}
	if loc.String() != "/docs/hello%20world?tag=x&tag=y" {
		t.Fatalf("unexpected location string: %q", loc.String())
	}
	if !EqualLocation(loc, loc) {
		t.Fatal("expected location to equal itself")
	}
	if EqualLocation(loc, Location{Segments: []string{"docs"}}) {
		t.Fatal("expected different locations to compare false")
	}

	decoded := NewLocationFromEscapedURI("/docs/hello%20world?tag=x")
	if decoded.Segments[1] != "hello world" || decoded.Query.Get("tag") != "x" {
		t.Fatalf("unexpected decoded location: %#v", decoded)
	}

	invalid := NewLocationFromEscapedURI("/%zz")
	if len(invalid.Segments) != 0 {
		t.Fatalf("expected invalid URI parse to return empty location, got %#v", invalid)
	}
}

func TestPathMakerAndMatch(t *testing.T) {
	pm := NewPathMaker("blue")
	if pm.ID() != "blue" {
		t.Fatalf("unexpected id: %q", pm.ID())
	}
	if pm.Prefix() != "/~/blue" {
		t.Fatalf("unexpected prefix: %q", pm.Prefix())
	}

	hookPath := pm.Hook("inst1", 20, "file.txt")
	req := httptest.NewRequest("GET", hookPath+"?t=7", nil)
	match, ok := pm.Match(req)
	if !ok {
		t.Fatal("expected hook path to match")
	}
	hook, ok := match.Hook()
	if !ok || hook.Instance != "inst1" || hook.Hook != 20 || hook.Track != 7 {
		t.Fatalf("unexpected hook match: %#v", hook)
	}

	registry := resources.NewRegistry(stubSettings{})
	res, err := registry.Static(resources.StaticBytes{Content: []byte("hello")}, "text/plain")
	if err != nil {
		t.Fatal(err)
	}
	resourcePath := pm.Resource(res, "hello.txt")
	match, ok = pm.Match(httptest.NewRequest("GET", resourcePath, nil))
	if !ok {
		t.Fatal("expected resource path to match")
	}
	id, ok := match.Resource()
	if !ok || id != res.ID() {
		t.Fatalf("unexpected resource match: %q", id)
	}

	match, ok = pm.Match(httptest.NewRequest("GET", pm.Prefix()+"/s/inst1", nil))
	if !ok {
		t.Fatal("expected sync path to match")
	}
	instanceID, ok := match.Sync()
	if !ok || instanceID != "inst1" {
		t.Fatalf("unexpected sync match: %q", instanceID)
	}

	match, ok = pm.Match(httptest.NewRequest("GET", pm.Prefix()+"/u/inst1/docs/guide?tag=x", nil))
	if !ok {
		t.Fatal("expected undo path to match")
	}
	undo, ok := match.Undo()
	if !ok || undo.Instance != "inst1" || undo.Location.String() != "/docs/guide?tag=x" {
		t.Fatalf("unexpected undo match: %#v", undo)
	}

	if _, ok := pm.Match(httptest.NewRequest("GET", hookPath+"?t=bad", nil)); ok {
		t.Fatal("expected bad hook track to fail matching")
	}
	if _, ok := pm.Match(httptest.NewRequest("GET", "/~/other/h/x/1/2", nil)); ok {
		t.Fatal("expected wrong prefix to fail matching")
	}
}

func TestPathValidationErrors(t *testing.T) {
	type boolThenEnum struct {
		Home    bool `path:"/"`
		Section int  `path:"/docs | /license"`
	}
	if _, err := GetModelAdapter[boolThenEnum](); err == nil {
		t.Fatal("expected bool path plus multipattern path to fail")
	}

	type enumThenBool struct {
		Section int  `path:"/docs | /license"`
		Home    bool `path:"/"`
	}
	if _, err := GetModelAdapter[enumThenBool](); err == nil {
		t.Fatal("expected multipattern path plus bool path to fail")
	}

	type stringPath struct {
		Section string `path:"/docs | /license"`
	}
	if _, err := GetModelAdapter[stringPath](); err == nil {
		t.Fatal("expected non-bool non-int path field to fail")
	}

	type optionalNotLast struct {
		V  bool `path:"/docs/:ID?/tail"`
		ID *string
	}
	if _, err := GetModelAdapter[optionalNotLast](); err == nil {
		t.Fatal("expected optional non-last segment error")
	}

	type multiNotLast struct {
		V    bool `path:"/docs/:Rest+/tail"`
		Rest []string
	}
	if _, err := GetModelAdapter[multiNotLast](); err == nil {
		t.Fatal("expected multi non-last segment error")
	}

	type optionalNonPtr struct {
		V  bool `path:"/docs/:ID?"`
		ID string
	}
	if _, err := GetModelAdapter[optionalNonPtr](); err == nil {
		t.Fatal("expected optional non-pointer error")
	}

	type requiredPtr struct {
		V  bool `path:"/docs/:ID"`
		ID *string
	}
	if _, err := GetModelAdapter[requiredPtr](); err == nil {
		t.Fatal("expected required pointer error")
	}

	type tailStar struct {
		V    bool `path:"/docs/:Rest*"`
		Rest []string
	}
	decoded, _ := testPath[tailStar](t, "/docs", true)
	if len(decoded.Rest) != 0 {
		t.Fatalf("expected empty optional tail, got %#v", decoded.Rest)
	}
	decoded, _ = testPath[tailStar](t, "/docs/a/b", true)
	if len(decoded.Rest) != 2 || decoded.Rest[0] != "a" || decoded.Rest[1] != "b" {
		t.Fatalf("unexpected optional tail capture: %#v", decoded.Rest)
	}
}
