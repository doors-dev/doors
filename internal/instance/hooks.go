// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package instance

import (
	"net/http"

	"github.com/doors-dev/doors/internal/door"
)

func (i *core[M]) TriggerHook(doorId uint64, hookId uint64, w http.ResponseWriter, r *http.Request, track uint64) bool {
	hook := i.getHook(doorId, hookId)
	if hook == nil {
		return false
	}
	done, ok := hook.Trigger(w, r)
	if !ok {
		return false
	}
	if track != 0 {
		i.solitaire.Call(reportHook(track))
	}
	if done {
		i.removeHook(doorId, hookId)
	}
	return true
}

func (i *core[M]) CancelHook(doorId uint64, hookId uint64, err error) {
	hook := i.removeHook(doorId, hookId)
	if hook == nil {
		return
	}
	hook.Cancel(err)
}

func (i *core[M]) CancelHooks(doorId uint64, err error) {
	i.hooksMu.Lock()
	hooks, ok := i.hooks[doorId]
	if !ok {
		i.hooksMu.Unlock()
		return
	}
	delete(i.hooks, doorId)
	i.hooksMu.Unlock()
	for id := range hooks {
		hooks[id].Cancel(err)
	}
}

func (i *core[M]) RegisterHook(doorId uint64, hookId uint64, hook *door.DoorHook) {
	i.hooksMu.Lock()
	defer i.hooksMu.Unlock()
	hooks, ok := i.hooks[doorId]
	if !ok {
		hooks = make(map[uint64]*door.DoorHook)
		i.hooks[doorId] = hooks
	}
	hooks[hookId] = hook
}

func (i *core[M]) getHook(doorId uint64, hookId uint64) *door.DoorHook {
	i.hooksMu.Lock()
	defer i.hooksMu.Unlock()
	hooks, ok := i.hooks[doorId]
	if !ok {
		return nil
	}
	hook, ok := hooks[hookId]
	if !ok {
		return nil
	}
	return hook
}

func (i *core[M]) removeHook(doorId uint64, hookId uint64) *door.DoorHook {
	i.hooksMu.Lock()
	defer i.hooksMu.Unlock()
	hooks, ok := i.hooks[doorId]
	if !ok {
		return nil
	}
	hook, ok := hooks[hookId]
	if !ok {
		return nil
	}
	delete(hooks, hookId)
	if len(hooks) == 0 {
		delete(i.hooks, doorId)
	}
	return hook
}
