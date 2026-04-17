package shredder

import "sync"

type starveFrame struct {
	baseFrame
	thread *ReadStarveWriteThread
	write  bool
	next   *starveFrame
}

func (s *starveFrame) appendWrite(f *starveFrame) *starveFrame {
	if s.next == nil {
		s.next = f
		if s.write {
			return nil
		}
		return s
	}
	return s.next.appendWrite(f)
}

func (s *starveFrame) onComplete() {
	s.thread.mu.Lock()
	s.thread.head = s.next
	s.thread.mu.Unlock()
	if s.next == nil {
		return
	}
	s.next.activate()
}

func (s *starveFrame) getRead() Frame {
	if !s.write {
		return Join(false, &s.baseFrame)
	}
	if s.next == nil {
		s.next = s.thread.newFrame(false)
	}
	return s.next.getRead()
}

type ReadStarveWriteThread struct {
	mu   sync.Mutex
	head *starveFrame
}

func (r *ReadStarveWriteThread) newFrame(write bool) *starveFrame {
	frame := &starveFrame{
		thread: r,
		write:  write,
	}
	frame.baseFrame.onComplete = frame.onComplete
	return frame
}

func (r *ReadStarveWriteThread) Read() Frame {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.head == nil {
		frame := r.newFrame(false)
		r.head = frame
		frame.activate()
	}
	return r.head.getRead()
}

func (r *ReadStarveWriteThread) Write() Frame {
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
