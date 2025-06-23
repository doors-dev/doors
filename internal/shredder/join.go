package shredder

func R(t *Thread) *JoinedThread {
	return &JoinedThread{
		write:  false,
		thread: t,
	}
}

func W(t *Thread) *JoinedThread {
	return &JoinedThread{
		write:  true,
		thread: t,
	}
}

type JoinedThread struct {
	write  bool
	thread *Thread
}
