package utils

import (
	"cmp"
	"context"
	"slices"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/gox"
)

type title struct {
	text  string
	attrs gox.Attrs
}

type metaMap = common.OrderedMap[string, *cake[gox.Attrs]]

func NewTitleMeta(inst core.Instance) core.TitleMeta {
	return &titleMeta{
		inst: inst,
	}
}

type titleMeta struct {
	mu        sync.Mutex
	rendered  bool
	inst      core.Instance
	title     cake[title]
	metaProps metaMap
	metaNames metaMap
}

func (t *titleMeta) Edit(cur gox.Cursor) error {
	t.mu.Lock()
	if t.rendered {
		t.mu.Unlock()
		panic("title meta rendered twice")
	}
	t.rendered = true
	titleID, title := t.title.current()
	metas := make([]metaEditor, 0, t.metaProps.Len()+t.metaNames.Len())
	t.dumpMeta(&metas, t.metaNames, false)
	t.dumpMeta(&metas, t.metaProps, true)
	t.mu.Unlock()
	if titleID != 0 {
		if err := cur.Init("title"); err != nil {
			return err
		}
		if title.attrs != nil {
			if err := cur.Modify(gox.ModifyFunc(func(ctx context.Context, tag string, attrs gox.Attrs) error {
				attrs.Inherit(title.attrs)
				return nil
			})); err != nil {
				return err
			}
		}
		if err := cur.Submit(); err != nil {
			return err
		}
		if err := cur.Text(title.text); err != nil {
			return err
		}
		if err := cur.Close(); err != nil {
			return err
		}
	}
	for _, m := range metas {
		if err := m.Edit(cur); err != nil {
			return err
		}
	}
	return nil
}

func (t *titleMeta) dumpMeta(s *[]metaEditor, m metaMap, property bool) {
	for name, c := range m.Iter() {
		_, attrs := c.current()
		*s = append(*s, metaEditor{
			name:     name,
			property: property,
			attrs:    attrs,
		})
	}
}

func (t *titleMeta) RemoveTitle(id uint) {
	t.mu.Lock()
	updated := t.title.remove(id)
	if !updated {
		t.mu.Unlock()
		return
	}
	id, value := t.title.current()
	t.mu.Unlock()
	t.callUpdateTitle(id, value.text, value.attrs)
}

func (t *titleMeta) UpdateTitle(value string, attrs gox.Attrs) (cancel context.CancelFunc) {
	t.mu.Lock()
	id := t.title.put(title{value, attrs})
	call := t.rendered
	t.mu.Unlock()
	cancel = func() {
		t.RemoveTitle(id)
	}
	if !call {
		return
	}
	t.callUpdateTitle(id, value, attrs)
	return
}

func (t *titleMeta) UpdateMeta(prop bool, name string, attrs gox.Attrs) (cancel context.CancelFunc) {
	m := &t.metaNames
	if prop {
		m = &t.metaProps
	}
	id, call := t.updateMeta(m, name, attrs)
	cancel = func() {
		t.RemoveMeta(prop, name, id)
	}
	if !call {
		return
	}
	t.callUpdateMeta(prop, id, name, attrs)
	return
}

func (t *titleMeta) RemoveMeta(prop bool, name string, id uint) {
	m := &t.metaNames
	if prop {
		m = &t.metaProps
	}
	id, attrs, call := t.removeMeta(m, name, id)
	if !call {
		return
	}
	if id == 0 {
		t.callRemoveMeta(prop, name)
		return
	}
	t.callUpdateMeta(prop, id, name, attrs)
}

func (t *titleMeta) updateMeta(m *metaMap, name string, attrs gox.Attrs) (id uint, call bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	c, ok := m.Get(name)
	if !ok {
		c = &cake[gox.Attrs]{}
		m.Set(name, c)
	}
	id = c.put(attrs)
	call = t.rendered
	return
}

func (t *titleMeta) removeMeta(m *metaMap, name string, id uint) (uint, gox.Attrs, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	c, ok := m.Get(name)
	if !ok {
		return 0, nil, false
	}
	call := c.remove(id)
	if c.isEmpty() {
		m.Delete(name)
	}
	if !call || !t.rendered {
		return 0, nil, false
	}
	id, e := c.current()
	return id, e, true
}

func (t *titleMeta) callRemoveMeta(prop bool, name string) {
	t.inst.UserCall(
		context.Background(),
		func() bool {
			t.mu.Lock()
			defer t.mu.Unlock()
			m := t.metaNames
			if prop {
				m = t.metaProps
			}
			_, ok := m.Get(name)
			return !ok
		},
		action.RemoveMeta{
			Name:     name,
			Property: prop,
		},
		nil,
		nil,
		action.CallParams{},
	)
}

func (t *titleMeta) callUpdateMeta(prop bool, id uint, name string, attrs gox.Attrs) {
	t.inst.UserCall(
		context.Background(),
		func() bool {
			t.mu.Lock()
			defer t.mu.Unlock()
			m := t.metaNames
			if prop {
				m = t.metaProps
			}
			c, ok := m.Get(name)
			if !ok {
				return false
			}
			return c.isCurrent(id)
		},
		action.UpdateMeta{
			Name:     name,
			Property: prop,
			Attrs:    common.AttrsToMap(attrs, t.inst.Logger()),
		},
		nil,
		nil,
		action.CallParams{},
	)
}

func (t *titleMeta) callUpdateTitle(id uint, value string, attrs gox.Attrs) {
	t.inst.UserCall(
		context.Background(),
		func() bool {
			t.mu.Lock()
			defer t.mu.Unlock()
			if id == 0 {
				return t.title.isEmpty()
			}
			return t.title.isCurrent(id)
		},
		action.UpdateTitle{
			Content: value,
			Attrs:   common.AttrsToMap(attrs, t.inst.Logger()),
		},
		nil,
		nil,
		action.CallParams{},
	)
}

type metaEditor struct {
	name     string
	property bool
	attrs    gox.Attrs
}

func (m metaEditor) Edit(cur gox.Cursor) error {
	if err := cur.InitVoid("meta"); err != nil {
		return err
	}
	if m.property {
		if err := cur.Set("property", m.name); err != nil {
			return err
		}
	} else {
		if err := cur.Set("name", m.name); err != nil {
			return err
		}
	}
	if err := cur.Modify(gox.ModifyFunc(func(ctx context.Context, tag string, attrs gox.Attrs) error {
		attrs.Inherit(m.attrs)
		return nil
	})); err != nil {
		return err
	}
	return cur.Submit()
}

type layer[T any] struct {
	id    uint
	value T
}

type cake[T any] struct {
	cursor uint
	layers []layer[T]
}

func (m *cake[T]) isEmpty() bool {
	return len(m.layers) == 0
}

func (m *cake[T]) isCurrent(id uint) bool {
	if len(m.layers) == 0 {
		return false
	}
	return m.layers[len(m.layers)-1].id == id
}

func (m *cake[T]) put(value T) uint {
	m.cursor += 1
	id := m.cursor
	m.layers = append(m.layers, layer[T]{id, value})
	return id
}

func (m *cake[T]) current() (uint, T) {
	if len(m.layers) == 0 {
		return 0, *new(T)
	}
	entry := m.layers[len(m.layers)-1]
	return entry.id, entry.value
}

func (m *cake[T]) remove(id uint) (changeCurrent bool) {
	index, ok := slices.BinarySearchFunc(m.layers, id, func(l layer[T], target uint) int {
		return cmp.Compare(l.id, target)
	})
	if !ok {
		return false
	}
	m.layers = slices.Delete(m.layers, index, index+1)
	return index == len(m.layers)
}
