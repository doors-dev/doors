// Managed by GoX v0.3.2

//line components.gox:1
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

package imports

import (
	"io/fs"
	"net/http"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

//line components.gox:26
func staticFiles(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:27
		__e = __c.Any(fileHref()); if __e != nil { return }
//line components.gox:28
		__e = __c.Any(fileRawHref()); if __e != nil { return }
//line components.gox:29
		__e = __c.Any(fileSrc()); if __e != nil { return }
//line components.gox:30
		__e = __c.Any(fileRawSrc()); if __e != nil { return }
//line components.gox:31
		__e = __c.Any(fileHrefModify()); if __e != nil { return }
//line components.gox:32
		__e = __c.Any(fileRawHrefModify()); if __e != nil { return }
//line components.gox:33
		__e = __c.Any(fileSrcModify()); if __e != nil { return }
//line components.gox:34
		__e = __c.Any(fileRawSrcModify()); if __e != nil { return }
//line components.gox:35
		__e = __c.Any(fileImgFS()); if __e != nil { return }
//line components.gox:36
		__e = __c.Any(fileCachedHref()); if __e != nil { return }
//line components.gox:37
		__e = __c.Any(fileCachedHrefModify()); if __e != nil { return }
//line components.gox:38
		__e = __c.Any(filePrivateHref()); if __e != nil { return }
//line components.gox:39
		__e = __c.Any(filePrivateHrefModify()); if __e != nil { return }
//line components.gox:40
		__e = __c.Any(framePrivateSrc()); if __e != nil { return }
	return })
//line components.gox:41
}

//line components.gox:43
func fileHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:44
			__e = __c.Set("id", "file-href"); if __e != nil { return }
//line components.gox:44
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:44
			__e = __c.Set("href", doors.ResourceLocalFS(modulePath + "/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:45
}

//line components.gox:47
func fileRawHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:49
			__e = __c.Set("id", "file-raw-href"); if __e != nil { return }
//line components.gox:50
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:51
			__e = __c.Set("href", func(w http.ResponseWriter, r *http.Request) {
			w.Write(styleRawBytes)
		}); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:54
}

//line components.gox:56
func fileSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:57
			__e = __c.Set("id", "file-src"); if __e != nil { return }
//line components.gox:57
			__e = __c.Set("src", doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:58
}

//line components.gox:60
func fileRawSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:62
			__e = __c.Set("id", "file-raw-src"); if __e != nil { return }
//line components.gox:63
			__e = __c.Set("src", func(w http.ResponseWriter, r *http.Request) {
			w.Write(moduleBytes)
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:66
}

//line components.gox:68
func fileHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:69
			__e = __c.Set("id", "file-href-modify"); if __e != nil { return }
//line components.gox:69
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:69
			__e = __c.Modify(doors.ResourceLocalFS(modulePath + "/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:70
}

//line components.gox:72
func fileRawHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:74
			__e = __c.Set("id", "file-raw-href-modify"); if __e != nil { return }
//line components.gox:75
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:76
			__e = __c.Modify(doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write(styleRawBytes)
		})); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:79
}

//line components.gox:81
func fileSrcModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:82
			__e = __c.Set("id", "file-src-modify"); if __e != nil { return }
//line components.gox:82
			__e = __c.Modify(doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:83
}

//line components.gox:85
func fileRawSrcModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:87
			__e = __c.Set("id", "file-raw-src-modify"); if __e != nil { return }
//line components.gox:88
			__e = __c.Modify(doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write(moduleBytes)
		})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:91
}

//line components.gox:93
func fileImgFS() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:95
		moduleDir, _ := fs.Sub(moduleFS, "module_src")

		__e = __c.InitVoid("img"); if __e != nil { return }
		{
//line components.gox:97
			__e = __c.Set("id", "file-img-fs"); if __e != nil { return }
//line components.gox:97
			__e = __c.Set("src", doors.ResourceFS(moduleDir, "pixel.svg")); if __e != nil { return }
//line components.gox:97
			__e = __c.Set("type", "image/svg+xml"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:98
}

//line components.gox:100
func fileCachedHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:102
			__e = __c.Set("id", "cached-href"); if __e != nil { return }
//line components.gox:103
			__e = __c.Set("href", doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
			__e = __c.Set("cache", true); if __e != nil { return }
//line components.gox:105
			__e = __c.Set("name", "hello.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:108
}

//line components.gox:110
func fileCachedHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:112
			__e = __c.Set("id", "cached-href-modify"); if __e != nil { return }
//line components.gox:113
			__e = __c.Modify(doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
			__e = __c.Set("cache", true); if __e != nil { return }
//line components.gox:115
			__e = __c.Set("name", "hello-modify.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:118
}

//line components.gox:120
func filePrivateHref() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:122
			__e = __c.Set("id", "private-href"); if __e != nil { return }
//line components.gox:123
			__e = __c.Set("href", doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
//line components.gox:124
			__e = __c.Set("name", "private.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:127
}

//line components.gox:129
func filePrivateHrefModify() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:131
			__e = __c.Set("id", "private-href-modify"); if __e != nil { return }
//line components.gox:132
			__e = __c.Modify(doors.ResourceBytes([]byte("hello"))); if __e != nil { return }
//line components.gox:133
			__e = __c.Set("name", "private-modify.txt"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:136
}

//line components.gox:138
func framePrivateSrc() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("iframe"); if __e != nil { return }
		{
//line components.gox:140
			__e = __c.Set("id", "private-frame"); if __e != nil { return }
//line components.gox:141
			__e = __c.Set("src", doors.ResourceString(`<html><body>frame</body></html>`)); if __e != nil { return }
//line components.gox:142
			__e = __c.Set("name", "frame.html"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:144
}

//line components.gox:146
func fileCachedHrefBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("a"); if __e != nil { return }
		{
//line components.gox:148
			__e = __c.Set("href", doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello"))
		})); if __e != nil { return }
			__e = __c.Set("cache", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Download"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:154
}

//line components.gox:156
func styleBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:157
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:157
			__e = __c.Set("href", doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:158
}

//line components.gox:160
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
//line components.gox:166
}

//line components.gox:168
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
//line components.gox:174
}

//line components.gox:176
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
//line components.gox:182
}

//line components.gox:184
func styleBytesShortHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:185
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:185
			__e = __c.Set("href", styleRawBytes); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:186
}

//line components.gox:188
func styleBytesModifyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:189
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:189
			__e = __c.Modify(doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:190
}

//line components.gox:192
func styleStringHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:193
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:193
			__e = __c.Set("href", doors.ResourceString(string(styleRawBytes))); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:194
}

//line components.gox:196
func styleExternalHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:197
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:197
			__e = __c.Set("href", doors.ResourceExternal(test.Host + "/module/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:198
}

//line components.gox:200
func styleProxyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:201
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:201
			__e = __c.Set("href", doors.ResourceProxy(test.Host + "/module/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:202
}

//line components.gox:204
func styleHostedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:205
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:205
			__e = __c.Set("href", "/module/style.css"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:206
}

//line components.gox:208
func styleHostedRawHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:209
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:209
			__e = __c.Set("href", "/module/style.css"); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:210
}

//line components.gox:212
func styleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:213
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:213
			__e = __c.Set("href", doors.ResourceLocalFS(modulePath + "/style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:214
}

//line components.gox:216
func styleFSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:218
		moduleDir, _ := fs.Sub(moduleFS, "module_src")

		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:220
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:220
			__e = __c.Set("href", doors.ResourceFS(moduleDir, "style.css")); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:221
}

//line components.gox:223
func styleNamedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:224
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:224
			__e = __c.Set("href", doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
//line components.gox:224
			__e = __c.Set("name", "named.css"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:225
}

//line components.gox:227
func stylePrivateNamedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:228
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:228
			__e = __c.Set("href", doors.ResourceBytes(styleRawBytes)); if __e != nil { return }
			__e = __c.Set("private", true); if __e != nil { return }
//line components.gox:228
			__e = __c.Set("name", "private.css"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:229
}

//line components.gox:231
func stylePrivateHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.Set("private", true); if __e != nil { return }
//line components.gox:232
			__e = __c.Set("name", "private-inline"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:237
}

//line components.gox:239
func stylePrivateNamedExtHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.Set("private", true); if __e != nil { return }
//line components.gox:240
			__e = __c.Set("name", "private-inline.css"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:245
}

//line components.gox:247
func styleNoCacheHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("style"); if __e != nil { return }
		{
			__e = __c.Set("nocache", true); if __e != nil { return }
//line components.gox:248
			__e = __c.Set("name", "nocache-inline"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("h1 {\n\t\t\tcolor: red;\n\t\t}"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:253
}

//line components.gox:255
func cspHead(b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:256
		__e = __c.Any(styleExternalHead(b)); if __e != nil { return }
//line components.gox:257
		__e = __c.Any(moduleExternalHead(b)); if __e != nil { return }
	return })
//line components.gox:258
}

//line components.gox:260
func moduleExternalHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:262
			__e = __c.Set("src", doors.ResourceExternal(test.Host + "/module/index.js")); if __e != nil { return }
//line components.gox:263
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:264
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:265
}

//line components.gox:267
func moduleBundleHostHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:269
			__e = __c.Set("src", "/module/index.js"); if __e != nil { return }
//line components.gox:270
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:271
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:272
}

//line components.gox:274
func moduleBundleFSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line components.gox:276
		moduleBundleDir, _ := fs.Sub(moduleBundleFS, "module_bundle_src")

		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:279
			__e = __c.Set("id", "module-bundle-fs"); if __e != nil { return }
//line components.gox:280
			__e = __c.Set("src", doors.ResourceFS(moduleBundleDir, "index.ts")); if __e != nil { return }
//line components.gox:281
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("bundle", true); if __e != nil { return }
//line components.gox:283
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:284
}

//line components.gox:286
func moduleRawBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:288
			__e = __c.Set("id", "module-raw-bytes"); if __e != nil { return }
//line components.gox:289
			__e = __c.Set("src", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:290
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
//line components.gox:292
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:293
}

//line components.gox:295
func moduleRawBytesShortHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:297
			__e = __c.Set("id", "module-raw-bytes-short"); if __e != nil { return }
//line components.gox:298
			__e = __c.Set("src", moduleRawBytes); if __e != nil { return }
//line components.gox:299
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
//line components.gox:301
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:302
}

//line components.gox:304
func moduleRawBytesModifyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:306
			__e = __c.Set("id", "module-raw-bytes-modify"); if __e != nil { return }
//line components.gox:307
			__e = __c.Modify(doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:308
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
//line components.gox:310
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:311
}

//line components.gox:313
func modulePreloadBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:315
			__e = __c.Set("rel", "modulepreload"); if __e != nil { return }
//line components.gox:316
			__e = __c.Set("href", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:317
			__e = __c.Set("specifier", "module"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:318
}

//line components.gox:320
func moduleRawHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:322
			__e = __c.Set("id", "module-raw"); if __e != nil { return }
//line components.gox:323
			__e = __c.Set("src", doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
//line components.gox:324
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
//line components.gox:326
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:327
}

//line components.gox:329
func moduleBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:331
			__e = __c.Set("id", "module-bytes"); if __e != nil { return }
//line components.gox:332
			__e = __c.Set("src", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:333
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:334
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:335
}

//line components.gox:337
func moduleStringHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:339
			__e = __c.Set("id", "module-string"); if __e != nil { return }
//line components.gox:340
			__e = __c.Set("src", doors.ResourceString(string(moduleRawBytes))); if __e != nil { return }
//line components.gox:341
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:342
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:343
}

//line components.gox:345
func moduleProxyHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:347
			__e = __c.Set("src", doors.ResourceProxy(test.Host + "/module/index.js")); if __e != nil { return }
//line components.gox:348
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:349
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:350
}

//line components.gox:352
func moduleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:354
			__e = __c.Set("src", doors.ResourceLocalFS(modulePath + "/index.ts")); if __e != nil { return }
//line components.gox:355
			__e = __c.Set("name", "module.js"); if __e != nil { return }
//line components.gox:356
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:357
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:358
}

//line components.gox:360
func moduleVisibleHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:362
			__e = __c.Set("src", doors.ResourceLocalFS(modulePath + "/index.ts")); if __e != nil { return }
//line components.gox:363
			__e = __c.Set("id", "module-tag"); if __e != nil { return }
//line components.gox:364
			__e = __c.Set("type", "module"); if __e != nil { return }
//line components.gox:365
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:366
}

//line components.gox:368
func modulePreloadNamedHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:370
			__e = __c.Set("rel", "modulepreload"); if __e != nil { return }
//line components.gox:371
			__e = __c.Set("href", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
//line components.gox:372
			__e = __c.Set("name", "module-preload.js"); if __e != nil { return }
//line components.gox:373
			__e = __c.Set("specifier", "module"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:374
}

//line components.gox:376
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
//line components.gox:380
}

//line components.gox:382
func scriptRawHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:383
			__e = __c.Set("id", "script-raw-inline"); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("window.__importsValue = \"hello\""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:386
}

//line components.gox:388
func scriptInlineNamedExtHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:389
			__e = __c.Set("id", "script-inline-ext"); if __e != nil { return }
//line components.gox:389
			__e = __c.Set("name", "inline-script.js"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("window.__importsValue = \"hello\""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:392
}

//line components.gox:394
func scriptInlineBytesHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:396
			__e = __c.Set("id", "script-inline-bytes"); if __e != nil { return }
//line components.gox:397
			__e = __c.Set("src", doors.ResourceBytes([]byte(`window.__importsValue = "hello"`))); if __e != nil { return }
			__e = __c.Set("inline", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:399
}

//line components.gox:401
func scriptStringHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:402
			__e = __c.Set("src", doors.ResourceString(`window.__importsValue = "hello"`)); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:403
}

//line components.gox:405
func scriptPrivateHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:407
			__e = __c.Set("id", "script-private"); if __e != nil { return }
//line components.gox:408
			__e = __c.Set("src", doors.ResourceString(`window.__importsValue = "hello"`)); if __e != nil { return }
			__e = __c.Set("private", true); if __e != nil { return }
//line components.gox:410
			__e = __c.Set("name", "private-script.js"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:411
}

//line components.gox:413
func scriptNoCacheHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:415
			__e = __c.Set("id", "script-nocache"); if __e != nil { return }
//line components.gox:416
			__e = __c.Set("src", doors.ResourceString(`window.__importsValue = "hello"`)); if __e != nil { return }
			__e = __c.Set("nocache", true); if __e != nil { return }
//line components.gox:418
			__e = __c.Set("name", "nocache-script.js"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:419
}

//line components.gox:421
func reactHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:423
			__e = __c.Set("id", "preact-bundle"); if __e != nil { return }
//line components.gox:424
			__e = __c.Set("src", doors.ResourceLocalFS(preactPath + "/index.tsx")); if __e != nil { return }
//line components.gox:425
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("bundle", true); if __e != nil { return }
//line components.gox:427
			__e = __c.Set("specifier", "preact"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:429
			__e = __c.Set("id", "react-bundle"); if __e != nil { return }
//line components.gox:430
			__e = __c.Set("src", doors.ResourceLocalFS(reactPath + "/index.tsx")); if __e != nil { return }
//line components.gox:431
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Set("bundle", true); if __e != nil { return }
//line components.gox:433
			__e = __c.Set("specifier", "react"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:434
}

//line components.gox:436
func moduleTypeTSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:438
			__e = __c.Set("src", doors.ResourceLocalFS(modulePath + "/index.ts")); if __e != nil { return }
//line components.gox:439
			__e = __c.Set("type", "module/typescript"); if __e != nil { return }
//line components.gox:440
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:441
}

//line components.gox:443
func moduleTypeJSHead(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:445
			__e = __c.Set("src", doors.ResourceLocalFS(modulePath + "/index.js")); if __e != nil { return }
//line components.gox:446
			__e = __c.Set("type", "module/javascript"); if __e != nil { return }
//line components.gox:447
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:448
}

//line components.gox:450
func fileHandlerTypeBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("img"); if __e != nil { return }
		{
//line components.gox:452
			__e = __c.Set("src", doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("bad"))
		})); if __e != nil { return }
//line components.gox:455
			__e = __c.Set("type", "text/plain"); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:456
}

//line components.gox:458
func scriptDuplicateOutputBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:459
			__e = __c.Set("src", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
			__e = __c.Set("bundle", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:460
}

//line components.gox:462
func scriptHandlerBundleBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:464
			__e = __c.Set("src", doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write(moduleRawBytes)
		})); if __e != nil { return }
			__e = __c.Set("bundle", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:468
}

//line components.gox:470
func scriptRawTSBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:472
			__e = __c.Set("src", doors.ResourceBytes(moduleBytes)); if __e != nil { return }
			__e = __c.Set("raw", true); if __e != nil { return }
//line components.gox:474
			__e = __c.Set("type", "text/typescript"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:475
}

//line components.gox:477
func scriptSpecifierNonModuleBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:479
			__e = __c.Set("src", doors.ResourceBytes(moduleBytes)); if __e != nil { return }
//line components.gox:480
			__e = __c.Set("specifier", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:481
}

//line components.gox:483
func scriptInlineModuleBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:484
			__e = __c.Set("type", "module"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("window.__importsValue = \"bad\""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:487
}

//line components.gox:489
func scriptInlineTSBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:490
			__e = __c.Set("type", "text/typescript"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("const value: string = \"bad\"\n\t\twindow.__importsValue = value"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:494
}

//line components.gox:496
func scriptDirectBundleBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:497
			__e = __c.Set("src", "/module/index.js"); if __e != nil { return }
			__e = __c.Set("bundle", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:498
}

//line components.gox:500
func scriptHandlerInlineBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("script"); if __e != nil { return }
		{
//line components.gox:502
			__e = __c.Set("src", doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write(moduleRawBytes)
		})); if __e != nil { return }
			__e = __c.Set("inline", true); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line components.gox:506
}

//line components.gox:508
func modulePreloadInlineBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:509
			__e = __c.Set("rel", "modulepreload"); if __e != nil { return }
//line components.gox:509
			__e = __c.Set("href", doors.ResourceBytes(moduleRawBytes)); if __e != nil { return }
			__e = __c.Set("inline", true); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:510
}

//line components.gox:512
func styleHandlerPrivateBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:514
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:515
			__e = __c.Set("href", doors.ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write(styleRawBytes)
		})); if __e != nil { return }
			__e = __c.Set("private", true); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:519
}

//line components.gox:521
func styleDirectPrivateBad(_b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitVoid("link"); if __e != nil { return }
		{
//line components.gox:523
			__e = __c.Set("rel", "stylesheet"); if __e != nil { return }
//line components.gox:524
			__e = __c.Set("href", "/module/style.css"); if __e != nil { return }
			__e = __c.Set("private", true); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
	return })
//line components.gox:526
}

type ModuleFragment struct {
	test.NoBeam
}

//line components.gox:532
func (f *ModuleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:533
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
//line components.gox:538
}

type ReactFragment struct {
	test.NoBeam
}

//line components.gox:544
func (f *ReactFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:545
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
//line components.gox:550
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
//line components.gox:555
}

type ValueFragment struct {
	test.NoBeam
}

//line components.gox:561
func (f *ValueFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line components.gox:562
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
//line components.gox:566
}

type Empty struct {
	test.NoBeam
}

//line components.gox:572
func (f *Empty) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
	return })
//line components.gox:572
}
