// Managed by GoX v0.1.6

package imports

import (
	"io/fs"
	"net/http"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

func staticFiles(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(fileHref()); if __e != nil { return }
		__e = __c.Any(fileRawHref()); if __e != nil { return }
		__e = __c.Any(fileSrc()); if __e != nil { return }
		__e = __c.Any(fileRawSrc()); if __e != nil { return }
	return })
}

func fileHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = doors.AFileHref{
		Path: modulePath + "/style.css",
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.InitVoid("link"); if __e != nil { return }
			{
				__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
}

func fileRawHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = doors.ARawFileHref{
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write(styleRawBytes)
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.InitVoid("link"); if __e != nil { return }
			{
				__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			}
			__e = __c.Submit(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
}

func fileSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = doors.ASrc{
		Path: modulePath + "/index.js",
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("script"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw(""); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
}

func fileRawSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = doors.ARawSrc{
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write(moduleBytes)
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("script"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Raw(""); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
}

func styleBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(doors.Style{
		Source: doors.SourceStyleBytes(styleRawBytes),
	}); if __e != nil { return }
	return })
}

func styleExternalHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(doors.Style{
		Source: doors.SourceExternal(test.Host + "/module/style.css",),
	}); if __e != nil { return }
	return })
}

func styleHostedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(doors.Style{
		Source: doors.SourceLocal("/module/style.css"),
	}); if __e != nil { return }
	return })
}

func styleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(doors.Style{
		Source: doors.SourcePath(modulePath + "/style.css"),
	}); if __e != nil { return }
	return })
}

func moduleExternalHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(doors.ScriptModule{
		Source: doors.SourceExternal(test.Host + "/module/index.js"),
		Specifier: "module",
	}); if __e != nil { return }
	return })
}

func moduleBundleHostHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(doors.ScriptModule{
		Specifier: "module",
		Source: doors.SourceLocal("/module/index.js"),
	}); if __e != nil { return }
	return })
}

func moduleBundleFSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		moduleBundleDir, _ := fs.Sub(moduleBundleFS, "module_bundle_src")

		__e = __c.Any(doors.ScriptModule{
		Specifier: "module",
		Output: doors.ScriptOutputBundle,
		Source: doors.SourceFS{
			FS: moduleBundleDir,
			Path: "index.ts",
			Name: "dd",
		},
	}); if __e != nil { return }
	return })
}

func moduleRawBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(doors.ScriptModule{
		Specifier: "module",
		Source: doors.SourceScriptBytes{
			Content: moduleRawBytes,
		},
		Output: doors.ScriptOutputRaw,
	}); if __e != nil { return }
	return })
}

func moduleRawHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(doors.ScriptModule{
		Specifier: "module",
		Source: doors.SourcePath(modulePath + "/index.js"),
		Output: doors.ScriptOutputRaw,
	}); if __e != nil { return }
	return })
}

func moduleBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(doors.ScriptModule{
		Specifier: "module",
		Source: doors.SourceScriptBytes{
			Content: moduleRawBytes,
		},
	}); if __e != nil { return }
	return })
}

func moduleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrMod(doors.ScriptModule{
			Specifier: "module",
			Source: doors.SourcePath(modulePath + "/index.ts"),
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func reactHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(doors.ScriptModule{
		Specifier: "preact",
		Source: doors.SourcePath(preactPath + "/index.tsx"),
		Output: doors.ScriptOutputBundle,
	}); if __e != nil { return }
		__e = __c.Any(doors.ScriptModule{
		Specifier: "react",
		Source: doors.SourcePath(reactPath + "/index.tsx"),
		Output: doors.ScriptOutputBundle,
	}); if __e != nil { return }
	return })
}

type ModuleFragment struct {
	test.NoBeam
}

func (f *ModuleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "report-0"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("const module = await import(\"module\")\n\t\tdocument.getElementById(\"report-0\").innerHTML = module.test()"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

type ReactFragment struct {
	test.NoBeam
}

func (f *ReactFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "preact"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("const app = await import(\"preact\")\n\t\tapp.init(document.getElementById(\"preact\"))"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "react"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("const app = await import(\"react\")\n\t\tapp.init(document.getElementById(\"react\"))"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

type Empty struct {
	test.NoBeam
}

func (f *Empty) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
	return })
}
