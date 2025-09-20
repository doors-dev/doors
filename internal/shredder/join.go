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
