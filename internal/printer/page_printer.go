// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

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
		return p.cur.Send(j)
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
		return p.cur.Send(j)
	}
	if strings.EqualFold(openJob.Tag, "head") {
		p.headID = openJob.ID
		p.state = pageHead
		return p.cur.Send(j)
	}
	if strings.EqualFold(openJob.Tag, "script") {
		p.state = pageDone
		if err := p.insert(); err != nil {
			return err
		}
		return p.cur.Send(j)
	}
	if strings.EqualFold(openJob.Tag, "body") {
		p.state = pageDone
		if err := p.insertHead(); err != nil {
			return err
		}
		return p.cur.Send(j)
	}
	return p.cur.Send(j)
}

func (p *pagePrinter) head(j gox.Job) error {
	if openJob, ok := j.(*gox.JobHeadOpen); ok {
		if strings.EqualFold(openJob.Tag, "script") {
			p.state = pageDone
			if err := p.insert(); err != nil {
				return err
			}
		}
		return p.cur.Send(j)
	}
	if closeJob, ok := j.(*gox.JobHeadClose); ok {
		if closeJob.ID == p.headID {
			p.state = pageDone
			if err := p.insert(); err != nil {
				return err
			}
		}
		return p.cur.Send(j)
	}
	return p.cur.Send(j)
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
		if err := p.cur.AttrSet("type", "importmap"); err != nil {
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
