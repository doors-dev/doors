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

package instance

import (
	"context"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/gox"
)

func newTitleMeta(inst core.Instance) *titleMeta {
	return &titleMeta{
		inst: inst,
	}
}

type meta struct {
	name     string
	property bool
	attrs    gox.Attrs
}

func (m meta) Edit(cur gox.Cursor) error {
	if err := cur.InitVoid("meta"); err != nil {
		return err
	}
	if m.property {
		if err := cur.AttrSet("property", m.name); err != nil {
			return err
		}
	} else {
		if err := cur.AttrSet("name", m.name); err != nil {
			return err
		}
	}
	if err := cur.AttrMod(gox.ModifyFunc(func(ctx context.Context, tag string, attrs gox.Attrs) error {
		attrs.Inherit(m.attrs)
		return nil
	})); err != nil {
		return err
	}
	return cur.Submit()
}

type titleMeta struct {
	mu         sync.Mutex
	rendered   bool
	inst       core.Instance
	title      string
	titleAttrs gox.Attrs
	meta       []meta
}

func (t *titleMeta) Edit(cur gox.Cursor) error {
	t.mu.Lock()
	if t.rendered {
		t.mu.Unlock()
		panic("title meta rendered twice")
	}
	t.rendered = true
	t.mu.Unlock()
	if err := cur.Init("title"); err != nil {
		return err
	}
	if t.titleAttrs != nil {
		if err := cur.AttrMod(gox.ModifyFunc(func(ctx context.Context, tag string, attrs gox.Attrs) error {
			attrs.Inherit(t.titleAttrs)
			return nil
		})); err != nil {
			return err
		}
	}
	if err := cur.Submit(); err != nil {
		return err
	}
	if err := cur.Text(t.title); err != nil {
		return err
	}
	if err := cur.Close(); err != nil {
		return err
	}
	for _, m := range t.meta {
		if err := m.Edit(cur); err != nil {
			return err
		}
	}
	t.meta = nil
	t.titleAttrs = nil
	t.title = ""
	return nil
}

func (t *titleMeta) updateTitle(content string, attrs gox.Attrs) {
	t.mu.Lock()
	if t.rendered {
		t.mu.Unlock()
		t.inst.CallCtx(
			context.Background(),
			action.UpdateTitle{
				Content: content,
				Attrs:   common.AttrsToMap(attrs),
			},
			nil,
			nil,
			action.CallParams{},
		)
		return
	}
	defer t.mu.Unlock()
	t.title = content
	t.titleAttrs = attrs
}

func (t *titleMeta) updateMeta(m meta) {
	t.mu.Lock()
	if t.rendered {
		t.mu.Unlock()
		t.inst.CallCtx(
			context.Background(),
			action.UpdateMeta{
				Name:     m.name,
				Property: m.property,
				Attrs:    common.AttrsToMap(m.attrs),
			},
			nil,
			nil,
			action.CallParams{},
		)
		return
	}
	defer t.mu.Unlock()
	t.meta = append(t.meta, m)

}

func (inst *Instance[M]) UpdateTitle(content string, attrs gox.Attrs) {
	inst.meta.updateTitle(content, attrs)
}

func (inst *Instance[M]) UpdateMeta(name string, property bool, attrs gox.Attrs) {
	inst.meta.updateMeta(meta{name, property, attrs})

}
