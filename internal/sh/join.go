// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package sh

type joinMode int

const (
	joinRead joinMode = iota
	joinWrite
	joinWriteStarve
)

type JoinedThread struct {
	mode    joinMode
	thread  *Thread
	instant bool
}

func (j *JoinedThread) start(t task) bool {
	switch j.mode {
	case joinRead:
		return j.thread.readTask(t)
	case joinWrite:
		return j.thread.writeTask(t, false)
	case joinWriteStarve:
		return j.thread.writeTask(t, true)
	}
	panic("wrong task mode")
}
