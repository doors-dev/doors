package instance2

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/doors-dev/doors/internal/door2"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

func (inst *Instance[M]) render(w http.ResponseWriter, r *http.Request, pipe door2.Pipe) error {
	gz := !inst.Conf().ServerDisableGzip && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
	importMap, importHash := inst.importMap.generate()
	inst.importMap = nil
	inst.renderHeaders(w, gz, importHash)
	if gz {
		wgz := gzip.NewWriter(w)
		pipe.SendTo(&pagePrinter[M]{
			inst:      inst,
			w:         wgz,
			importMap: importMap,
		})
		return wgz.Close()
	}
	pipe.SendTo(&pagePrinter[M]{
		inst:      inst,
		w:         w,
		importMap: importMap,
	})
	return nil
}

func (inst *Instance[M]) renderHeaders(w http.ResponseWriter, gz bool, importHash []byte) {
	if inst.csp != nil {
		if importHash != nil {
			inst.csp.ScriptHash(importHash)
		}
		header := inst.csp.Generate()
		w.Header().Add("Content-Security-Policy", header)
		inst.csp = nil
	}
	if gz {
		w.Header().Set("Content-Encoding", "gzip")
	}
}

func (inst *Instance[M]) renderResources(ctx context.Context, w io.Writer, importMap []byte, wrap bool) error {
	cur := gox.NewCursor(ctx, gox.NewPrinter(w))
	if wrap {
		if err := cur.Init("head"); err != nil {
			return err
		}
		if err := cur.Submit(); err != nil {
			return err
		}
	}

	if err := front.Include(cur); err != nil {
		return err
	}
	if importMap != nil {
		if err := cur.Init("script"); err != nil {
			return err
		}
		if err := cur.AttrSet("type", "importmap"); err != nil {
			return err
		}
		if err := cur.Submit(); err != nil {
			return err
		}
		if err := cur.Bytes(importMap); err != nil {
			return err
		}
		if err := cur.Close(); err != nil {
			return err
		}
	}
	if wrap {
		if err := cur.Close(); err != nil {
			return err
		}
	}
	return nil
}

type printerResourceState int

const (
	printerLooking printerResourceState = iota
	printerInside
	printerInserted
)

type pagePrinter[M any] struct {
	inst        *Instance[M]
	insertState printerResourceState
	importMap   []byte
	headID      uint64
	w           io.Writer
}


func (p *pagePrinter[M]) Send(job gox.Job) error {
	switch p.insertState {
	case printerLooking:
		job, ok := job.(*gox.JobHeadOpen)
		switch true {
		case !ok:
		case strings.EqualFold(job.Tag, "head"):
			p.headID = job.ID
			p.insertState = printerInside
		case strings.EqualFold(job.Tag, "body"):
			if err := p.inst.renderResources(job.Context(), p.w, p.importMap, true); err != nil {
				return err
			}
			p.insertState = printerInserted
		}
	case printerInside:
		openJob, openOk := job.(*gox.JobHeadOpen)
		closeJob, closeOk := job.(*gox.JobHeadClose)
		if (openOk && strings.EqualFold(openJob.Tag, "script")) || (closeOk && closeJob.ID == p.headID) {
			if err := p.inst.renderResources(job.Context(), p.w, p.importMap, true); err != nil {
				return err
			}
			p.insertState = printerInserted
		}
	}
	return job.Output(p.w)
}

/*
type renderState int

const (
	lookingForHead printerResourceState = iota
	insideHead
	headDone
)

func (inst *Instance[M]) renderBody(w io.Writer, js door2.JobStream, importMap []byte) error {
	var state printerResourceState
	var headId uint64
	for {
		job, ok := js.Next()
		if !ok {
			break
		}
		switch state {
		case lookingForHead:
			job, ok := job.(*gox.JobHeadOpen)
			switch true {
			case !ok:
			case strings.EqualFold(job.Tag, "head"):
				headId = job.Id
				state = insideHead
			case strings.EqualFold(job.Tag, "body"):
				if err := inst.renderResources(job.Context(), w, importMap, true); err != nil {
					return err
				}
				state = headDone
			}
		case insideHead:
			openJob, openOk := job.(*gox.JobHeadOpen)
			closeJob, closeOk := job.(*gox.JobHeadClose)
			if (openOk && strings.EqualFold(openJob.Tag, "script")) || (closeOk && closeJob.Id == headId) {
				if err := inst.renderResources(job.Context(), w, importMap, true); err != nil {
					return err
				}
				state = headDone
			}
		}
		if err := job.Output(w); err != nil {
			return err
		}
	}
	return nil
} */

