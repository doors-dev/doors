// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package shredder



func R(t *Thread) *JoinedThread {
	return &JoinedThread{
		mode: joinRead,
		thread: t,
	}
}

func W(t *Thread) *JoinedThread {
	return &JoinedThread{
		mode: joinWrite,
		thread: t,
	}
}
func WS(t *Thread) *JoinedThread {
	return &JoinedThread{
		mode: joinWriteStarve,
		thread: t,
	}
}

type joinMode int

const (
	joinRead joinMode = iota
	joinWrite
	joinWriteStarve
)

type JoinedThread struct {
	mode   joinMode
	thread *Thread
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
