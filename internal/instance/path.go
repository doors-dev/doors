// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package instance

import (
	"context"
	"errors"
	"fmt"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/path"
)

type Link struct {
	location *path.Location
	on       func()
}

func (h *Link) Path() (string, bool) {
	if h.location == nil {
		return "", false
	}
	return h.location.String(), true
}

func (h *Link) ClickHandler() (func(), bool) {
	if h.on == nil {
		return nil, false
	}
	return h.on, true
}

func (inst *Instance[M]) newLink(ctx context.Context, m any) (*Link, error) {
	thisModel, ok := m.(*M)
	if !ok {
		direct, ok := m.(M)
		if ok {
			thisModel = &direct
		}
	}
	if thisModel != nil {
		location, err := inst.adapter.Encode(thisModel)
		if err != nil {
			return nil, err
		}
		on := func() {
			inst.beam.Update(ctx, *thisModel)
		}
		return &Link{
			location: location,
			on:       on,
		}, nil
	}
	name := path.GetAdapterName(m)
	adapter, found := inst.session.router.Adapters()[name]
	if !found {
		return nil, errors.New(fmt.Sprint("Adapter for ", name, " is not registered"))
	}
	location, err := adapter.EncodeAny(m)
	if err != nil {
		return nil, err
	}
	return &Link{
		location: location,
		on:       nil,
	}, nil
}

func (inst *Instance[M]) setupPathSync(ctx context.Context) {
	ctx = context.WithValue(ctx, common.ThreadCtxKey, nil)
	inst.ctx = context.WithValue(ctx, common.RenderMapCtxKey, nil)
	if inst.opt.Detached {
		return
	}
	type path struct {
		path string
		err  error
	}
	var call *setPathCaller
	sync := func(path string, replace bool) {
		if call != nil {
			call.cancel()
		}
		call = &setPathCaller{
			path:    path,
			replace: replace,
		}
		inst.core.Call(call)
	}
	beam := beam.NewBeam(inst.beam, func(model M) path {
		loc, err := inst.adapter.Encode(&model)
		if err != nil {
			return path{
				path: "",
				err:  err,
			}
		}
		return path{
			path: loc.String(),
			err:  nil,
		}
	})
	v, _ := beam.ReadAndSub(ctx, func(ctx context.Context, p path) bool {

		if p.err != nil {
			println(p.err.Error)
			return false
		}
		sync(p.path, false)
		return false
	})
	if !inst.opt.Rerouted {
		return
	}
	sync(v.path, true)
}

func (inst *Instance[M]) UpdatePath(m any, adapter path.AnyAdapter) bool {
	model, ok := m.(*M)
	if ok {
		inst.beam.Update(inst.ctx, *model)
		return true
	}
	location, err := adapter.EncodeAny(m)
	if err != nil {
		println(err.Error())
		return false
	}
	inst.core.Call(&LocationAssign{
		Origin: true,
		Href:   location.String(),
	})
	return true
}

