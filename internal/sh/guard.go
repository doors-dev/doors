package sh

import (
	"slices"
	"sync"
)

type guardTask struct {
	group uint8
	f     func()
}

type Guard struct {
	mu           sync.Mutex
	openedGroups []uint8
	tasks        []guardTask
}

func (v *Guard) Run(group uint8, f func()) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if slices.Contains(v.openedGroups, group) {
		f()
		return
	}
	v.tasks = append(v.tasks, guardTask{
		group: group,
		f:     f,
	})
}

func (v *Guard) Open(group ...uint8) {
	v.mu.Lock()
	defer v.mu.Unlock()
	n := 0
	for _, t := range v.tasks {
		if slices.Contains(group, t.group) {
			v.tasks[n] = t
			n++
			continue
		}
		t.f()
	}
	for i := n; i < len(v.tasks); i++ {
		v.tasks[i] = guardTask{}
	}
	v.tasks = v.tasks[:n]
	v.openedGroups = append(v.openedGroups, group...)
}
