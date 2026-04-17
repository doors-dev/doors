package shredder

import "sync"

type readWriteFrame struct {
	baseFrame
	thread *ReadWriteThread
	write  bool
	next   *readWriteFrame
}

func (s *readWriteFrame) appendWrite(f *readWriteFrame) *readWriteFrame {
	if s.next == nil {
		s.next = f
		if s.write {
			return nil
		}
		return s
	}
	return s.next.appendWrite(f)
}

func (s *readWriteFrame) onComplete() {
	s.thread.mu.Lock()
	s.thread.head = s.next
	s.thread.mu.Unlock()
	if s.next == nil {
		return
	}
	s.next.activate()
}

func (s *readWriteFrame) getRead() Frame {
	if s.next != nil {
		return s.next.getRead()
	}
	if !s.write {
		return Join(false, &s.baseFrame)
	}
	s.next = s.thread.newFrame(false)
	return s.next.getRead()
}

type ReadWriteThread struct {
	mu   sync.Mutex
	head *readWriteFrame
}

func (r *ReadWriteThread) newFrame(write bool) *readWriteFrame {
	frame := &readWriteFrame{
		thread: r,
		write:  write,
	}
	frame.baseFrame.onComplete = frame.onComplete
	return frame
}

func (r *ReadWriteThread) Read() Frame {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.head == nil {
		frame := r.newFrame(false)
		r.head = frame
		frame.activate()
	}
	return r.head.getRead()
}

func (r *ReadWriteThread) Write() Frame {
	r.mu.Lock()
	frame := r.newFrame(true)
	if r.head == nil {
		r.head = frame
		r.mu.Unlock()
		frame.activate()
		return frame
	}
	frameToRelease := r.head.appendWrite(frame)
	r.mu.Unlock()
	if frameToRelease != nil {
		frameToRelease.Release()
	}
	return frame
}
