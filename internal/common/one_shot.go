// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package common

import "sync/atomic"


type Valve atomic.Pointer[func()]

func (t *Valve) pointer() *atomic.Pointer[func()] {
	return (*atomic.Pointer[func()])(t)
}

var noop = func(){}

func (t *Valve) Reset() {
	t.pointer().Swap(nil)
}

func (t *Valve) Open() {
	f := t.pointer().Swap(&noop)
	if f == nil {
		return
	}
	(*f)()
}

func (t *Valve) Put(f func()) {
	state := t.pointer().Swap(&f)
	if state == nil {
		return
	}
	f()
}
