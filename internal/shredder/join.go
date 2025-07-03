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
