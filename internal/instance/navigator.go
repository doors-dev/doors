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
	"log/slog"
	"net/http"
	"reflect"
	"sync"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/path"
)

func newNavigator[M any](adapter *path.Adapter[M], adapters map[string]path.AnyAdapter, model *M, detached bool, rerouted bool, optimistic bool) *navigator[M] {
	return &navigator[M]{
		adapter:     adapter,
		adapters:    adapters,
		detached:    detached,
		rerouted:    rerouted,
		optimistic:  optimistic,
		historyHead: &historyHead[M]{},
		cancelPrev:  func() {},
		model: beam.NewSourceBeamExt(*model, func(new, old M) bool {
			return !reflect.DeepEqual(new, old)
		}),
	}
}

const historyLimit = 32

type navigator[M any] struct {
	inst        Core
	adapter     *path.Adapter[M]
	adapters    map[string]path.AnyAdapter
	optimistic  bool
	detached    bool
	rerouted    bool
	model       beam.SourceBeam[M]
	mu          sync.Mutex
	historyHead *historyHead[M]
	cancelPrev  context.CancelFunc
	ctx         context.Context
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
		n.cancelPrev()
		n.cancelPrev = n.inst.SimpleCall(ctx, &action.SetPath{Path: ns.path, Replace: replace}, nil, nil, action.CallParams{Optimistic: n.optimistic})
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
