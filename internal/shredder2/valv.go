// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package shredder2

import (
	"sync"
)

type valveState int

const (
	closedValve valveState = iota
	openedValve
	brokenValve
)

type Valv struct {
	mu    sync.Mutex
	state valveState
	tasks []func(bool)
}

func (v *Valv) Put(f func(bool)) {
	v.mu.Lock()
	defer v.mu.Unlock()
	switch v.state {
	case closedValve:
		v.tasks = append(v.tasks, f)
	case openedValve:
		f(true)
	case brokenValve:
		f(false)
	}
}

func (v *Valv) Open() {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.state != closedValve {
		panic("valve is not in closed state")
	}
	v.state = openedValve
	for i, f := range v.tasks {
		f(true)
		v.tasks[i] = nil
	}
	v.tasks = nil
}

func (v *Valv) Break() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.state = brokenValve
	for _, f := range v.tasks {
		f(false)
	}
}

func (v *Valv) Reset() {
	v.mu.Lock()
	defer v.mu.Unlock()
	if len(v.tasks) != 0 {
		panic("can't reset valve under pressure")
	}
	v.state = closedValve
}
