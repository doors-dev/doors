// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package shredder

func R(t *Thread) *JoinedThread {
	return &JoinedThread{
		mode:   joinRead,
		thread: t,
	}
}

func W(t *Thread) *JoinedThread {
	return &JoinedThread{
		mode:   joinWrite,
		thread: t,
	}
}
func Ws(t *Thread) *JoinedThread {
	return &JoinedThread{
		mode:   joinWriteStarve,
		thread: t,
	}
}

func Ri(t *Thread) *JoinedThread {
	return &JoinedThread{
		mode:    joinRead,
		thread:  t,
		instant: true,
	}
}

func Wi(t *Thread) *JoinedThread {
	return &JoinedThread{
		mode:    joinWrite,
		thread:  t,
		instant: true,
	}
}
func Wsi(t *Thread) *JoinedThread {
	return &JoinedThread{
		mode:    joinWriteStarve,
		thread:  t,
		instant: true,
	}
}

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
