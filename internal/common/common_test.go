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

package common

import (
	"bytes"
	"compress/gzip"
	"io"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/gox"
)

func TestInitDefaultsAndSolitaireConf(t *testing.T) {
	conf := &SystemConf{}
	InitDefaults(conf)

	if conf.RequestTimeout != 30*time.Second {
		t.Fatalf("unexpected request timeout: %v", conf.RequestTimeout)
	}
	if conf.SessionInstanceLimit != 12 {
		t.Fatalf("unexpected session instance limit: %d", conf.SessionInstanceLimit)
	}
	if conf.InstanceGoroutineLimit != 8 {
		t.Fatalf("unexpected goroutine limit: %d", conf.InstanceGoroutineLimit)
	}
	if conf.InstanceConnectTimeout != conf.RequestTimeout {
		t.Fatal("expected instance connect timeout to default to request timeout")
	}
	if conf.InstanceTTL != 40*time.Minute {
		t.Fatalf("unexpected instance ttl: %v", conf.InstanceTTL)
	}
	if conf.DisconnectHiddenTimer != conf.InstanceTTL/2 {
		t.Fatal("expected hidden disconnect timer to default to half of instance ttl")
	}
	if conf.ServerCacheControl != DefaultCacheControl {
		t.Fatalf("unexpected cache control: %q", conf.ServerCacheControl)
	}
	if conf.SessionTTL != conf.InstanceTTL {
		t.Fatal("expected session ttl to be raised to instance ttl")
	}
	if conf.SolitaireSyncTimeout != conf.InstanceTTL {
		t.Fatal("expected solitaire sync timeout to default to instance ttl")
	}

	solitaire := GetSolitaireConf(conf)
	if solitaire.Queue != conf.SolitaireQueue || solitaire.Pending != conf.SolitairePending {
		t.Fatal("expected solitaire config to mirror system config")
	}
	if solitaire.DisableGzip != conf.SolitaireDisableGzip {
		t.Fatal("expected solitaire gzip setting to mirror system config")
	}

	custom := &SystemConf{
		RequestTimeout:          2 * time.Second,
		InstanceTTL:             3 * time.Second,
		SessionTTL:              1 * time.Second,
		SolitaireSyncTimeout:    9 * time.Second,
		SolitaireFlushSizeLimit: -1,
	}
	InitDefaults(custom)
	if custom.InstanceTTL != 4*time.Second {
		t.Fatalf("expected instance ttl to be raised to 2x request timeout, got %v", custom.InstanceTTL)
	}
	if custom.SessionTTL != custom.InstanceTTL {
		t.Fatal("expected session ttl to be raised to instance ttl")
	}
	if custom.SolitaireSyncTimeout != custom.InstanceTTL {
		t.Fatal("expected solitaire sync timeout to be clipped to instance ttl")
	}
	if custom.SolitaireFlushSizeLimit != 32*1024 {
		t.Fatal("expected solitaire flush size default")
	}
}

func TestCSPGenerateAndCollector(t *testing.T) {
	csp := &CSP{
		DefaultSources:      nil,
		ScriptSources:       []string{"https://scripts.example"},
		ScriptStrictDynamic: true,
		StyleSources:        []string{"https://styles.example"},
		ConnectSources:      []string{"https://api.example"},
		FormActions:         nil,
		ObjectSources:       []string{},
		FrameSources:        []string{"https://frame.example"},
		FrameAcestors:       nil,
		BaseURIAllow:        []string{},
		ImgSources:          []string{"data:"},
		ReportTo:            "csp-endpoint",
	}
	collector := csp.NewCollector()
	collector.StyleSource("https://style-cdn.example")
	collector.ScriptSource("https://script-cdn.example")
	collector.StyleHash([]byte("style-hash"))
	collector.ScriptHash([]byte("script-hash"))

	out := collector.Generate()
	expectedParts := []string{
		"default-src 'self'",
		"connect-src 'self' https://api.example",
		"script-src 'self'",
		"'strict-dynamic'",
		"https://scripts.example",
		"https://script-cdn.example",
		"style-src 'self'",
		"https://styles.example",
		"https://style-cdn.example",
		"form-action 'none'",
		"frame-src https://frame.example",
		"frame-ancestors 'none'",
		"img-src data:",
		"report-to csp-endpoint",
	}
	for _, part := range expectedParts {
		if !strings.Contains(out, part) {
			t.Fatalf("expected generated csp to contain %q, got %q", part, out)
		}
	}
	if strings.Contains(out, "object-src") {
		t.Fatal("expected object-src to be omitted for empty slice")
	}
	if strings.Contains(out, "base-uri") {
		t.Fatal("expected base-uri to be omitted for empty slice")
	}

	var nilCSP *CSP
	if nilCSP.NewCollector() != nil {
		t.Fatal("expected nil csp collector to stay nil")
	}
	if got := csp.simple("img-src", nil, []string{}, nil); got != "" {
		t.Fatalf("expected empty user directive to be omitted, got %q", got)
	}
	if got := csp.join([]string{"a", "", "b"}); got != "a; b" {
		t.Fatalf("unexpected joined csp value: %q", got)
	}
}

func TestCollectionAndEncodingHelpers(t *testing.T) {
	set := NewSet[string]()
	if !set.IsEmpty() {
		t.Fatal("expected new set to be empty")
	}
	if !set.Add("a") || set.Add("a") {
		t.Fatal("unexpected add result")
	}
	if !set.Has("a") || set.Len() != 1 {
		t.Fatal("expected set to contain added value")
	}
	if !set.Remove("a") || set.Remove("a") {
		t.Fatal("unexpected remove result")
	}
	set.Add("x")
	set.Add("y")
	if got := set.Slice(); len(got) != 2 || !slices.Contains(got, "x") || !slices.Contains(got, "y") {
		t.Fatalf("expected slice to expose current members, got %#v", got)
	}
	if got := set.Iter(); len(got) != 2 {
		t.Fatalf("expected iter to expose current members, got %#v", got)
	} else if _, ok := got["x"]; !ok {
		t.Fatalf("expected iter to contain %q, got %#v", "x", got)
	} else if _, ok := got["y"]; !ok {
		t.Fatalf("expected iter to contain %q, got %#v", "y", got)
	}
	set.Clear()
	if !set.IsEmpty() {
		t.Fatal("expected clear to empty the set")
	}

	attrs := gox.NewAttrs()
	attrs.Get("data-a").Set("1")
	attrs.Get("data-b").Set("2")
	mapped := AttrsToMap(attrs)
	if mapped["data-a"] != "1" || mapped["data-b"] != "2" {
		t.Fatalf("unexpected attrs map: %#v", mapped)
	}

	buf := &bytes.Buffer{}
	writer := NewJsonWriter(buf)
	n, err := writer.Write([]byte("hello\n"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 6 || buf.String() != "hello" {
		t.Fatalf("unexpected json writer result: n=%d body=%q", n, buf.String())
	}

	if got := AsString(&[]byte{'h', 'i'}); got != "hi" {
		t.Fatalf("unexpected string conversion: %q", got)
	}
	if got := string(AsBytes("ok")); got != "ok" {
		t.Fatalf("unexpected bytes conversion: %q", got)
	}
	if id := RandId(); id == "" || len(id) > 22 {
		t.Fatalf("unexpected rand id: %q", id)
	}
	if id := EncodeId([]byte("hello")); id == "" || len(id) > 22 {
		t.Fatalf("unexpected encoded id: %q", id)
	}

	zipped, err := Zip([]byte("hello world"))
	if err != nil || len(zipped) == 0 {
		t.Fatal("expected zip to produce data")
	}
	reader, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		t.Fatal(err)
	}
	unzipped, err := io.ReadAll(reader)
	reader.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(unzipped) != "hello world" {
		t.Fatalf("unexpected zipped round-trip: %q", string(unzipped))
	}
	minified, err := MinifyCSS([]byte("h1 { color: red; }"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(minified), "h1{color:red}") && !strings.Contains(string(minified), "h1{color:red;}") {
		t.Fatalf("unexpected minified css: %q", string(minified))
	}
}

func TestPrimeIDAndEndCause(t *testing.T) {
	id1 := NewID()
	id2 := NewID()
	if id1 == nil || id2 == nil || id1 == id2 {
		t.Fatal("expected unique opaque ids")
	}

	prime := NewPrime()
	v1 := prime.Gen()
	v2 := prime.Gen()
	if v1 == 0 || v2 == 0 || v1 == v2 {
		t.Fatalf("unexpected prime generator output: %d %d", v1, v2)
	}

	if got := EndCauseSuspend.Error(); got != "cause: 1" {
		t.Fatalf("unexpected end cause error: %q", got)
	}
}
