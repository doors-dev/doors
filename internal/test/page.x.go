// Managed by GoX v0.1.32

//line page.gox:1
package test

import (
	"github.com/doors-dev/doors"
	"github.com/doors-dev/gox"
)

type NoBeam struct{}

type PathLens = doors.Source[Path]

func (f NoBeam) setBeam(_ PathLens) {}

type Beam struct {
	B PathLens
}

func (f *Beam) setBeam(b PathLens) {
	f.B = b
}

type Path struct {
	Vh bool `path:""`
	Vs bool `path:"/s"`
	Vp bool `path:"/s/:P"`
	P string
}

type Fragment interface {
	setBeam(b PathLens)
	gox.Comp
}

type Page struct {
	Source PathLens
	F Fragment
	H func(PathLens) gox.Elem
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

//line page.gox:52
func (p *Page) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line page.gox:53
		if p.F != nil {
//line page.gox:55
			p.F.setBeam(p.Source)

//line page.gox:57
			__e = __c.Any(p.F); if __e != nil { return }
		}
	return })
//line page.gox:59
}

//line page.gox:61
func (p *Page) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line page.gox:62
		__e = __c.Any(Document(p)); if __e != nil { return }
	return })
//line page.gox:63
}
