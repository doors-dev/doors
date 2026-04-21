// Managed by GoX v0.1.28

//line page.gox:1
package test

import (
	"github.com/doors-dev/doors"
	"github.com/doors-dev/gox"
)

type NoBeam struct{}

func (f NoBeam) setBeam(_ doors.Source[Path]) {}

type Beam struct {
	B doors.Source[Path]
}

func (f *Beam) setBeam(b doors.Source[Path]) {
	f.B = b
}

type Path struct {
	Vh bool `path:""`
	Vs bool `path:"/s"`
	Vp bool `path:"/s/:P"`
	P string
}

type Fragment interface {
	setBeam(b doors.Source[Path])
	gox.Comp
}

type Page struct {
	Source doors.Source[Path]
	F Fragment
	H func(doors.Source[Path]) gox.Elem
	Header string
}

func (p *Page) h1() string {
	return p.Header
}

func (p *Page) head() gox.Elem {
	if p.H == nil {
		return nil
	}
	return p.H(p.Source)
}

//line page.gox:50
func (p *Page) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line page.gox:51
		if p.F != nil {
//line page.gox:53
			p.F.setBeam(p.Source)

//line page.gox:55
			__e = __c.Any(p.F); if __e != nil { return }
		}
	return })
//line page.gox:57
}

//line page.gox:59
func (p *Page) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line page.gox:60
		__e = __c.Any(Document(p)); if __e != nil { return }
	return })
//line page.gox:61
}
