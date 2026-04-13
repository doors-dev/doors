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

package resources

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/evanw/esbuild/pkg/api"
)

func writeTempFile(t *testing.T, name string, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func entryIDString(t *testing.T, entry interface{ entryID(idWriter) }) string {
	t.Helper()
	var buf bytes.Buffer
	entry.entryID(&buf)
	return buf.String()
}

func formatIDString(t *testing.T, entry interface{ formatID(idWriter) }) string {
	t.Helper()
	var buf bytes.Buffer
	entry.formatID(&buf)
	return buf.String()
}

func TestStaticEntries(t *testing.T) {
	fsys := fstest.MapFS{
		"asset.txt": {Data: []byte("fs-content")},
	}
	path := writeTempFile(t, "asset.txt", "path-content")

	fsEntry := StaticFS{FS: fsys, Path: "asset.txt"}
	data, err := fsEntry.Read()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "fs-content" {
		t.Fatalf("StaticFS.Read got %q", string(data))
	}
	if got := entryIDString(t, fsEntry); got != "fsasset.txtfs-content" {
		t.Fatalf("StaticFS.entryID got %q", got)
	}

	fsNamed := StaticFS{FS: fsys, Path: "asset.txt", Name: "named"}
	if got := entryIDString(t, fsNamed); got != "fsasset.txtnamed" {
		t.Fatalf("StaticFS named entryID got %q", got)
	}

	pathEntry := StaticPath{Path: path}
	data, err = pathEntry.Read()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "path-content" {
		t.Fatalf("StaticPath.Read got %q", string(data))
	}
	if got := entryIDString(t, pathEntry); got != "path"+path {
		t.Fatalf("StaticPath.entryID got %q", got)
	}

	bytesEntry := StaticBytes{Content: []byte("bytes-content")}
	data, err = bytesEntry.Read()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "bytes-content" {
		t.Fatalf("StaticBytes.Read got %q", string(data))
	}
	if got := entryIDString(t, bytesEntry); got != "contentbytes-content" {
		t.Fatalf("StaticBytes.entryID got %q", got)
	}

	stringEntry := StaticString{Content: "string-content"}
	data, err = stringEntry.Read()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "string-content" {
		t.Fatalf("StaticString.Read got %q", string(data))
	}
	if got := entryIDString(t, stringEntry); got != "contentstring-content" {
		t.Fatalf("StaticString.entryID got %q", got)
	}
}

func TestScriptEntries(t *testing.T) {
	fsys := fstest.MapFS{
		"index.js": {Data: []byte(`window.value = "js"`)},
		"index.ts": {Data: []byte(`const value: string = "ts"`)},
	}
	jsPath := writeTempFile(t, "index.js", `window.value = "path-js"`)
	tsPath := writeTempFile(t, "index.ts", `const value: string = "path-ts"`)

	t.Run("ScriptFS", func(t *testing.T) {
		entry := ScriptFS{FS: fsys, Path: "index.js"}
		data, err := entry.Read()
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `window.value = "js"` {
			t.Fatalf("ScriptFS.Read got %q", string(data))
		}
		var opt api.BuildOptions
		if err := entry.Apply(&opt); err != nil {
			t.Fatal(err)
		}
		if len(opt.EntryPoints) != 1 || opt.EntryPoints[0] != "index.js" {
			t.Fatalf("ScriptFS entry points = %#v", opt.EntryPoints)
		}
		if len(opt.Plugins) != 1 {
			t.Fatalf("ScriptFS plugins = %#v", opt.Plugins)
		}
		if got := entryIDString(t, entry); got != "fsindex.jswindow.value = \"js\"" {
			t.Fatalf("ScriptFS.entryID got %q", got)
		}
		named := ScriptFS{FS: fsys, Path: "index.js", Name: "module"}
		if got := entryIDString(t, named); got != "fsindex.jsmodule" {
			t.Fatalf("ScriptFS named entryID got %q", got)
		}
	})

	t.Run("ScriptInlineFS", func(t *testing.T) {
		js := ScriptInlineFS{FS: fsys, Path: "index.js"}
		data, err := js.Read()
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `window.value = "js"` {
			t.Fatalf("ScriptInlineFS.Read got %q", string(data))
		}
		var jsOpt api.BuildOptions
		if err := js.Apply(&jsOpt); err != nil {
			t.Fatal(err)
		}
		if jsOpt.Stdin == nil || jsOpt.Stdin.Sourcefile != "index.js" || jsOpt.Stdin.Loader != api.LoaderJS {
			t.Fatalf("ScriptInlineFS JS stdin = %#v", jsOpt.Stdin)
		}
		if !strings.Contains(jsOpt.Stdin.Contents, "_d0r(document.currentScript, async ($on, $data, $hook, $fetch, $G, $sys, HookErr)") {
			t.Fatalf("ScriptInlineFS JS contents = %q", jsOpt.Stdin.Contents)
		}
		if got := entryIDString(t, js); got != "inline_fsindex.jswindow.value = \"js\"" {
			t.Fatalf("ScriptInlineFS JS entryID got %q", got)
		}

		ts := ScriptInlineFS{FS: fsys, Path: "index.ts", Name: "inline"}
		var tsOpt api.BuildOptions
		if err := ts.Apply(&tsOpt); err != nil {
			t.Fatal(err)
		}
		if tsOpt.Stdin == nil || tsOpt.Stdin.Sourcefile != "index.ts" || tsOpt.Stdin.Loader != api.LoaderTS {
			t.Fatalf("ScriptInlineFS TS stdin = %#v", tsOpt.Stdin)
		}
		if !strings.Contains(tsOpt.Stdin.Contents, "$data: <T = any>(name: string) => T | Promise<ArrayBuffer>") {
			t.Fatalf("ScriptInlineFS TS contents = %q", tsOpt.Stdin.Contents)
		}
		if got := entryIDString(t, ts); got != "inline_fsindex.tsinline" {
			t.Fatalf("ScriptInlineFS TS entryID got %q", got)
		}

		bad := ScriptInlineFS{FS: fsys, Path: "missing.ts"}
		if err := bad.Apply(&api.BuildOptions{}); err == nil {
			t.Fatal("ScriptInlineFS.Apply should fail for missing file")
		}
	})

	t.Run("ScriptPath", func(t *testing.T) {
		entry := ScriptPath{Path: jsPath}
		data, err := entry.Read()
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `window.value = "path-js"` {
			t.Fatalf("ScriptPath.Read got %q", string(data))
		}
		var opt api.BuildOptions
		if err := entry.Apply(&opt); err != nil {
			t.Fatal(err)
		}
		if len(opt.EntryPoints) != 1 || opt.EntryPoints[0] != jsPath {
			t.Fatalf("ScriptPath entry points = %#v", opt.EntryPoints)
		}
		if got := entryIDString(t, entry); got != "path"+jsPath {
			t.Fatalf("ScriptPath.entryID got %q", got)
		}
	})

	t.Run("ScriptBytes", func(t *testing.T) {
		entry := ScriptBytes{Content: []byte(`console.log("bytes")`), Kind: KindJS}
		data, err := entry.Read()
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `console.log("bytes")` {
			t.Fatalf("ScriptBytes.Read got %q", string(data))
		}
		var opt api.BuildOptions
		if err := entry.Apply(&opt); err != nil {
			t.Fatal(err)
		}
		if opt.Stdin == nil || opt.Stdin.Sourcefile != "index.js" || opt.Stdin.Loader != api.LoaderJS {
			t.Fatalf("ScriptBytes stdin = %#v", opt.Stdin)
		}
		if got := entryIDString(t, entry); got != `contentjsconsole.log("bytes")` {
			t.Fatalf("ScriptBytes.entryID got %q", got)
		}
	})

	t.Run("ScriptInlinePath", func(t *testing.T) {
		js := ScriptInlinePath{Path: jsPath}
		data, err := js.Read()
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `window.value = "path-js"` {
			t.Fatalf("ScriptInlinePath.Read got %q", string(data))
		}
		var jsOpt api.BuildOptions
		if err := js.Apply(&jsOpt); err != nil {
			t.Fatal(err)
		}
		if jsOpt.Stdin == nil || jsOpt.Stdin.Sourcefile != "index.js" || jsOpt.Stdin.Loader != api.LoaderJS {
			t.Fatalf("ScriptInlinePath JS stdin = %#v", jsOpt.Stdin)
		}
		if got := entryIDString(t, js); got != "inline_path"+jsPath {
			t.Fatalf("ScriptInlinePath JS entryID got %q", got)
		}

		ts := ScriptInlinePath{Path: tsPath}
		var tsOpt api.BuildOptions
		if err := ts.Apply(&tsOpt); err != nil {
			t.Fatal(err)
		}
		if tsOpt.Stdin == nil || tsOpt.Stdin.Sourcefile != "index.ts" || tsOpt.Stdin.Loader != api.LoaderTS {
			t.Fatalf("ScriptInlinePath TS stdin = %#v", tsOpt.Stdin)
		}
		if !strings.Contains(tsOpt.Stdin.Contents, "$data: <T = any>(name: string) => T | Promise<ArrayBuffer>") {
			t.Fatalf("ScriptInlinePath TS contents = %q", tsOpt.Stdin.Contents)
		}

		bad := ScriptInlinePath{Path: filepath.Join(t.TempDir(), "missing.ts")}
		if err := bad.Apply(&api.BuildOptions{}); err == nil {
			t.Fatal("ScriptInlinePath.Apply should fail for missing file")
		}
	})

	t.Run("ScriptInlineString", func(t *testing.T) {
		js := ScriptInlineString{Content: `console.log("inline-string")`, Kind: KindJS}
		data, err := js.Read()
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `console.log("inline-string")` {
			t.Fatalf("ScriptInlineString.Read got %q", string(data))
		}
		var jsOpt api.BuildOptions
		if err := js.Apply(&jsOpt); err != nil {
			t.Fatal(err)
		}
		if jsOpt.Stdin == nil || jsOpt.Stdin.Sourcefile != "index.js" || jsOpt.Stdin.Loader != api.LoaderJS {
			t.Fatalf("ScriptInlineString JS stdin = %#v", jsOpt.Stdin)
		}
		if got := entryIDString(t, js); got != `inline_stringjsconsole.log("inline-string")` {
			t.Fatalf("ScriptInlineString JS entryID got %q", got)
		}

		ts := ScriptInlineString{Content: `const value: string = "ts"`, Kind: KindTS}
		var tsOpt api.BuildOptions
		if err := ts.Apply(&tsOpt); err != nil {
			t.Fatal(err)
		}
		if tsOpt.Stdin == nil || tsOpt.Stdin.Sourcefile != "index.ts" || tsOpt.Stdin.Loader != api.LoaderTS {
			t.Fatalf("ScriptInlineString TS stdin = %#v", tsOpt.Stdin)
		}
		if !strings.Contains(tsOpt.Stdin.Contents, "$data: <T = any>(name: string) => T | Promise<ArrayBuffer>") {
			t.Fatalf("ScriptInlineString TS contents = %q", tsOpt.Stdin.Contents)
		}
	})

	t.Run("ScriptInlineBytes", func(t *testing.T) {
		js := ScriptInlineBytes{Content: []byte(`console.log("inline-bytes")`), Kind: KindJS}
		data, err := js.Read()
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `console.log("inline-bytes")` {
			t.Fatalf("ScriptInlineBytes.Read got %q", string(data))
		}
		var jsOpt api.BuildOptions
		if err := js.Apply(&jsOpt); err != nil {
			t.Fatal(err)
		}
		if jsOpt.Stdin == nil || jsOpt.Stdin.Sourcefile != "index.js" || jsOpt.Stdin.Loader != api.LoaderJS {
			t.Fatalf("ScriptInlineBytes JS stdin = %#v", jsOpt.Stdin)
		}
		if got := entryIDString(t, js); got != `inline_bytesjsconsole.log("inline-bytes")` {
			t.Fatalf("ScriptInlineBytes JS entryID got %q", got)
		}

		ts := ScriptInlineBytes{Content: []byte(`const value: string = "ts"`), Kind: KindTS}
		var tsOpt api.BuildOptions
		if err := ts.Apply(&tsOpt); err != nil {
			t.Fatal(err)
		}
		if tsOpt.Stdin == nil || tsOpt.Stdin.Sourcefile != "index.ts" || tsOpt.Stdin.Loader != api.LoaderTS {
			t.Fatalf("ScriptInlineBytes TS stdin = %#v", tsOpt.Stdin)
		}
		if !strings.Contains(tsOpt.Stdin.Contents, "$data: <T = any>(name: string) => T | Promise<ArrayBuffer>") {
			t.Fatalf("ScriptInlineBytes TS contents = %q", tsOpt.Stdin.Contents)
		}
	})

	t.Run("ScriptString", func(t *testing.T) {
		entry := ScriptString{Content: `console.log("string")`, Kind: KindTS}
		data, err := entry.Read()
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `console.log("string")` {
			t.Fatalf("ScriptString.Read got %q", string(data))
		}
		var opt api.BuildOptions
		if err := entry.Apply(&opt); err != nil {
			t.Fatal(err)
		}
		if opt.Stdin == nil || opt.Stdin.Sourcefile != "index.ts" || opt.Stdin.Loader != api.LoaderTS {
			t.Fatalf("ScriptString stdin = %#v", opt.Stdin)
		}
		if got := entryIDString(t, entry); got != `contenttsconsole.log("string")` {
			t.Fatalf("ScriptString.entryID got %q", got)
		}
	})
}

func TestKindAndFormats(t *testing.T) {
	if KindJS.Loader() != api.LoaderJS {
		t.Fatal("KindJS loader mismatch")
	}
	if KindTS.Loader() != api.LoaderTS {
		t.Fatal("KindTS loader mismatch")
	}
	if Kind(99).Loader() != api.LoaderJS {
		t.Fatal("unknown kind should default to JS loader")
	}
	if KindJS.String() != "js" || KindTS.String() != "ts" || Kind(99).String() != "unknown" {
		t.Fatal("kind string mismatch")
	}

	var opt api.BuildOptions
	FormatDefault{}.Apply(&opt)
	if opt.Format != api.FormatDefault || opt.Bundle {
		t.Fatalf("FormatDefault modified options: %#v", opt)
	}
	if got := formatIDString(t, FormatDefault{}); got != "auto" {
		t.Fatalf("FormatDefault.formatID got %q", got)
	}

	opt = api.BuildOptions{}
	FormatModule{Bundle: true}.Apply(&opt)
	if opt.Format != api.FormatESModule || !opt.Bundle {
		t.Fatalf("FormatModule options: %#v", opt)
	}
	if got := formatIDString(t, FormatModule{Bundle: true}); got != "modulebundle" {
		t.Fatalf("FormatModule.formatID got %q", got)
	}

	opt = api.BuildOptions{}
	FormatCommon{Bundle: true}.Apply(&opt)
	if opt.Format != api.FormatCommonJS || !opt.Bundle {
		t.Fatalf("FormatCommon options: %#v", opt)
	}
	if got := formatIDString(t, FormatCommon{Bundle: true}); got != "commonbundle" {
		t.Fatalf("FormatCommon.formatID got %q", got)
	}

	opt = api.BuildOptions{}
	FormatIIFE{Bundle: true, GlobalName: "App"}.Apply(&opt)
	if opt.Format != api.FormatIIFE || !opt.Bundle || opt.GlobalName != "App" {
		t.Fatalf("FormatIIFE options: %#v", opt)
	}
	if got := formatIDString(t, FormatIIFE{Bundle: true, GlobalName: "App"}); got != "iifebundleApp" {
		t.Fatalf("FormatIIFE.formatID got %q", got)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("FormatRaw.Apply should panic")
		}
	}()
	FormatRaw{}.Apply(&opt)
}

func TestFormatRawID(t *testing.T) {
	if got := formatIDString(t, FormatRaw{}); got != "raw" {
		t.Fatalf("FormatRaw.formatID got %q", got)
	}
}
