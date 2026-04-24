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

package beam

import (
	"context"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
)

type Door interface {
	Runtime() shredder.Runtime
	ReadFrame() shredder.Frame
	Context() context.Context
}

type parentScreen interface {
	removeSub(*screen)
}

type Cinema = *cinema

func NewCinema(parent Cinema, door Door) Cinema {
	return &cinema{
		parent: parent,
		door:   door,
	}
}

type cinema struct {
	mu          sync.Mutex
	parent      *cinema
	door        Door
	screens     map[common.ID]*screen
	removeGuard shredder.ReadStarveWriteThread
}

func (c Cinema) runtime() shredder.Runtime {
	return c.door.Runtime()
}

func (c Cinema) ReadFrame() shredder.Frame {
	return c.removeGuard.Read()
}

func (c Cinema) writeFrame() shredder.Frame {
	return c.removeGuard.Write()
}

func (c Cinema) IsEmpty() bool {
	return len(c.screens) == 0
}

func (c *cinema) isKilled() bool {
	return c.door.Context().Err() != nil
}

func (c *cinema) Cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, screen := range c.screens {
		screen.cancel()
	}
	clear(c.screens)

}

func (c *cinema) ctx() context.Context {
	return c.door.Context()
}

func (c *cinema) tryRemove(sourceId common.ID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	scr, ok := c.screens[sourceId]
	if !ok {
		return
	}
	if scr.tryRemove() {
		delete(c.screens, sourceId)
	}
}

func (c *cinema) addWatcher(src anySource, w *watcher) bool {
	c.mu.Lock()
	s, ok := c.getScreen(src)
	if !ok {
		c.mu.Unlock()
		return false
	}
	c.mu.Unlock()
	seq, frame := s.addWatcher(w)
	defer frame.Release()
	ctx := ctex.SyncFrameInsert(c.ctx(), frame)
	w.init(ctx, seq)
	return true
}

func (c *cinema) getScreen(src anySource) (*screen, bool) {
	if c.isKilled() {
		return nil, false
	}
	if c.screens == nil {
		c.screens = make(map[common.ID]*screen)
	}
	scr, ok := c.screens[src.getID()]
	if ok {
		return scr, true
	}
	scr = &screen{
		sourceID: src.getID(),
		cinema:   c,
	}
	c.screens[src.getID()] = scr
	if c.parent == nil {
		src.addSub(scr)
	} else {
		if !c.parent.addSub(src, scr) {
			return nil, false
		}
	}
	return scr, true
}

func (c *cinema) addSub(src anySource, scr *screen) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	s, ok := c.getScreen(src)
	if !ok {
		return false
	}
	s.addSub(scr)
	return true
}
