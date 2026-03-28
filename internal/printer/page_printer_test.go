package printer

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

func noMeta() gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		return nil
	})
}

type pagePrinterSettings struct {
	conf *common.SystemConf
}

func (s pagePrinterSettings) Conf() *common.SystemConf {
	return s.conf
}

func (s pagePrinterSettings) BuildProfiles() resources.BuildProfiles {
	return resources.BaseProfile{}
}

func TestPagePrinterInsertsHeadBeforeBody(t *testing.T) {
	var out bytes.Buffer
	meta := gox.EditorFunc(func(cur gox.Cursor) error {
		if err := cur.InitVoid("meta"); err != nil {
			return err
		}
		if err := cur.AttrSet("name", "description"); err != nil {
			return err
		}
		if err := cur.AttrSet("content", "inserted"); err != nil {
			return err
		}
		return cur.Submit()
	})

	p := NewPagePrinter(&out, context.Background(), true, []byte(`{"imports":{"app":"/app.js"}}`), meta)

	if err := p.Send(gox.NewJobHeadOpen(context.Background(), 1, gox.KindRegular, "body", gox.NewAttrs())); err != nil {
		t.Fatal(err)
	}
	if err := p.Send(gox.NewJobHeadClose(context.Background(), 1, gox.KindRegular, "body")); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if !strings.Contains(got, "<head>") {
		t.Fatalf("expected inserted head, got %q", got)
	}
	if !strings.Contains(got, `meta content="inserted" name="description"`) &&
		!strings.Contains(got, `meta name="description" content="inserted"`) {
		t.Fatalf("expected inserted meta, got %q", got)
	}
	if !strings.Contains(got, `<script type="importmap">`) {
		t.Fatalf("expected importmap script, got %q", got)
	}
	if !strings.Contains(got, `{"imports":{"app":"/app.js"}}`) {
		t.Fatalf("expected importmap contents, got %q", got)
	}
}

func TestPagePrinterInsertsIntoExplicitHead(t *testing.T) {
	var out bytes.Buffer
	meta := gox.EditorFunc(func(cur gox.Cursor) error {
		if err := cur.InitVoid("meta"); err != nil {
			return err
		}
		if err := cur.AttrSet("name", "robots"); err != nil {
			return err
		}
		if err := cur.AttrSet("content", "noindex"); err != nil {
			return err
		}
		return cur.Submit()
	})

	p := NewPagePrinter(&out, context.Background(), true, []byte(`{"imports":{"extra":"/extra.js"}}`), meta)

	if err := p.Send(gox.NewJobHeadOpen(context.Background(), 1, gox.KindRegular, "head", gox.NewAttrs())); err != nil {
		t.Fatal(err)
	}
	if err := p.Send(gox.NewJobHeadClose(context.Background(), 1, gox.KindRegular, "head")); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if !strings.Contains(got, "<head>") || !strings.Contains(got, "</head>") {
		t.Fatalf("expected explicit head output, got %q", got)
	}
	if !strings.Contains(got, `meta content="noindex" name="robots"`) &&
		!strings.Contains(got, `meta name="robots" content="noindex"`) {
		t.Fatalf("expected inserted robots meta, got %q", got)
	}
	if !strings.Contains(got, `{"imports":{"extra":"/extra.js"}}`) {
		t.Fatalf("expected explicit head importmap, got %q", got)
	}
}

func TestPagePrinterInsertsBeforeFirstScript(t *testing.T) {
	var out bytes.Buffer
	p := NewPagePrinter(&out, context.Background(), true, []byte(`{"imports":{"boot":"/boot.js"}}`), noMeta())

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
	p := NewPagePrinter(&out, context.Background(), true, []byte(`{"imports":{"head":"/head.js"}}`), noMeta())

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
	conf := common.SystemConf{}
	common.InitDefaults(&conf)
	registry := resources.NewRegistry(pagePrinterSettings{conf: &conf})
	inst := &titleInstance{
		registry: registry,
		conf:     conf,
		license:  "licensed",
	}
	ctx := context.WithValue(context.Background(), ctex.KeyCore, core.NewCore(inst, titleDoor{}))

	var out bytes.Buffer
	p := NewPagePrinter(&out, ctx, false, nil, noMeta())
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
	if !strings.Contains(got, `data-lic="licensed"`) {
		t.Fatalf("expected license marker, got %q", got)
	}
}

func TestPagePrinterInsertedHeadPropagatesMetaError(t *testing.T) {
	expected := errors.New("meta boom")
	p := NewPagePrinter(&bytes.Buffer{}, context.Background(), true, nil, gox.EditorFunc(func(gox.Cursor) error {
		return expected
	}))

	err := p.Send(gox.NewJobHeadOpen(context.Background(), 1, gox.KindRegular, "body", gox.NewAttrs()))
	if !errors.Is(err, expected) {
		t.Fatalf("expected meta error propagation, got %v", err)
	}
}

func TestPagePrinterWaitsForMatchingHeadClose(t *testing.T) {
	var out bytes.Buffer
	p := NewPagePrinter(&out, context.Background(), true, []byte(`{"imports":{"late":"/late.js"}}`), noMeta())

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
