// Managed by GoX v0.3.2

//line page.gox:1
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

//line page.gox:66
func (p *Page) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line page.gox:67
		if p.F != nil {
//line page.gox:69
			p.F.setBeam(p.Source)

//line page.gox:71
			__e = __c.Any(p.F); if __e != nil { return }
		}
	return })
//line page.gox:73
}

//line page.gox:75
func (p *Page) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line page.gox:76
		__e = __c.Any(Document(p)); if __e != nil { return }
	return })
//line page.gox:77
}
