package door2

import "github.com/doors-dev/gox"

type JobStream = *jobStream

func newJobStream(root *pipe) JobStream {
	return &jobStream{
		stack: []*pipe{root},
	}
}

type jobStream struct {
	stack []*pipe
}


func (j JobStream) last() *pipe {
	return j.stack[len(j.stack)-1]
}

func (j JobStream) pop() bool {
	if len(j.stack) == 0 {
		return false
	}
	// last := p.last()
	j.stack[len(j.stack)-1] = nil
	j.stack = j.stack[:len(j.stack)-1]
	return len(j.stack) != 0
}

func (j JobStream) push(q *pipe) {
	j.stack = append(j.stack, q)
}

func (j JobStream) Next() (gox.Job, bool) {
	item, closed := j.last().Get()
	if closed {
		if !j.pop() {
			return nil, false
		}
		return j.Next()
	}
	switch i := item.(type) {
	case *pipe:
		j.push(i)
		return j.Next()
	case gox.Job:
		return i, true
	default:
		panic("invalid type")
	}
}

