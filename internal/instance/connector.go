package instance

import (
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/node"
)

type linkInstance interface {
	kill(bool)
}

func newConnector(instance linkInstance, ttl time.Duration) *connector {
	l := &connector{
		mu:       sync.Mutex{},
		calls:    make(map[uint]*call),
		ttl:      ttl,
		instance: instance,
	}
	l.resetKillTimer()
	return l
}

type connector struct {
	ttl        time.Duration
	killTimer  *time.Timer
	touchTimer *time.Timer
	instance   linkInstance
	mu         sync.Mutex
	killed     bool
	conn       common.EventSender
	seq        uint
	sendSeq    uint
	ackSeq     uint
	calls      map[uint]*call
}

func (l *connector) Kill(suspend bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.killed {
		return
	}
	if l.conn != nil {
		if suspend {
			l.conn.Tx(common.SuspendEvent{})
		}
		l.conn.Close()
	}
	if l.killTimer != nil {
		l.killTimer.Stop()
	}
	if l.touchTimer != nil {
		l.touchTimer.Stop()
	}
}

func (l *connector) Connect(conn common.EventSender) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.killed {
		return
	}
	l.resetTouchTimer()
	l.conn = conn
	l.reset(l.sendSeq)
	l.touch()
	l.sendNext()
}

func (l *connector) Call(caller node.Caller) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.killed {
		return
	}
	ready := l.seq == l.sendSeq
	l.seq += 1
	l.calls[l.seq] = &call{
		link:   l,
		seq:    l.seq,
		caller: caller,
	}
	if ready {
		l.sendNext()
	}
}

func (l *connector) CallResponse(resp *CallResponse) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.killed {
		return false
	}
	l.ackSeq = resp.Ack
	if l.sendSeq < l.ackSeq {
		l.sendSeq = l.ackSeq
	}
	if len(resp.Ready) == 0 {
		return true
	}
	if !l.resetKillTimer() {
		return false
	}
	l.resetTouchTimer()
	for _, res := range resp.Ready {
		call, ok := l.calls[res.seq]
		if !ok {
			continue
		}
		delete(l.calls, res.seq)
		call.result(res.err)
	}
	return true
}

func (l *connector) touch() {
	if l.seq != l.sendSeq {
		return
	}
	if l.conn == nil {
		return
	}
	l.seq += 1
	l.calls[l.seq] = &call{
		link:   l,
		seq:    l.seq,
		caller: &touchCaller{},
	}
}

func (l *connector) sendNext() {
	if l.sendSeq == l.seq || l.conn == nil {
		return
	}
	call, has := l.calls[l.sendSeq+1]
	if !has {
		l.sendSeq += 1
		l.sendNext()
		return
	}
	if !l.conn.Tx(call) {
		l.conn = nil
		return
	}
}

func (l *connector) onWrite(seq uint, success bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.killed {
		return
	}
	if success == true {
		l.sendSeq = seq
		l.sendNext()
		return
	}
	l.conn = nil
	l.reset(seq)
}

func (l *connector) reset(seq uint) {
	for i := l.ackSeq + 1; i <= seq; i += 1 {
		call, has := l.calls[i]
		if !has {
			continue
		}
		call.writeErr()
	}
	l.sendSeq = l.ackSeq
}

func (l *connector) cancelCall(seq uint) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.killed {
		return
	}
	delete(l.calls, seq)
	l.sendNext()
}

func (l *connector) resetTouchTimer() {
	d := min(l.ttl / 4, 15 * time.Second)
	if l.touchTimer != nil {
		l.touchTimer.Stop()
		l.touchTimer.Reset(d)
		return
	}
	l.touchTimer = time.AfterFunc(d, func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		if l.killed {
			return
		}
		l.touch()
		l.sendNext()
		l.resetTouchTimer()
	})
}

func (l *connector) resetKillTimer() bool {
	if l.killTimer != nil {
		stopped := l.killTimer.Stop()
		if !stopped {
			return false
		}
		l.killTimer.Reset(l.ttl)
		return true
	}
	l.killTimer = time.AfterFunc(l.ttl, func() {
		l.instance.kill(false)
	})
	return true
}
