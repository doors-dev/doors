// Managed by GoX v0.1.17+dirty

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
		__e = __c.Any(fileSrc()); if __e != nil { return }
		__e = __c.Any(fileRawSrc()); if __e != nil { return }
		__e = __c.Any(fileHrefModify()); if __e != nil { return }
		__e = __c.Any(fileRawHrefModify()); if __e != nil { return }
		__e = __c.Any(fileSrcModify()); if __e != nil { return }
		__e = __c.Any(fileRawSrcModify()); if __e != nil { return }
		__e = __c.Any(fileCachedHref()); if __e != nil { return }
		__e = __c.Any(fileCachedHrefModify()); if __e != nil { return }
	return })
}

func fileHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrSet("href", doors.ResourceLocalFS(modulePath + "/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func fileRawHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrSet("href", func(w http.ResponseWriter, r *http.Request) {
		w.Write(styleRawBytes)
	}); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func fileSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func fileRawSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", func(w http.ResponseWriter, r *http.Request) {
		w.Write(moduleBytes)
	}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func fileHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrMod(doors.ResourceLocalFS(modulePath + "/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func fileRawHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrMod(doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Write(styleRawBytes)
	})); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func fileSrcModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrMod(doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func fileRawSrcModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrMod(doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Write(moduleBytes)
	})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func fileCachedHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("a"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "cached-href"); if __e != nil { return }
			__e = __c.AttrSet("href", doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
			__e = __c.AttrSet("cache", true); if __e != nil { return }
			__e = __c.AttrSet("name", "hello.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func fileCachedHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("a"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "cached-href-modify"); if __e != nil { return }
			__e = __c.AttrMod(doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
			__e = __c.AttrSet("cache", true); if __e != nil { return }
			__e = __c.AttrSet("name", "hello-modify.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func fileCachedHrefBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("a"); if __e != nil { return }
		{
			__e = __c.AttrSet("href", doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello"))
		})); if __e != nil { return }
			__e = __c.AttrSet("cache", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func styleBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrSet("href", doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func styleInlineHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func styleBytesShortHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrSet("href", styleRawBytes); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func styleBytesModifyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrMod(doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func styleStringHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrSet("href", doors.ResourceString(string(styleRawBytes))); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func styleExternalHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrSet("href", doors.ResourceExternal(test.Host + "/module/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func styleProxyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrSet("href", doors.ResourceProxy(test.Host + "/module/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func styleHostedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrSet("href", "/module/style.css"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func styleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrSet("href", doors.ResourceLocalFS(modulePath + "/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func styleFSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		moduleDir, _ := fs.Sub(moduleFS, "module_src")

		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrSet("href", doors.ResourceFS(moduleDir, "style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func styleNamedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrSet("href", doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
			__e = __c.AttrSet("name", "named.css"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func stylePrivateNamedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
			__e = __c.AttrSet("href", doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
			__e = __c.AttrSet("private", true); if __e != nil { return }
			__e = __c.AttrSet("name", "private.css"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func moduleExternalHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceExternal(test.Host + "/module/index.js")); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func moduleBundleHostHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", "/module/index.js"); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func moduleBundleFSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		moduleBundleDir, _ := fs.Sub(moduleBundleFS, "module_bundle_src")

		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceFS(moduleBundleDir, "index.ts")); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("output", "bundle"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func moduleRawBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("output", "raw"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func moduleRawBytesShortHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", moduleRawBytes); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("output", "raw"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func moduleRawBytesModifyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrMod(doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("output", "raw"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func modulePreloadBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "modulepreload"); if __e != nil { return }
			__e = __c.AttrSet("href", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func moduleRawHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("output", "raw"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func moduleBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func moduleStringHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceString(string(moduleRawBytes))); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func moduleProxyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceProxy(test.Host + "/module/index.js")); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func moduleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceLocalFS(modulePath + "/index.ts")); if __e != nil { return }
			__e = __c.AttrSet("name", "module.js"); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func moduleVisibleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceLocalFS(modulePath + "/index.ts")); if __e != nil { return }
			__e = __c.AttrSet("id", "module-tag"); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func modulePreloadNamedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
			__e = __c.AttrSet("rel", "modulepreload"); if __e != nil { return }
			__e = __c.AttrSet("href", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
			__e = __c.AttrSet("name", "module-preload.js"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
}

func scriptInlineHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("window.__importsValue = \"hello\""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func scriptStringHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceString(`window.__importsValue = "hello"`)); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func reactHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceLocalFS(preactPath + "/index.tsx")); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("output", "bundle"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "preact"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.AttrSet("src", doors.ResourceLocalFS(reactPath + "/index.tsx")); if __e != nil { return }
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
			__e = __c.AttrSet("output", "bundle"); if __e != nil { return }
			__e = __c.AttrSet("specifier", "react"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
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

type ValueFragment struct {
	test.NoBeam
}

func (f *ValueFragment) Main() gox.Elem {
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
			__e = __c.Raw("document.getElementById(\"report-0\").innerHTML = window.__importsValue"); if __e != nil { return }
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
