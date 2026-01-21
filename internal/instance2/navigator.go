// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package instance2

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"sync"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/path"
)

func newNavigator[M any](adapter *path.Adapter[M], adapters map[string]path.AnyAdapter, model *M, detached bool, rerouted bool) *navigator[M] {
	return &navigator[M]{
		adapter:     adapter,
		adapters:    adapters,
		detached:    detached,
		rerouted:    rerouted,
		historyHead: &historyHead[M]{},
		model: beam.NewSourceBeamEqual(*model, func(new, old M) bool {
			return reflect.DeepEqual(new, old)
		}),
	}
}

const historyLimit = 32

type navigator[M any] struct {
	inst        Core
	adapter     *path.Adapter[M]
	adapters    map[string]path.AnyAdapter
	detached    bool
	rerouted    bool
	model       beam.SourceBeam[M]
	mu          sync.Mutex
	historyHead *historyHead[M]
	ctx         context.Context
	seq         int
}

func (n *navigator[M]) isDetached() bool {
	return n.detached
}

func (n *navigator[M]) getBeam() beam.SourceBeam[M] {
	return n.model
}

type navigatorState[M any] struct {
	model *M
	path  string
	err   error
}

func (n *navigator[M]) newLink(m any) (*Link, error) {
	thisModel, ok := m.(*M)
	if !ok {
		direct, ok := m.(M)
		if ok {
			thisModel = &direct
		}
	}
	if thisModel != nil {
		location, err := n.adapter.Encode(thisModel)
		if err != nil {
			return nil, err
		}
		return &Link{
			location: location,
			on: func(ctx context.Context) {
				n.model.Update(ctx, *thisModel)
			},
		}, nil
	}
	name := path.GetAdapterName(m)
	adapter, found := n.adapters[name]
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

func (n *navigator[M]) init(ctx context.Context, inst Core) {
	n.inst = inst
	n.ctx = ctx
	state := beam.NewBeam(n.model, func(m M) navigatorState[M] {
		l, err := n.adapter.Encode(&m)
		if err != nil {
			slog.Error(
				"Path model encoding error on beam update",
				slog.String("error", err.Error()),
				slog.String("model", fmt.Sprintf("%+v", m)),
			)
			return navigatorState[M]{
				model: &m,
				err:   err,
			}
		}
		return navigatorState[M]{
			model: &m,
			path:  l.String(),
		}
	})
	ns, ok := state.ReadAndSub(ctx, func(ctx context.Context, ns navigatorState[M]) bool {
		n.pushHistory(ctx, &ns, !n.detached, false)
		return false
	})
	if !ok {
		return
	}
	n.pushHistory(ctx, &ns, n.rerouted && !n.detached, true)
}

func (n *navigator[M]) restore(r *http.Request) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	entry := n.historyHead.retrieve(r.RequestURI)
	if entry != nil {
		n.model.Update(n.ctx, *entry.model)
		return true
	}
	return false
}

func (n *navigator[M]) pushHistory(ctx context.Context, ns *navigatorState[M], sync bool, replace bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if sync {
		n.seq += 1
		seq := n.seq
		n.inst.CallCheck(
			func() bool {
				n.mu.Lock()
				defer n.mu.Unlock()
				return seq == n.seq
			},
			&action.SetPath{Path: ns.path, Replace: replace},
			nil,
			nil,
			action.CallParams{},
		)
	}
	n.historyHead.push(ns)
}

type historyHead[M any] struct {
	entry *historyEntry[M]
}

func (h *historyHead[M]) retrieve(path string) *navigatorState[M] {
	if h.entry == nil {
		return nil
	}
	return h.entry.retrieve(path)
}

func (h *historyHead[M]) push(n *navigatorState[M]) {
	entry := &historyEntry[M]{
		n: n,
	}
	if h.entry == nil {
		h.entry = entry
		return
	}
	entry.next = h.entry
	entry.next.shake(entry, n.path, 1)

	h.entry = entry
}

type historyEntry[M any] struct {
	next *historyEntry[M]
	n    *navigatorState[M]
}

func (e *historyEntry[M]) retrieve(path string) *navigatorState[M] {
	if e.n.path == path {
		return e.n
	}
	if e.next == nil {
		return nil
	}
	return e.next.retrieve(path)
}

func (e *historyEntry[M]) shake(prev *historyEntry[M], path string, count int) {
	if count == historyLimit {
		prev.next = nil
		return
	}
	if e.n.path == path {
		prev.next = e.next
		return
	}
	if e.next == nil {
		return
	}
	e.next.shake(e, path, count+1)
}

type Link struct {
	location *path.Location
	on       func(context.Context)
}

func (h *Link) Path() (string, bool) {
	if h.location == nil {
		return "", false
	}
	return h.location.String(), true
}

func (h *Link) ClickHandler() (func(context.Context), bool) {
	if h.on == nil {
		return nil, false
	}
	return h.on, true
}
