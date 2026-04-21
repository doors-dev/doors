// Managed by GoX v0.1.28

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
		__e = __c.Any(fileRawHref()); if __e != nil { return }
//line components.gox:15
		__e = __c.Any(fileSrc()); if __e != nil { return }
//line components.gox:16
		__e = __c.Any(fileRawSrc()); if __e != nil { return }
//line components.gox:17
		__e = __c.Any(fileHrefModify()); if __e != nil { return }
//line components.gox:18
		__e = __c.Any(fileRawHrefModify()); if __e != nil { return }
//line components.gox:19
		__e = __c.Any(fileSrcModify()); if __e != nil { return }
//line components.gox:20
		__e = __c.Any(fileRawSrcModify()); if __e != nil { return }
//line components.gox:21
		__e = __c.Any(fileImgFS()); if __e != nil { return }
//line components.gox:22
		__e = __c.Any(fileCachedHref()); if __e != nil { return }
//line components.gox:23
		__e = __c.Any(fileCachedHrefModify()); if __e != nil { return }
//line components.gox:24
		__e = __c.Any(filePrivateHref()); if __e != nil { return }
//line components.gox:25
		__e = __c.Any(filePrivateHrefModify()); if __e != nil { return }
//line components.gox:26
		__e = __c.Any(framePrivateSrc()); if __e != nil { return }
	return })
//line components.gox:27
}

//line components.gox:29
func fileHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:30
			__e = __c.Set("id", "file-href"); if __e != nil { return }
//line components.gox:30
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:30
			__e = __c.Set("href", doors.ResourceLocalFS(modulePath + "/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:31
}

//line components.gox:33
func fileRawHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:35
			__e = __c.Set("id", "file-raw-href"); if __e != nil { return }
//line components.gox:36
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:37
			__e = __c.Set("href", func(w http.ResponseWriter, r *http.Request) {
			w.Write(styleRawBytes)
		}); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:40
}

//line components.gox:42
func fileSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:43
			__e = __c.Set("id", "file-src"); if __e != nil { return }
//line components.gox:43
			__e = __c.Set("src", doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:44
}

//line components.gox:46
func fileRawSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:48
			__e = __c.Set("id", "file-raw-src"); if __e != nil { return }
//line components.gox:49
			__e = __c.Set("src", func(w http.ResponseWriter, r *http.Request) {
			w.Write(moduleBytes)
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:52
}

//line components.gox:54
func fileHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:55
			__e = __c.Set("id", "file-href-modify"); if __e != nil { return }
//line components.gox:55
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:55
			__e = __c.Modify(doors.ResourceLocalFS(modulePath + "/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:56
}

//line components.gox:58
func fileRawHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:60
			__e = __c.Set("id", "file-raw-href-modify"); if __e != nil { return }
//line components.gox:61
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:62
			__e = __c.Modify(doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write(styleRawBytes)
		})); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:65
}

//line components.gox:67
func fileSrcModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:68
			__e = __c.Set("id", "file-src-modify"); if __e != nil { return }
//line components.gox:68
			__e = __c.Modify(doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:69
}

//line components.gox:71
func fileRawSrcModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:73
			__e = __c.Set("id", "file-raw-src-modify"); if __e != nil { return }
//line components.gox:74
			__e = __c.Modify(doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write(moduleBytes)
		})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:77
}

//line components.gox:79
func fileImgFS() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:81
		moduleDir, _ := fs.Sub(moduleFS, "module_src")

		__e = __c.InitVoid("img"); if __e != nil { return }
		{
//line components.gox:83
			__e = __c.Set("id", "file-img-fs"); if __e != nil { return }
//line components.gox:83
			__e = __c.Set("src", doors.ResourceFS(moduleDir, "pixel.svg")); if __e != nil { return }
//line components.gox:83
			__e = __c.Set("type", "image/svg+xml"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:84
}

//line components.gox:86
func fileCachedHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:88
			__e = __c.Set("id", "cached-href"); if __e != nil { return }
//line components.gox:89
			__e = __c.Set("href", doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
			__e = __c.Set("cache", true); if __e != nil { return }
//line components.gox:91
			__e = __c.Set("name", "hello.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:94
}

//line components.gox:96
func fileCachedHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:98
			__e = __c.Set("id", "cached-href-modify"); if __e != nil { return }
//line components.gox:99
			__e = __c.Modify(doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
			__e = __c.Set("cache", true); if __e != nil { return }
//line components.gox:101
			__e = __c.Set("name", "hello-modify.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:104
}

//line components.gox:106
func filePrivateHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:108
			__e = __c.Set("id", "private-href"); if __e != nil { return }
//line components.gox:109
			__e = __c.Set("href", doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
//line components.gox:110
			__e = __c.Set("name", "private.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:113
}

//line components.gox:115
func filePrivateHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:117
			__e = __c.Set("id", "private-href-modify"); if __e != nil { return }
//line components.gox:118
			__e = __c.Modify(doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
//line components.gox:119
			__e = __c.Set("name", "private-modify.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:122
}

//line components.gox:124
func framePrivateSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("iframe"); if __e != nil { return }
		{
//line components.gox:126
			__e = __c.Set("id", "private-frame"); if __e != nil { return }
//line components.gox:127
			__e = __c.Set("src", doors.ResourceString(`<html><body>frame</body></html>`)); if __e != nil { return }
//line components.gox:128
			__e = __c.Set("name", "frame.html"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:130
}

//line components.gox:132
func fileCachedHrefBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:134
			__e = __c.Set("href", doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello"))
		})); if __e != nil { return }
			__e = __c.Set("cache", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:140
}

//line components.gox:142
func styleBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:143
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:143
			__e = __c.Set("href", doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:144
}

//line components.gox:146
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
//line components.gox:152
}

//line components.gox:154
func styleRawHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.Set("raw", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:160
}

//line components.gox:162
func styleMinifyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:168
}

//line components.gox:170
func styleBytesShortHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:171
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:171
			__e = __c.Set("href", styleRawBytes); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:172
}

//line components.gox:174
func styleBytesModifyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:175
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:175
			__e = __c.Modify(doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:176
}

//line components.gox:178
func styleStringHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:179
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:179
			__e = __c.Set("href", doors.ResourceString(string(styleRawBytes))); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:180
}

//line components.gox:182
func styleExternalHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:183
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:183
			__e = __c.Set("href", doors.ResourceExternal(test.Host + "/module/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:184
}

//line components.gox:186
func styleProxyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:187
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:187
			__e = __c.Set("href", doors.ResourceProxy(test.Host + "/module/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:188
}

//line components.gox:190
func styleHostedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:191
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:191
			__e = __c.Set("href", "/module/style.css"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:192
}

//line components.gox:194
func styleHostedRawHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:195
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:195
			__e = __c.Set("href", "/module/style.css"); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:196
}

//line components.gox:198
func styleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:199
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:199
			__e = __c.Set("href", doors.ResourceLocalFS(modulePath + "/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:200
}

//line components.gox:202
func styleFSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:204
		moduleDir, _ := fs.Sub(moduleFS, "module_src")

		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:206
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:206
			__e = __c.Set("href", doors.ResourceFS(moduleDir, "style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:207
}

//line components.gox:209
func styleNamedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:210
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:210
			__e = __c.Set("href", doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
//line components.gox:210
			__e = __c.Set("name", "named.css"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:211
}

//line components.gox:213
func stylePrivateNamedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:214
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:214
			__e = __c.Set("href", doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
			__e = __c.Set("private", true); if __e != nil { return }
//line components.gox:214
			__e = __c.Set("name", "private.css"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:215
}

//line components.gox:217
func stylePrivateHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.Set("private", true); if __e != nil { return }
//line components.gox:218
			__e = __c.Set("name", "private-inline"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:223
}

//line components.gox:225
func stylePrivateNamedExtHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.Set("private", true); if __e != nil { return }
//line components.gox:226
			__e = __c.Set("name", "private-inline.css"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:231
}

//line components.gox:233
func styleNoCacheHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.Set("nocache", true); if __e != nil { return }
//line components.gox:234
			__e = __c.Set("name", "nocache-inline"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:239
}

//line components.gox:241
func cspHead(b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:242
		__e = __c.Any(styleExternalHead(b)); if __e != nil { return }
//line components.gox:243
		__e = __c.Any(moduleExternalHead(b)); if __e != nil { return }
	return })
//line components.gox:244
}

//line components.gox:246
func moduleExternalHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:248
			__e = __c.Set("src", doors.ResourceExternal(test.Host + "/module/index.js")); if __e != nil { return }
//line components.gox:249
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:250
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:251
}

//line components.gox:253
func moduleBundleHostHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:255
			__e = __c.Set("src", "/module/index.js"); if __e != nil { return }
//line components.gox:256
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:257
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:258
}

//line components.gox:260
func moduleBundleFSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:262
		moduleBundleDir, _ := fs.Sub(moduleBundleFS, "module_bundle_src")

		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:265
			__e = __c.Set("id", "module-bundle-fs"); if __e != nil { return }
//line components.gox:266
			__e = __c.Set("src", doors.ResourceFS(moduleBundleDir, "index.ts")); if __e != nil { return }
//line components.gox:267
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("bundle", true); if __e != nil { return }
//line components.gox:269
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:270
}

//line components.gox:272
func moduleRawBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:274
			__e = __c.Set("id", "module-raw-bytes"); if __e != nil { return }
//line components.gox:275
			__e = __c.Set("src", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:276
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
//line components.gox:278
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:279
}

//line components.gox:281
func moduleRawBytesShortHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:283
			__e = __c.Set("id", "module-raw-bytes-short"); if __e != nil { return }
//line components.gox:284
			__e = __c.Set("src", moduleRawBytes); if __e != nil { return }
//line components.gox:285
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
//line components.gox:287
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:288
}

//line components.gox:290
func moduleRawBytesModifyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:292
			__e = __c.Set("id", "module-raw-bytes-modify"); if __e != nil { return }
//line components.gox:293
			__e = __c.Modify(doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:294
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
//line components.gox:296
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:297
}

//line components.gox:299
func modulePreloadBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:301
			__e = __c.Set("rel", "modulepreload"); if __e != nil { return }
//line components.gox:302
			__e = __c.Set("href", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:303
			__e = __c.Set("specifier", "module"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:304
}

//line components.gox:306
func moduleRawHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:308
			__e = __c.Set("id", "module-raw"); if __e != nil { return }
//line components.gox:309
			__e = __c.Set("src", doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
//line components.gox:310
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
//line components.gox:312
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:313
}

//line components.gox:315
func moduleBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:317
			__e = __c.Set("id", "module-bytes"); if __e != nil { return }
//line components.gox:318
			__e = __c.Set("src", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:319
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:320
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:321
}

//line components.gox:323
func moduleStringHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:325
			__e = __c.Set("id", "module-string"); if __e != nil { return }
//line components.gox:326
			__e = __c.Set("src", doors.ResourceString(string(moduleRawBytes))); if __e != nil { return }
//line components.gox:327
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:328
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:329
}

//line components.gox:331
func moduleProxyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:333
			__e = __c.Set("src", doors.ResourceProxy(test.Host + "/module/index.js")); if __e != nil { return }
//line components.gox:334
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:335
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:336
}

//line components.gox:338
func moduleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:340
			__e = __c.Set("src", doors.ResourceLocalFS(modulePath + "/index.ts")); if __e != nil { return }
//line components.gox:341
			__e = __c.Set("name", "module.js"); if __e != nil { return }
//line components.gox:342
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:343
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:344
}

//line components.gox:346
func moduleVisibleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:348
			__e = __c.Set("src", doors.ResourceLocalFS(modulePath + "/index.ts")); if __e != nil { return }
//line components.gox:349
			__e = __c.Set("id", "module-tag"); if __e != nil { return }
//line components.gox:350
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:351
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:352
}

//line components.gox:354
func modulePreloadNamedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:356
			__e = __c.Set("rel", "modulepreload"); if __e != nil { return }
//line components.gox:357
			__e = __c.Set("href", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:358
			__e = __c.Set("name", "module-preload.js"); if __e != nil { return }
//line components.gox:359
			__e = __c.Set("specifier", "module"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:360
}

//line components.gox:362
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
//line components.gox:366
}

//line components.gox:368
func scriptRawHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:369
			__e = __c.Set("id", "script-raw-inline"); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("window.__importsValue = \"hello\""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:372
}

//line components.gox:374
func scriptInlineNamedExtHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:375
			__e = __c.Set("id", "script-inline-ext"); if __e != nil { return }
//line components.gox:375
			__e = __c.Set("name", "inline-script.js"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("window.__importsValue = \"hello\""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:378
}

//line components.gox:380
func scriptInlineBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:382
			__e = __c.Set("id", "script-inline-bytes"); if __e != nil { return }
//line components.gox:383
			__e = __c.Set("src", doors.ResourceBytes([]byte(`window.__importsValue = "hello"`))); if __e != nil { return }
			__e = __c.Set("inline", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:385
}

//line components.gox:387
func scriptStringHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:388
			__e = __c.Set("src", doors.ResourceString(`window.__importsValue = "hello"`)); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:389
}

//line components.gox:391
func scriptPrivateHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:393
			__e = __c.Set("id", "script-private"); if __e != nil { return }
//line components.gox:394
			__e = __c.Set("src", doors.ResourceString(`window.__importsValue = "hello"`)); if __e != nil { return }
			__e = __c.Set("private", true); if __e != nil { return }
//line components.gox:396
			__e = __c.Set("name", "private-script.js"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:397
}

//line components.gox:399
func scriptNoCacheHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:401
			__e = __c.Set("id", "script-nocache"); if __e != nil { return }
//line components.gox:402
			__e = __c.Set("src", doors.ResourceString(`window.__importsValue = "hello"`)); if __e != nil { return }
			__e = __c.Set("nocache", true); if __e != nil { return }
//line components.gox:404
			__e = __c.Set("name", "nocache-script.js"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:405
}

//line components.gox:407
func reactHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:409
			__e = __c.Set("id", "preact-bundle"); if __e != nil { return }
//line components.gox:410
			__e = __c.Set("src", doors.ResourceLocalFS(preactPath + "/index.tsx")); if __e != nil { return }
//line components.gox:411
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("bundle", true); if __e != nil { return }
//line components.gox:413
			__e = __c.Set("specifier", "preact"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:415
			__e = __c.Set("id", "react-bundle"); if __e != nil { return }
//line components.gox:416
			__e = __c.Set("src", doors.ResourceLocalFS(reactPath + "/index.tsx")); if __e != nil { return }
//line components.gox:417
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("bundle", true); if __e != nil { return }
//line components.gox:419
			__e = __c.Set("specifier", "react"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:420
}

//line components.gox:422
func moduleTypeTSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:424
			__e = __c.Set("src", doors.ResourceLocalFS(modulePath + "/index.ts")); if __e != nil { return }
//line components.gox:425
			__e = __c.Set("type", "module/typescript"); if __e != nil { return }
//line components.gox:426
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:427
}

//line components.gox:429
func moduleTypeJSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:431
			__e = __c.Set("src", doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
//line components.gox:432
			__e = __c.Set("type", "module/javascript"); if __e != nil { return }
//line components.gox:433
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:434
}

//line components.gox:436
func fileHandlerTypeBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("img"); if __e != nil { return }
		{
//line components.gox:438
			__e = __c.Set("src", doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("bad"))
		})); if __e != nil { return }
//line components.gox:441
			__e = __c.Set("type", "text/plain"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:442
}

//line components.gox:444
func scriptDuplicateOutputBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:445
			__e = __c.Set("src", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
			__e = __c.Set("bundle", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:446
}

//line components.gox:448
func scriptHandlerBundleBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:450
			__e = __c.Set("src", doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write(moduleRawBytes)
		})); if __e != nil { return }
			__e = __c.Set("bundle", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:454
}

//line components.gox:456
func scriptRawTSBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:458
			__e = __c.Set("src", doors.ResourceBytes(moduleBytes)); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
//line components.gox:460
			__e = __c.Set("type", "text/typescript"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:461
}

//line components.gox:463
func scriptSpecifierNonModuleBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:465
			__e = __c.Set("src", doors.ResourceBytes(moduleBytes)); if __e != nil { return }
//line components.gox:466
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:467
}

//line components.gox:469
func scriptInlineModuleBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:470
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("window.__importsValue = \"bad\""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:473
}

//line components.gox:475
func scriptInlineTSBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:476
			__e = __c.Set("type", "text/typescript"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("const value: string = \"bad\"\n\t\twindow.__importsValue = value"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:480
}

//line components.gox:482
func scriptDirectBundleBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:483
			__e = __c.Set("src", "/module/index.js"); if __e != nil { return }
			__e = __c.Set("bundle", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:484
}

//line components.gox:486
func scriptHandlerInlineBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:488
			__e = __c.Set("src", doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write(moduleRawBytes)
		})); if __e != nil { return }
			__e = __c.Set("inline", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:492
}

//line components.gox:494
func modulePreloadInlineBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:495
			__e = __c.Set("rel", "modulepreload"); if __e != nil { return }
//line components.gox:495
			__e = __c.Set("href", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
			__e = __c.Set("inline", true); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:496
}

//line components.gox:498
func styleHandlerPrivateBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:500
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:501
			__e = __c.Set("href", doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write(styleRawBytes)
		})); if __e != nil { return }
			__e = __c.Set("private", true); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:505
}

//line components.gox:507
func styleDirectPrivateBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:509
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:510
			__e = __c.Set("href", "/module/style.css"); if __e != nil { return }
			__e = __c.Set("private", true); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:512
}

type ModuleFragment struct {
	test.NoBeam
}

//line components.gox:518
func (f *ModuleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:519
			__e = __c.Set("id", "report-0"); if __e != nil { return }
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
//line components.gox:524
}

type ReactFragment struct {
	test.NoBeam
}

//line components.gox:530
func (f *ReactFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:531
			__e = __c.Set("id", "preact"); if __e != nil { return }
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
//line components.gox:536
			__e = __c.Set("id", "react"); if __e != nil { return }
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
//line components.gox:541
}

type ValueFragment struct {
	test.NoBeam
}

//line components.gox:547
func (f *ValueFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:548
			__e = __c.Set("id", "report-0"); if __e != nil { return }
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
//line components.gox:552
}

type Empty struct {
	test.NoBeam
}

//line components.gox:558
func (f *Empty) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
	return })
//line components.gox:558
}
