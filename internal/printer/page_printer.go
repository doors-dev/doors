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

package printer

import (
	"context"
	"io"
	"strings"

	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

func NewPagePrinter(w io.Writer, ctx context.Context, static bool, importMap []byte, meta gox.Editor) gox.Printer {
	cur := gox.NewCursor(ctx, defaultPrinter{w})
	return &pagePrinter{cur: cur, static: static, importMap: importMap, meta: meta}
}

type pagePrinterState int

const (
	pageScan pagePrinterState = iota
	pageHead
	pageDone
)

type pagePrinter struct {
	cur       gox.Cursor
	static    bool
	importMap []byte
	state     pagePrinterState
	meta      gox.Editor
	headID    uint64
}

func (p *pagePrinter) Send(j gox.Job) error {
	switch p.state {
	case pageDone:
		return p.cur.Printer().Send(j)
	case pageScan:
		return p.scan(j)
	case pageHead:
		return p.head(j)
	default:
		panic("unknown state")
	}
}

func (p *pagePrinter) scan(j gox.Job) error {
	openJob, ok := j.(*gox.JobHeadOpen)
	if !ok {
		return p.cur.Printer().Send(j)
	}
	if strings.EqualFold(openJob.Tag, "head") {
		p.headID = openJob.ID
		p.state = pageHead
		return p.cur.Printer().Send(j)
	}
	if strings.EqualFold(openJob.Tag, "script") {
		p.state = pageDone
		if err := p.insert(); err != nil {
			return err
		}
		return p.cur.Printer().Send(j)
	}
	if strings.EqualFold(openJob.Tag, "body") {
		p.state = pageDone
		if err := p.insertHead(); err != nil {
			return err
		}
		return p.cur.Printer().Send(j)
	}
	return p.cur.Printer().Send(j)
}

func (p *pagePrinter) head(j gox.Job) error {
	if openJob, ok := j.(*gox.JobHeadOpen); ok {
		if strings.EqualFold(openJob.Tag, "script") {
			p.state = pageDone
			if err := p.insert(); err != nil {
				return err
			}
		}
		return p.cur.Printer().Send(j)
	}
	if closeJob, ok := j.(*gox.JobHeadClose); ok {
		if closeJob.ID == p.headID {
			p.state = pageDone
			if err := p.insert(); err != nil {
				return err
			}
		}
		return p.cur.Printer().Send(j)
	}
	return p.cur.Printer().Send(j)
}

func (p *pagePrinter) insertHead() error {
	if err := p.cur.Init("head"); err != nil {
		return err
	}
	if err := p.cur.Submit(); err != nil {
		return err
	}
	if err := p.insert(); err != nil {
		return err
	}
	if err := p.cur.Close(); err != nil {
		return err
	}
	return nil
}

func (p *pagePrinter) insert() error {
	if !p.static {
		if err := p.cur.Comp(front.Include); err != nil {
			return err
		}
	}
	if err := p.meta.Edit(p.cur); err != nil {
		return err
	}
	if len(p.importMap) > 0 {
		if err := p.cur.Init("script"); err != nil {
			return err
		}
		if err := p.cur.Set("type", "importmap"); err != nil {
			return err
		}
		if err := p.cur.Submit(); err != nil {
			return err
		}
		if err := p.cur.Bytes(p.importMap); err != nil {
			return err
		}
		if err := p.cur.Close(); err != nil {
			return err
		}
	}
	return nil
}
