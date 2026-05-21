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

package printer

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
	"github.com/evanw/esbuild/pkg/api"
)

func noMeta() gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		return nil
	})
}

func indexAny(s string, parts ...string) int {
	for _, part := range parts {
		if idx := strings.Index(s, part); idx >= 0 {
			return idx
		}
	}
	return -1
}

type pagePrinterSettings struct {
	conf *common.Conf
}

func (s pagePrinterSettings) Conf() *common.Conf {
	return s.conf
}

func (s pagePrinterSettings) ESProfile(string) api.BuildOptions {
	return api.BuildOptions{}
}

func (s pagePrinterSettings) Logger() *slog.Logger {
	return slog.Default()
}

func TestPagePrinterInsertsHeadBeforeBody(t *testing.T) {
	var out bytes.Buffer
	meta := gox.EditorFunc(func(cur gox.Cursor) error {
		if err := cur.InitVoid("meta"); err != nil {
			return err
		}
		if err := cur.Set("name", "description"); err != nil {
			return err
		}
		if err := cur.Set("content", "inserted"); err != nil {
			return err
		}
		return cur.Submit()
	})

	p := NewPagePrinter(&out, true, nil, []byte(`{"imports":{"app":"/app.js"}}`), meta)

	if err := p.Send(gox.NewJobHeadOpen(context.Background(), 1, gox.KindRegular, "body", gox.NewAttrs())); err != nil {
		t.Fatal(err)
	}
	if err := p.Send(gox.NewJobHeadClose(context.Background(), 1, gox.KindRegular, "body")); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	headPos := strings.Index(got, "<head>")
	headClosePos := strings.Index(got, "</head>")
	bodyPos := strings.Index(got, "<body>")
	metaPos := indexAny(got, `meta content="inserted" name="description"`, `meta name="description" content="inserted"`)
	importPos := strings.Index(got, `<script type="importmap">`)
	if headPos == -1 || headClosePos == -1 || bodyPos == -1 || strings.Count(got, "<head>") != 1 || strings.Count(got, "</head>") != 1 {
		t.Fatalf("expected inserted head, got %q", got)
	}
	if metaPos == -1 {
		t.Fatalf("expected inserted meta, got %q", got)
	}
	if importPos == -1 {
		t.Fatalf("expected importmap script, got %q", got)
	}
	if !strings.Contains(got, `{"imports":{"app":"/app.js"}}`) {
		t.Fatalf("expected importmap contents, got %q", got)
	}
	if !(headPos < metaPos && metaPos < importPos && importPos < headClosePos && headClosePos < bodyPos) {
		t.Fatalf("expected inserted head content before body, got %q", got)
	}
}

func TestPagePrinterInsertsIntoExplicitHead(t *testing.T) {
	var out bytes.Buffer
	meta := gox.EditorFunc(func(cur gox.Cursor) error {
		if err := cur.InitVoid("meta"); err != nil {
			return err
		}
		if err := cur.Set("name", "robots"); err != nil {
			return err
		}
		if err := cur.Set("content", "noindex"); err != nil {
			return err
		}
		return cur.Submit()
	})

	p := NewPagePrinter(&out, true, nil, []byte(`{"imports":{"extra":"/extra.js"}}`), meta)

	if err := p.Send(gox.NewJobHeadOpen(context.Background(), 1, gox.KindRegular, "head", gox.NewAttrs())); err != nil {
		t.Fatal(err)
	}
	if err := p.Send(gox.NewJobHeadClose(context.Background(), 1, gox.KindRegular, "head")); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	headPos := strings.Index(got, "<head>")
	headClosePos := strings.Index(got, "</head>")
	metaPos := indexAny(got, `meta content="noindex" name="robots"`, `meta name="robots" content="noindex"`)
	importPos := strings.Index(got, `<script type="importmap">`)
	if headPos == -1 || headClosePos == -1 || strings.Count(got, "<head>") != 1 || strings.Count(got, "</head>") != 1 {
		t.Fatalf("expected explicit head output, got %q", got)
	}
	if metaPos == -1 {
		t.Fatalf("expected inserted robots meta, got %q", got)
	}
	if !strings.Contains(got, `{"imports":{"extra":"/extra.js"}}`) {
		t.Fatalf("expected explicit head importmap, got %q", got)
	}
	if importPos == -1 || !(headPos < metaPos && metaPos < importPos && importPos < headClosePos) {
		t.Fatalf("expected inserted content to stay inside explicit head, got %q", got)
	}
}

func TestPagePrinterInsertsBeforeFirstScript(t *testing.T) {
	var out bytes.Buffer
	p := NewPagePrinter(&out, true, nil, []byte(`{"imports":{"boot":"/boot.js"}}`), noMeta())

	if err := p.Send(gox.NewJobHeadOpen(context.Background(), 1, gox.KindRegular, "script", gox.NewAttrs())); err != nil {
		t.Fatal(err)
	}
	if err := p.Send(gox.NewJobHeadClose(context.Background(), 1, gox.KindRegular, "script")); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if !strings.Contains(got, `<script type="importmap">{"imports":{"boot":"/boot.js"}}</script><script></script>`) {
		t.Fatalf("expected importmap before first script, got %q", got)
	}
}

func TestPagePrinterInsertsInsideHeadBeforeNestedScript(t *testing.T) {
	var out bytes.Buffer
	p := NewPagePrinter(&out, true, nil, []byte(`{"imports":{"head":"/head.js"}}`), noMeta())

	if err := p.Send(gox.NewJobHeadOpen(context.Background(), 1, gox.KindRegular, "head", gox.NewAttrs())); err != nil {
		t.Fatal(err)
	}
	if err := p.Send(gox.NewJobHeadOpen(context.Background(), 2, gox.KindRegular, "script", gox.NewAttrs())); err != nil {
		t.Fatal(err)
	}
	if err := p.Send(gox.NewJobHeadClose(context.Background(), 2, gox.KindRegular, "script")); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if !strings.Contains(got, `<head><script type="importmap">{"imports":{"head":"/head.js"}}</script><script></script>`) {
		t.Fatalf("expected importmap inserted inside head before nested script, got %q", got)
	}
}

func TestPagePrinterIncludesFrontAssetsWhenNotStatic(t *testing.T) {
	conf := common.Conf{}
	common.InitDefaults(&conf)
	registry := resources.NewRegistry(pagePrinterSettings{conf: &conf})
	inst := &titleInstance{
		registry: registry,
		conf:     conf,
		location: beam.NewSource(path.Location{}, path.EqualLocation, false),
	}
	inst.session = &titleSession{app: titleApp{conf: &inst.conf, registry: registry}}
	ctx := context.WithValue(context.Background(), common.KeyCore, core.NewCore(titleDoor{inst: inst}))

	var out bytes.Buffer
	p := NewPagePrinter(&out, false, front.Include(inst), nil, noMeta())
	if err := p.Send(gox.NewJobHeadOpen(ctx, 1, gox.KindRegular, "body", gox.NewAttrs())); err != nil {
		t.Fatal(err)
	}
	if err := p.Send(gox.NewJobHeadClose(ctx, 1, gox.KindRegular, "body")); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if !strings.Contains(got, "d0r.css") || !strings.Contains(got, "d0r.js") {
		t.Fatalf("expected front assets in non-static head, got %q", got)
	}
	if !strings.Contains(got, `id="instance"`) {
		t.Fatalf("expected instance id attribute, got %q", got)
	}
	if !strings.Contains(got, `data-prefix="/~/srv"`) {
		t.Fatalf("expected data-prefix, got %q", got)
	}
}

func TestPagePrinterInsertedHeadPropagatesMetaError(t *testing.T) {
	expected := errors.New("meta boom")
	p := NewPagePrinter(&bytes.Buffer{}, true, nil, nil, gox.EditorFunc(func(gox.Cursor) error {
		return expected
	}))

	err := p.Send(gox.NewJobHeadOpen(context.Background(), 1, gox.KindRegular, "body", gox.NewAttrs()))
	if !errors.Is(err, expected) {
		t.Fatalf("expected meta error propagation, got %v", err)
	}
}

func TestPagePrinterWaitsForMatchingHeadClose(t *testing.T) {
	var out bytes.Buffer
	p := NewPagePrinter(&out, true, nil, []byte(`{"imports":{"late":"/late.js"}}`), noMeta())

	if err := p.Send(gox.NewJobHeadOpen(context.Background(), 1, gox.KindRegular, "head", gox.NewAttrs())); err != nil {
		t.Fatal(err)
	}
	if err := p.Send(gox.NewJobHeadClose(context.Background(), 2, gox.KindRegular, "div")); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), `late`) {
		t.Fatalf("importmap should not be inserted before matching head close, got %q", out.String())
	}
	if err := p.Send(gox.NewJobHeadClose(context.Background(), 1, gox.KindRegular, "head")); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `{"imports":{"late":"/late.js"}}`) {
		t.Fatalf("expected importmap after matching close, got %q", out.String())
	}
}
