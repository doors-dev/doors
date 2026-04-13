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

func (stubSettings) Conf() *common.SystemConf {
	return &common.SystemConf{ServerCacheControl: common.DefaultCacheControl}
}

func (stubSettings) BuildProfiles() resources.BuildProfiles {
	return stubProfiles{}
}

func TestAdaptersAndLocationAdapter(t *testing.T) {
	var adapters Adapters
	adapters.Add(NewLocationAdapter())

	loc := Location{
		Segments: []string{"a b", "c"},
		Query:    url.Values{"tag": {"x"}},
	}
	got, err := adapters.Encode(loc)
	if err != nil {
		t.Fatal(err)
	}
	if !EqualLocation(got, loc) {
		t.Fatalf("unexpected encoded location: %#v", got)
	}

	if _, err := (Adapters{}).Encode(loc); err == nil {
		t.Fatal("expected missing adapter error")
	}

	adapter := NewLocationAdapter()
	if decoded, ok := adapter.Decode(loc); !ok || !EqualLocation(*decoded, loc) {
		t.Fatal("expected location adapter to decode location values")
	}
	if decoded, ok := adapter.Decode(&loc); !ok || !EqualLocation(*decoded, loc) {
		t.Fatal("expected location adapter to decode location pointers")
	}
	if _, ok := adapter.Decode(123); ok {
		t.Fatal("expected location adapter to reject wrong type")
	}
	if out, err, ok := adapter.EncodeAny(loc); !ok || err != nil || !EqualLocation(out, loc) {
		t.Fatal("expected location adapter to encode location values")
	}
	if _, _, ok := adapter.EncodeAny(123); ok {
		t.Fatal("expected location adapter to reject wrong type in EncodeAny")
	}
	if out, err := adapter.Encode(&loc); err != nil || !EqualLocation(out, loc) {
		t.Fatal("expected location adapter Encode to return same location")
	}
}

func TestGenericAdapterHelpers(t *testing.T) {
	type page struct {
		Home bool `path:"/"`
		Post bool `path:"/posts/:ID"`
		ID   int
		Tag  *string `query:"tag"`
	}

	adapter, err := NewAdapter[page]()
	if err != nil {
		t.Fatal(err)
	}
	tag := "go"
	model := page{Post: true, ID: 42, Tag: &tag}

	if asserted, ok := adapter.Assert(model); !ok || asserted.ID != 42 {
		t.Fatal("expected adapter to assert model values")
	}
	if asserted, ok := adapter.Assert(&model); !ok || asserted.ID != 42 {
		t.Fatal("expected adapter to assert model pointers")
	}
	if _, ok := adapter.Assert("bad"); ok {
		t.Fatal("expected adapter assert to reject wrong type")
	}

	loc, err, ok := adapter.EncodeAny(model)
	if !ok || err != nil || loc.String() != "/posts/42?tag=go" {
		t.Fatalf("unexpected generic EncodeAny result: %#v %v %v", loc, err, ok)
	}
	if _, _, ok := adapter.EncodeAny("bad"); ok {
		t.Fatal("expected generic EncodeAny to reject wrong type")
	}
	decodeAny := any(adapter).(interface {
		DecodeAny(any) (any, bool)
	})
	if decoded, ok := decodeAny.DecodeAny(loc); !ok || decoded.(*page).ID != 42 {
		t.Fatal("expected generic DecodeAny to decode locations")
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

	decoded, err := NewLocationFromEscapedURI("/docs/hello%20world?tag=x")
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Segments[1] != "hello world" || decoded.Query.Get("tag") != "x" {
		t.Fatalf("unexpected decoded location: %#v", decoded)
	}

	if _, err := NewLocationFromEscapedURI("/%zz"); err == nil {
		t.Fatal("expected invalid escaped path to fail")
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
	if zero := NewPathMaker(""); zero.ID() != "0" {
		t.Fatalf("expected empty server id to normalize to 0, got %q", zero.ID())
	}

	hookPath := pm.Hook("inst1", 10, 20, "file.txt")
	req := httptest.NewRequest("GET", hookPath+"?t=7", nil)
	match, ok := pm.Match(req)
	if !ok {
		t.Fatal("expected hook path to match")
	}
	hook, ok := match.Hook()
	if !ok || hook.Instance != "inst1" || hook.Door != 10 || hook.Hook != 20 || hook.Track != 7 {
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

	defer func() {
		if recover() == nil {
			t.Fatal("expected slash in server id to panic")
		}
	}()
	_ = NewPathMaker("bad/id")
}

func TestPathValidationErrors(t *testing.T) {
	type optionalNotLast struct {
		V  bool `path:"/docs/:ID?/tail"`
		ID *string
	}
	if _, err := NewAdapter[optionalNotLast](); err == nil {
		t.Fatal("expected optional non-last segment error")
	}

	type multiNotLast struct {
		V    bool `path:"/docs/:Rest+/tail"`
		Rest []string
	}
	if _, err := NewAdapter[multiNotLast](); err == nil {
		t.Fatal("expected multi non-last segment error")
	}

	type optionalNonPtr struct {
		V  bool `path:"/docs/:ID?"`
		ID string
	}
	if _, err := NewAdapter[optionalNonPtr](); err == nil {
		t.Fatal("expected optional non-pointer error")
	}

	type requiredPtr struct {
		V  bool `path:"/docs/:ID"`
		ID *string
	}
	if _, err := NewAdapter[requiredPtr](); err == nil {
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
