package instance

import (
	"net/http"

	"github.com/doors-dev/doors/internal/node"
)

func (i *core[M]) TriggerHook(nodeId uint64, hookId uint64, w http.ResponseWriter, r *http.Request) bool {
	hook := i.getHook(nodeId, hookId)
	if hook == nil {
		return false
	}
	done, ok := hook.Trigger(w, r)
	if !ok {
		return false
	}
	if done {
		i.removeHook(nodeId, hookId)
	}
	return true
}

func (i *core[M]) CancelHook(nodeId uint64, hookId uint64, err error) {
	hook := i.removeHook(nodeId, hookId)
	if hook == nil {
		return
	}
	hook.Cancel(err)
}

func (i *core[M]) CancelHooks(nodeId uint64, err error) {
	i.hooksMu.Lock()
	hooks, ok := i.hooks[nodeId]
	if !ok {
		i.hooksMu.Unlock()
		return
	}
	delete(i.hooks, nodeId)
	i.hooksMu.Unlock()
	for id := range hooks {
		hooks[id].Cancel(err)
	}
}

func (i *core[M]) RegisterHook(nodeId uint64, hookId uint64, hook *node.NodeHook) {
	i.hooksMu.Lock()
	defer i.hooksMu.Unlock()
	hooks, ok := i.hooks[nodeId]
	if !ok {
		hooks = make(map[uint64]*node.NodeHook)
		i.hooks[nodeId] = hooks
	}
	hooks[hookId] = hook
}

func (i *core[M]) getHook(nodeId uint64, hookId uint64) *node.NodeHook {
	i.hooksMu.Lock()
	defer i.hooksMu.Unlock()
	hooks, ok := i.hooks[nodeId]
	if !ok {
		return nil
	}
	hook, ok := hooks[hookId]
	if !ok {
		return nil
	}
	return hook
}

func (i *core[M]) removeHook(nodeId uint64, hookId uint64) *node.NodeHook {
	i.hooksMu.Lock()
	defer i.hooksMu.Unlock()
	hooks, ok := i.hooks[nodeId]
	if !ok {
		return nil
	}
	hook, ok := hooks[hookId]
	if !ok {
		return nil
	}
	delete(hooks, hookId)
	if len(hooks) == 0 {
		delete(i.hooks, nodeId)
	}
	return hook
}
