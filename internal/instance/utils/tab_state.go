package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"slices"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front/actions"
)

type state int

const (
	initial state = iota
	synced
	updated
)

type TabState struct {
	state   state
	storage *map[string]json.RawMessage
}

func (s TabState) clone() TabState {
	m := maps.Clone(s.get())
	return TabState{
		state:   s.state,
		storage: &m,
	}
}

func (s TabState) get() map[string]json.RawMessage {
	return (*s.storage)
}

func (s TabState) equal(other TabState) bool {
	if s.state != other.state {
		return false
	}
	if s.storage == other.storage {
		return true
	}
	return maps.EqualFunc(s.get(), other.get(), func(a json.RawMessage, b json.RawMessage) bool {
		return bytes.Equal(a, b)
	})
}

type TabStateManager = *tabStateManager

func NewTabStateManager(inst core.Instance, ctx context.Context) TabStateManager {
	s := make(map[string]json.RawMessage)
	return &tabStateManager{
		inst: inst,
		ctx:  ctx,
		source: beam.NewSource(TabState{
			storage: &s,
		}, func(a TabState, b TabState) bool {
			return a.equal(b)
		}, false),
	}
}

type tabStateManager struct {
	inst   core.Instance
	ctx    context.Context
	source beam.Source[TabState]
	seq    atomic.Int32
}

func (t TabStateManager) Initialize(data map[string]json.RawMessage) {
	noop := false
	t.source.Mutate(t.ctx, func(s TabState) TabState {
		if s.state != initial {
			noop = true
			return s
		}
		clone := s.clone()
		clone.state = synced
		if len(data) == 0 {
			if len(clone.get()) != 0 {
				clone.state = updated
			}
			return clone
		}
		for key, existingValue := range clone.get() {
			initValue, ok := data[key]
			if ok {
				delete(data, key)
			}
			if clone.state == updated {
				continue
			}
			if !ok {
				clone.state = updated
			} else if !slices.Equal(existingValue, initValue) {
				clone.state = updated
			}
		}
		maps.Copy(clone.get(), data)
		return clone
	})
	if noop {
		return
	}
	t.source.Sub(t.ctx, func(ctx context.Context, s TabState) bool {
		if s.state != updated {
			return false
		}
		seq := t.seq.Add(1)
		t.inst.UserCallCheck(
			func() bool {
				return seq == t.seq.Load()
			},
			actions.UpdateState{State: s.get()},
			nil,
			nil,
			actions.CallParams{},
		)
		return false
	})
}

func (t TabStateManager) Derive[T any](key string, equal func(new T, old T) bool) beam.Lens[TabState, *T] {
	logger := common.Logger(t.ctx)
	return beam.NewLens(t.source, func(s TabState) *T {
		raw, ok := s.get()[key]
		if !ok {
			if s.state == initial {
				return nil
			}
			return new(T)
		}
		var value T
		if err := json.Unmarshal(raw, &value); err != nil {
			logger.Error("Tab state decoding error", "key", key, "error", err)
			if s.state == initial {
				return nil
			}
			return new(T)
		}
		return &value
	}, func(s TabState, t *T) TabState {
		if t == nil {
			if _, ok := s.get()[key]; !ok {
				return s
			}
			clone := s.clone()
			if clone.state == synced {
				clone.state = updated
			}
			delete(clone.get(), key)
			return clone
		}
		raw, err := json.Marshal(*t)
		if err != nil {
			logger.Error("Tab state encoding error", "key", key, "error", err)
			return s
		}
		clone := s.clone()
		if clone.state == synced {
			clone.state = updated
		}
		clone.get()[key] = raw
		return clone
	}, func(new, old *T) bool {
		if new == old {
			return true
		}
		if new == nil || old == nil {
			return false
		}
		return equal(*new, *old)
	})
}
