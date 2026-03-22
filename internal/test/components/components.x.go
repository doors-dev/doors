// Managed by GoX v0.1.15+dirty

package components

import (
	"context"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

func head(b doors.Source[test.Path]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(doors.TitleMeta(b, func(p test.Path) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			if p.Vh {
				__e = __c.Init("title"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("home"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
					__e = __c.AttrSet("name", "description"); if __e != nil { return }
					__e = __c.AttrSet("content", "Welcome to the home page"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
					__e = __c.AttrSet("name", "keywords"); if __e != nil { return }
					__e = __c.AttrSet("content", "home, main, index"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
					__e = __c.AttrSet("name", "og:title"); if __e != nil { return }
					__e = __c.AttrSet("content", "Home Page"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
			} else if p.Vs {
				__e = __c.Init("title"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("s"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
					__e = __c.AttrSet("name", "description"); if __e != nil { return }
					__e = __c.AttrSet("content", "String page description"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
					__e = __c.AttrSet("name", "category"); if __e != nil { return }
					__e = __c.AttrSet("content", "text-content"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
			} else  {
				__e = __c.Init("title"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Any(p.P); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
					__e = __c.AttrSet("name", "description"); if __e != nil { return }
					__e = __c.AttrSet("content", "Page for parameter: " + p.P); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
					__e = __c.AttrSet("name", "keywords"); if __e != nil { return }
					__e = __c.AttrSet("content", "param, " + p.P); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
					__e = __c.AttrSet("name", "og:title"); if __e != nil { return }
					__e = __c.AttrSet("content", "Param: " + p.P); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.InitVoid("meta"); if __e != nil { return }
				{
					__e = __c.AttrSet("name", "author"); if __e != nil { return }
					__e = __c.AttrSet("content", "Parameter Author"); if __e != nil { return }
				}
				__e = __c.Submit(); if __e != nil { return }
			}
		return })
	})); if __e != nil { return }
	return })
}

type LinksFragment struct {
	test.NoBeam
	Param string
}

func (f *LinksFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = doors.AHref{
		Model: test.Path{
			Vh: true,
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("a"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "home"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("home"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = doors.AHref{
		Model: test.Path{
			Vp: true,
			P: f.Param,
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("a"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "param"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("param"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = doors.AHref{
		Model: test.Path{
			Vs: true,
			P: f.Param,
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("a"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "string"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("string"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.InitVoid("br"); if __e != nil { return }
		{
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.InitVoid("br"); if __e != nil { return }
		{
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			r.After(doors.ActionOnlyEmit("alert", "Hello!"))
			return false
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("alert"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw("$on(\"alert\", (message) => {\n\t\t\talert(message)\n\t\t})"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}
