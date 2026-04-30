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
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestNotStruct(t *testing.T) {
	_, err := GetModelAdapter[string]()
	if err == nil {
		t.Error("Adapeter must accept only struct")
	}
	_, err = GetModelAdapter[chan int]()
	if err == nil {
		t.Error("Adapeter must accept only struct")
	}
}

func TestNoPath(t *testing.T) {
	type noPath struct {
		V  bool
		Id string
	}
	_, err := GetModelAdapter[noPath]()
	if err == nil {
		t.Error("There must be path tag")
	}
}

func testPath[V any](t *testing.T, path string, expected bool) (*V, ModelAdapter[V]) {
	a, err := GetModelAdapter[V]()
	if err != nil {
		t.Error("Can't create adapter", err)
	}
	li := NewLocationFromEscapedURI(path)
	p, ok := a.Decode(li)
	if !expected {
		if ok {
			t.Error("encoded wrong path")
		}
		return nil, ModelAdapter[V]{}
	}
	if !ok {
		t.Error("could not encode correct path")
	}
	lo, err := a.Encode(p)
	if err != nil {
		t.Error(err)
	}
	if li.String() != lo.String() {
		t.Error("encoding output did not match path input "+path, p, lo.String(), li.String())
	}
	return p, a
}

func TestRootPath(t *testing.T) {
	type path struct {
		V bool `path:""`
	}
	testPath[path](t, "/", true)
	testPath[path](t, "/abc", false)
}

func TestSimplePath(t *testing.T) {
	type path struct {
		V  bool `path:""`
		V2 bool `path:"abc"`
		V3 bool `path:"abc/1"`
	}
	v, _ := testPath[path](t, "/", true)
	if !v.V || v.V2 || v.V3 {
		t.Error("markers did not match 1")
	}
	v, _ = testPath[path](t, "/abc/", true)
	if v.V || !v.V2 || v.V3 {
		t.Error("markers did not match 2")
	}
	v, _ = testPath[path](t, "/abc/1/", true)
	if v.V || v.V2 || !v.V3 {
		t.Error("markers did not match 3")
	}
	testPath[path](t, "/a/", false)
	testPath[path](t, "/abc/2/", false)
	testPath[path](t, "/abc/1/a/", false)
}

func TestEnumPathVariants(t *testing.T) {
	type Section int
	const (
		SectionHome Section = iota
		SectionDocs
		SectionTutorial
		SectionRecepies
		SectionLicense
	)
	type path struct {
		Section Section `path:"/ | docs/:ID? | /tutorial/:ID? | /recepies/:ID? | /license"`
		ID      *string
		Mode    *string `query:"mode"`
	}

	v, a := testPath[path](t, "/", true)
	if v.Section != SectionHome || v.ID != nil {
		t.Fatalf("unexpected home variant: %#v", v)
	}
	v, _ = testPath[path](t, "/docs", true)
	if v.Section != SectionDocs || v.ID != nil {
		t.Fatalf("unexpected docs variant without id: %#v", v)
	}
	v, _ = testPath[path](t, "/docs/intro?mode=full", true)
	if v.Section != SectionDocs || v.ID == nil || *v.ID != "intro" || v.Mode == nil || *v.Mode != "full" {
		t.Fatalf("unexpected docs variant with id/query: %#v", v)
	}
	v, _ = testPath[path](t, "/tutorial/start", true)
	if v.Section != SectionTutorial || v.ID == nil || *v.ID != "start" {
		t.Fatalf("unexpected tutorial variant: %#v", v)
	}
	v, _ = testPath[path](t, "/recepies/soup", true)
	if v.Section != SectionRecepies || v.ID == nil || *v.ID != "soup" {
		t.Fatalf("unexpected recepies variant: %#v", v)
	}
	v, _ = testPath[path](t, "/license", true)
	if v.Section != SectionLicense || v.ID != nil {
		t.Fatalf("unexpected license variant: %#v", v)
	}
	testPath[path](t, "/missing", false)

	id := "advanced"
	mode := "deep"
	encoded, err := a.Encode(&path{
		Section: SectionTutorial,
		ID:      &id,
		Mode:    &mode,
	})
	if err != nil {
		t.Fatal(err)
	}
	if encoded.String() != "/tutorial/advanced?mode=deep" {
		t.Fatalf("unexpected encoded tutorial location: %q", encoded.String())
	}
	encoded, err = a.Encode(&path{Section: SectionDocs})
	if err != nil {
		t.Fatal(err)
	}
	if encoded.String() != "/docs" {
		t.Fatalf("unexpected encoded optional docs location: %q", encoded.String())
	}
	encoded, err = a.Encode(&path{Section: SectionLicense, ID: &id})
	if err != nil {
		t.Fatal(err)
	}
	if encoded.String() != "/license" {
		t.Fatalf("unexpected encoded license location: %q", encoded.String())
	}
	if _, err := a.Encode(&path{Section: Section(99)}); err == nil {
		t.Fatal("expected invalid enum variant to fail encoding")
	}
}

func TestRawIntPathVariants(t *testing.T) {
	type path struct {
		Variant int `path:"/ | /one | /two/:ID"`
		ID      string
	}

	v, a := testPath[path](t, "/two/value", true)
	if v.Variant != 2 || v.ID != "value" {
		t.Fatalf("unexpected int variant decode: %#v", v)
	}
	encoded, err := a.Encode(&path{Variant: 1})
	if err != nil {
		t.Fatal(err)
	}
	if encoded.String() != "/one" {
		t.Fatalf("unexpected int variant encode: %q", encoded.String())
	}
}

func TestCapture(t *testing.T) {
	type path struct {
		V    bool `path:"/:P1/a/:P2/:P3/:P4+"`
		Valt bool `path:"/a/:P5+"`
		P1   int
		P2   float64
		P3   string
		P4   []string
		P5   []string
	}
	p1 := 593842435
	p2 := 122433.3454
	p3 := "oloadae"
	p4 := "koi/anma/edr"
	p := fmt.Sprintf("/%d/a/%s/%s/%s/", p1, strconv.FormatFloat(p2, 'f', -1, 64), p3, p4)
	v, _ := testPath[path](t, p, true)
	if !v.V || v.P1 != p1 || v.P2 != p2 || !slices.Equal(v.P4, strings.Split(p4, "/")) || p3 != v.P3 {
		t.Error("caputed values did not match")
	}
	v, _ = testPath[path](t, "/a/"+p4+"/", true)
	if !v.Valt {
		t.Error("Did not match alt branch")
	}
	parts := strings.Split(p4, "/")
	if !reflect.DeepEqual(parts, v.P5) {
		t.Error("slice parts did not match")
	}
}

func TestOptional(t *testing.T) {
	type path struct {
		V bool `path:"/a/:B?"`
		B *string
	}
	v, _ := testPath[path](t, "/a/b", true)
	if *v.B != "b" {
		t.Error("Did not match branch")
	}
	v, _ = testPath[path](t, "/a", true)
	if v.B != nil {
		t.Error("Did not match branch")
	}
}
func TestOptionalMulti(t *testing.T) {
	type path struct {
		V bool `path:"a/:B+?"`
		B []string
	}
	v, _ := testPath[path](t, "/a/b/c", true)
	if !slices.Equal(v.B, []string{"b", "c"}) {
		t.Error("Did not match branch")
	}

	v, _ = testPath[path](t, "/a", true)
	if len(v.B) > 0 {
		t.Error("Did not match branch")
	}
}

func TestQuery(t *testing.T) {
	type path struct {
		V      bool     `path:""`
		Colors []string `query:"color"`
		Price  int      `query:"price"`
		Empty  *string  `query:"empty"`
	}
	a := &path{
		V:      true,
		Colors: []string{"black", "yellow"},
		Price:  111,
	}
	v, _ := testPath[path](t, "/?color=black&color=yellow&price=111", true)
	if !reflect.DeepEqual(v, a) {
		t.Error("wrong query handeling", v, a)
	}

}
