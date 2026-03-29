// Managed by GoX v0.1.20-0.20260329154612-7e48b7c342d5+dirty

//line components.gox:1
package imports

import (
	"io/fs"
	"net/http"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

//line components.gox:12
func staticFiles(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:13
		__e = __c.Any(fileHref()); if __e != nil { return }
//line components.gox:14
		__e = __c.Any(fileSrc()); if __e != nil { return }
//line components.gox:15
		__e = __c.Any(fileRawSrc()); if __e != nil { return }
//line components.gox:16
		__e = __c.Any(fileHrefModify()); if __e != nil { return }
//line components.gox:17
		__e = __c.Any(fileRawHrefModify()); if __e != nil { return }
//line components.gox:18
		__e = __c.Any(fileSrcModify()); if __e != nil { return }
//line components.gox:19
		__e = __c.Any(fileRawSrcModify()); if __e != nil { return }
//line components.gox:20
		__e = __c.Any(fileCachedHref()); if __e != nil { return }
//line components.gox:21
		__e = __c.Any(fileCachedHrefModify()); if __e != nil { return }
//line components.gox:22
		__e = __c.Any(filePrivateHref()); if __e != nil { return }
//line components.gox:23
		__e = __c.Any(filePrivateHrefModify()); if __e != nil { return }
//line components.gox:24
		__e = __c.Any(framePrivateSrc()); if __e != nil { return }
	return })
//line components.gox:25
}

//line components.gox:27
func fileHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:28
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:28
			__e = __c.AttrSet("href", doors.ResourceLocalFS(modulePath + "/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:29
}

//line components.gox:31
func fileRawHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:33
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:34
			__e = __c.AttrSet("href", func(w http.ResponseWriter, r *http.Request) {
			w.Write(styleRawBytes)
		}); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:37
}

//line components.gox:39
func fileSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:40
			__e = __c.AttrSet("src", doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:41
}

//line components.gox:43
func fileRawSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:45
			__e = __c.AttrSet("src", func(w http.ResponseWriter, r *http.Request) {
			w.Write(moduleBytes)
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:48
}

//line components.gox:50
func fileHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:51
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:51
			__e = __c.AttrMod(doors.ResourceLocalFS(modulePath + "/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:52
}

//line components.gox:54
func fileRawHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:56
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:57
			__e = __c.AttrMod(doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write(styleRawBytes)
		})); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:60
}

//line components.gox:62
func fileSrcModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:63
			__e = __c.AttrMod(doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:64
}

//line components.gox:66
func fileRawSrcModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:68
			__e = __c.AttrMod(doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write(moduleBytes)
		})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:71
}

//line components.gox:73
func fileCachedHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:75
			__e = __c.AttrSet("id", "cached-href"); if __e != nil { return }
//line components.gox:76
			__e = __c.AttrSet("href", doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
			__e = __c.AttrSet("cache", true); if __e != nil { return }
//line components.gox:78
			__e = __c.AttrSet("name", "hello.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:81
}

//line components.gox:83
func fileCachedHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:85
			__e = __c.AttrSet("id", "cached-href-modify"); if __e != nil { return }
//line components.gox:86
			__e = __c.AttrMod(doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
			__e = __c.AttrSet("cache", true); if __e != nil { return }
//line components.gox:88
			__e = __c.AttrSet("name", "hello-modify.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:91
}

//line components.gox:93
func filePrivateHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:95
			__e = __c.AttrSet("id", "private-href"); if __e != nil { return }
//line components.gox:96
			__e = __c.AttrSet("href", doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
//line components.gox:97
			__e = __c.AttrSet("name", "private.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:100
}

//line components.gox:102
func filePrivateHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:104
			__e = __c.AttrSet("id", "private-href-modify"); if __e != nil { return }
//line components.gox:105
			__e = __c.AttrMod(doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
//line components.gox:106
			__e = __c.AttrSet("name", "private-modify.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:109
}

//line components.gox:111
func framePrivateSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("iframe"); if __e != nil { return }
		{
//line components.gox:113
			__e = __c.AttrSet("id", "private-frame"); if __e != nil { return }
//line components.gox:114
			__e = __c.AttrSet("src", doors.ResourceString(`<html><body>frame</body></html>`)); if __e != nil { return }
//line components.gox:115
			__e = __c.AttrSet("name", "frame.html"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:117
}

//line components.gox:119
func fileCachedHrefBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:121
			__e = __c.AttrSet("href", doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello"))
		})); if __e != nil { return }
			__e = __c.AttrSet("cache", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:127
}

//line components.gox:129
func styleBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:130
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:130
			__e = __c.AttrSet("href", doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:131
}

//line components.gox:133
func styleInlineHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:139
}

//line components.gox:141
func styleRawHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
//line components.gox:142
			__e = __c.AttrSet("output", "raw"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:147
}

//line components.gox:149
func styleMinifyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
//line components.gox:150
			__e = __c.AttrSet("output", "minify"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:155
}

//line components.gox:157
func styleBytesShortHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:158
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:158
			__e = __c.AttrSet("href", styleRawBytes); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:159
}

//line components.gox:161
func styleBytesModifyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:162
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:162
			__e = __c.AttrMod(doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:163
}

//line components.gox:165
func styleStringHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:166
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:166
			__e = __c.AttrSet("href", doors.ResourceString(string(styleRawBytes))); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:167
}

//line components.gox:169
func styleExternalHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:170
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:170
			__e = __c.AttrSet("href", doors.ResourceExternal(test.Host + "/module/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:171
}

//line components.gox:173
func styleProxyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:174
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:174
			__e = __c.AttrSet("href", doors.ResourceProxy(test.Host + "/module/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:175
}

//line components.gox:177
func styleHostedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:178
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:178
			__e = __c.AttrSet("href", "/module/style.css"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:179
}

//line components.gox:181
func styleHostedRawHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:182
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:182
			__e = __c.AttrSet("href", "/module/style.css"); if __e != nil { return }
//line components.gox:182
			__e = __c.AttrSet("output", "raw"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:183
}

//line components.gox:185
func styleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:186
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:186
			__e = __c.AttrSet("href", doors.ResourceLocalFS(modulePath + "/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:187
}

//line components.gox:189
func styleFSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:191
		moduleDir, _ := fs.Sub(moduleFS, "module_src")

		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:193
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:193
			__e = __c.AttrSet("href", doors.ResourceFS(moduleDir, "style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:194
}

//line components.gox:196
func styleNamedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:197
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:197
			__e = __c.AttrSet("href", doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
//line components.gox:197
			__e = __c.AttrSet("name", "named.css"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:198
}

//line components.gox:200
func stylePrivateNamedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:201
			__e = __c.AttrSet("rel", "stylesheet"); if __e != nil { return }
//line components.gox:201
			__e = __c.AttrSet("href", doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
			__e = __c.AttrSet("private", true); if __e != nil { return }
//line components.gox:201
			__e = __c.AttrSet("name", "private.css"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:202
}

//line components.gox:204
func stylePrivateHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.AttrSet("private", true); if __e != nil { return }
//line components.gox:205
			__e = __c.AttrSet("name", "private-inline"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:210
}

//line components.gox:212
func stylePrivateNamedExtHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.AttrSet("private", true); if __e != nil { return }
//line components.gox:213
			__e = __c.AttrSet("name", "private-inline.css"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:218
}

//line components.gox:220
func styleNoCacheHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.AttrSet("nocache", true); if __e != nil { return }
//line components.gox:221
			__e = __c.AttrSet("name", "nocache-inline"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:226
}

//line components.gox:228
func cspHead(b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:229
		__e = __c.Any(styleExternalHead(b)); if __e != nil { return }
//line components.gox:230
		__e = __c.Any(moduleExternalHead(b)); if __e != nil { return }
	return })
//line components.gox:231
}

//line components.gox:233
func moduleExternalHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:235
			__e = __c.AttrSet("src", doors.ResourceExternal(test.Host + "/module/index.js")); if __e != nil { return }
//line components.gox:236
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:237
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:238
}

//line components.gox:240
func moduleBundleHostHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:242
			__e = __c.AttrSet("src", "/module/index.js"); if __e != nil { return }
//line components.gox:243
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:244
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:245
}

//line components.gox:247
func moduleBundleFSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:249
		moduleBundleDir, _ := fs.Sub(moduleBundleFS, "module_bundle_src")

		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:252
			__e = __c.AttrSet("src", doors.ResourceFS(moduleBundleDir, "index.ts")); if __e != nil { return }
//line components.gox:253
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:254
			__e = __c.AttrSet("output", "bundle"); if __e != nil { return }
//line components.gox:255
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:256
}

//line components.gox:258
func moduleRawBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:260
			__e = __c.AttrSet("src", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:261
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:262
			__e = __c.AttrSet("output", "raw"); if __e != nil { return }
//line components.gox:263
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:264
}

//line components.gox:266
func moduleRawBytesShortHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:268
			__e = __c.AttrSet("src", moduleRawBytes); if __e != nil { return }
//line components.gox:269
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:270
			__e = __c.AttrSet("output", "raw"); if __e != nil { return }
//line components.gox:271
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:272
}

//line components.gox:274
func moduleRawBytesModifyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:276
			__e = __c.AttrMod(doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:277
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:278
			__e = __c.AttrSet("output", "raw"); if __e != nil { return }
//line components.gox:279
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:280
}

//line components.gox:282
func modulePreloadBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:284
			__e = __c.AttrSet("rel", "modulepreload"); if __e != nil { return }
//line components.gox:285
			__e = __c.AttrSet("href", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:286
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:287
}

//line components.gox:289
func moduleRawHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:291
			__e = __c.AttrSet("src", doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
//line components.gox:292
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:293
			__e = __c.AttrSet("output", "raw"); if __e != nil { return }
//line components.gox:294
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:295
}

//line components.gox:297
func moduleBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:299
			__e = __c.AttrSet("src", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:300
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:301
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:302
}

//line components.gox:304
func moduleStringHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:306
			__e = __c.AttrSet("src", doors.ResourceString(string(moduleRawBytes))); if __e != nil { return }
//line components.gox:307
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:308
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:309
}

//line components.gox:311
func moduleProxyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:313
			__e = __c.AttrSet("src", doors.ResourceProxy(test.Host + "/module/index.js")); if __e != nil { return }
//line components.gox:314
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:315
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:316
}

//line components.gox:318
func moduleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:320
			__e = __c.AttrSet("src", doors.ResourceLocalFS(modulePath + "/index.ts")); if __e != nil { return }
//line components.gox:321
			__e = __c.AttrSet("name", "module.js"); if __e != nil { return }
//line components.gox:322
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:323
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:324
}

//line components.gox:326
func moduleVisibleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:328
			__e = __c.AttrSet("src", doors.ResourceLocalFS(modulePath + "/index.ts")); if __e != nil { return }
//line components.gox:329
			__e = __c.AttrSet("id", "module-tag"); if __e != nil { return }
//line components.gox:330
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:331
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:332
}

//line components.gox:334
func modulePreloadNamedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:336
			__e = __c.AttrSet("rel", "modulepreload"); if __e != nil { return }
//line components.gox:337
			__e = __c.AttrSet("href", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:338
			__e = __c.AttrSet("name", "module-preload.js"); if __e != nil { return }
//line components.gox:339
			__e = __c.AttrSet("specifier", "module"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:340
}

//line components.gox:342
func scriptInlineHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("window.__importsValue = \"hello\""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:346
}

//line components.gox:348
func scriptInlineNamedExtHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:349
			__e = __c.AttrSet("id", "script-inline-ext"); if __e != nil { return }
//line components.gox:349
			__e = __c.AttrSet("name", "inline-script.js"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("window.__importsValue = \"hello\""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:352
}

//line components.gox:354
func scriptInlineBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:356
			__e = __c.AttrSet("src", doors.ResourceBytes([]byte(`window.__importsValue = "hello"`))); if __e != nil { return }
//line components.gox:357
			__e = __c.AttrSet("output", "inline"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:358
}

//line components.gox:360
func scriptStringHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:361
			__e = __c.AttrSet("src", doors.ResourceString(`window.__importsValue = "hello"`)); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:362
}

//line components.gox:364
func scriptPrivateHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:366
			__e = __c.AttrSet("id", "script-private"); if __e != nil { return }
//line components.gox:367
			__e = __c.AttrSet("src", doors.ResourceString(`window.__importsValue = "hello"`)); if __e != nil { return }
			__e = __c.AttrSet("private", true); if __e != nil { return }
//line components.gox:369
			__e = __c.AttrSet("name", "private-script.js"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:370
}

//line components.gox:372
func scriptNoCacheHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:374
			__e = __c.AttrSet("id", "script-nocache"); if __e != nil { return }
//line components.gox:375
			__e = __c.AttrSet("src", doors.ResourceString(`window.__importsValue = "hello"`)); if __e != nil { return }
			__e = __c.AttrSet("nocache", true); if __e != nil { return }
//line components.gox:377
			__e = __c.AttrSet("name", "nocache-script.js"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:378
}

//line components.gox:380
func reactHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:382
			__e = __c.AttrSet("src", doors.ResourceLocalFS(preactPath + "/index.tsx")); if __e != nil { return }
//line components.gox:383
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:384
			__e = __c.AttrSet("output", "bundle"); if __e != nil { return }
//line components.gox:385
			__e = __c.AttrSet("specifier", "preact"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:387
			__e = __c.AttrSet("src", doors.ResourceLocalFS(reactPath + "/index.tsx")); if __e != nil { return }
//line components.gox:388
			__e = __c.AttrSet("type", "module"); if __e != nil { return }
//line components.gox:389
			__e = __c.AttrSet("output", "bundle"); if __e != nil { return }
//line components.gox:390
			__e = __c.AttrSet("specifier", "react"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:391
}

type ModuleFragment struct {
	test.NoBeam
}

//line components.gox:397
func (f *ModuleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:398
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
//line components.gox:403
}

type ReactFragment struct {
	test.NoBeam
}

//line components.gox:409
func (f *ReactFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:410
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
//line components.gox:415
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
//line components.gox:420
}

type ValueFragment struct {
	test.NoBeam
}

//line components.gox:426
func (f *ValueFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:427
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
//line components.gox:431
}

type Empty struct {
	test.NoBeam
}

//line components.gox:437
func (f *Empty) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
	return })
//line components.gox:437
}
