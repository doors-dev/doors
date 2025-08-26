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
	"github.com/doors-dev/doors/internal/path"
)

func newNavigator[M any](adapter *path.Adapter[M], adapters map[string]path.AnyAdapter, model *M, detached bool, rerouted bool) *navigator[M] {
	return &navigator[M]{
		adapter:  adapter,
		adapters: adapters,
		detached: detached,
		rerouted: rerouted,
		model: beam.NewSourceBeamExt(*model, func(new, old M) bool {
			return !reflect.DeepEqual(new, old)
		}),
	}
}

const historyLimit = 32

type navigator[M any] struct {
	adapter      *path.Adapter[M]
	adapters     map[string]path.AnyAdapter
	detached     bool
	rerouted     bool
	model        beam.SourceBeam[M]
	solitaire    *solitaire
	mu           sync.Mutex
	historyIndex int
	history      [historyLimit]*navigatorState[M]
	pathCall     *setPathCall
	ctx          context.Context
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

func (n *navigator[M]) newLink(ctx context.Context, m any) (*Link, error) {
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
		on := func() {
			n.model.Update(ctx, *thisModel)
		}
		return &Link{
			location: location,
			on:       on,
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

func (n *navigator[M]) restore(r *http.Request) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	index := n.findPath(r.RequestURI)
	if index != -1 {
		entry := n.history[index]
		n.model.Update(n.ctx, *entry.model)
		return true
	}
	return false
}

func (n *navigator[M]) findPath(path string) int {
	index := n.historyIndex - 1
	for {
		if index == -1 {
			index = historyLimit - 1
		}
		entry := n.history[index]
		if entry == nil {
			break
		}
		if entry.path == path {
			return index
		}
		if index == n.historyIndex {
			break
		}
		index -= 1
	}
	return -1
}

func (n *navigator[M]) pushHistory(ns *navigatorState[M], sync bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if sync {
		if n.pathCall != nil {
			n.pathCall.cancel()
		}
		n.pathCall = &setPathCall{
			path:    ns.path,
			replace: false,
		}
		n.solitaire.Call(n.pathCall)
	}
	index := n.findPath(ns.path)
	if index != -1 {
		n.history[index] = ns
		return
	}
	n.history[n.historyIndex] = ns
	n.historyIndex += 1
	if n.historyIndex >= historyLimit {
		n.historyIndex = 0
	}
}

func (n *navigator[M]) init(ctx context.Context, solitaire *solitaire) {
	n.solitaire = solitaire
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
		n.pushHistory(&ns, !n.detached)
		return false
	})
	if !ok {
		return
	}
	n.pushHistory(&ns, n.rerouted && !n.detached)
}

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
