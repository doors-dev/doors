// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package path

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestNotStruct(t *testing.T) {
	_, err := NewAdapter[string]()
	if err == nil {
		t.Error("Adapeter must accept only struct")
	}
	_, err = NewAdapter[chan int]()
	if err == nil {
		t.Error("Adapeter must accept only struct")
	}
}

func TestNoPath(t *testing.T) {
	type noPath struct {
		V  bool
		Id string
	}
	_, err := NewAdapter[noPath]()
	if err == nil {
		t.Error("There must be path tag")
	}
}

func testPath[V any](t *testing.T, path string, expected bool) (*V, *Adapter[V]) {
	a, err := NewAdapter[V]()
	if err != nil {
		t.Error("Can't create adapter", err)
	}
	parsedURL, err := url.Parse(path)
	li := &Location{
		Path:  parsedURL.Path,
		Query: parsedURL.Query(),
	}
	p, ok := a.Decode(li)
	if !expected {
		if ok {
			t.Error("encoded wrong path")
		}
		return nil, nil
	}
	if !ok {
		t.Error("could not encode correct path")
	}
	lo, err := a.Encode(p)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(lo.Path, li.Path) {
		t.Error("encoding output did not match path input", lo.Path, li.Path)
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
		t.Error("markers did not match")
	}
	v, _ = testPath[path](t, "/abc/", true)
	if v.V || !v.V2 || v.V3 {
		t.Error("markers did not match")
	}
	v, _ = testPath[path](t, "/abc/1/", true)
	if v.V || v.V2 || !v.V3 {
		t.Error("markers did not match")
	}
	testPath[path](t, "/a/", false)
	testPath[path](t, "/abc/2/", false)
	testPath[path](t, "/abc/1/a/", false)
}

func TestCapture(t *testing.T) {
	type path struct {
		V    bool `path:"/:P1/a/:P2/:P3/:P4+"`
		Valt bool `path:"/a/:P5+"`
		P1   int
		P2   float64
		P3   string
		P4   string
		P5   []string
	}
	p1 := 593842435
	p2 := 122433.3454
	p3 := "oloadae"
	p4 := "koi/anma/edr"
	p := fmt.Sprintf("/%d/a/%s/%s/%s/", p1, strconv.FormatFloat(p2, 'f', -1, 64), p3, p4)
	v, _ := testPath[path](t, p, true)
	if !v.V || v.P1 != p1 || v.P2 != p2 || v.P4 != p4 || p3 != v.P3 {
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

func TestQuery(t *testing.T) {
	type path struct {
		V      bool     `path:""`
		Colors []string `query:"color"`
		Price  int      `query:"price"`
		Empty  string   `query:"empty"`
	}
	a := &path{
		V:      true,
		Colors: []string{"black", "yellow"},
		Price:  111,
	}
	v, _ := testPath[path](t, "/?color=black&color=yellow&price=111", true)
	if !reflect.DeepEqual(v, a) {
		t.Error("wrong query handeling")
	}

}
