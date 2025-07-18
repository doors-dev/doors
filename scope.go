package doors

import (
	"time"

	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
)

type Scope = front.Scope
type ScopeSet = front.ScopeSet

type scopeFunc func(inst instance.Core) *ScopeSet

func (s scopeFunc) Scope(inst instance.Core) *ScopeSet {
	return s(inst)
}

type BlockingScope struct {
	id front.ScopeAutoId
}

func ScopeBlocking() []Scope {
	return []Scope{&BlockingScope{}}
}

func (b *BlockingScope) Scope(inst instance.Core) *ScopeSet {
	return front.BlockingScope(b.id.Id(inst))
}


type SerialScope struct {
	id front.ScopeAutoId
}

func (b *SerialScope) Scope(inst instance.Core) *ScopeSet {
	return front.SerialScope(b.id.Id(inst))
}

func ScopeSerial() []Scope {
	return []Scope{&SerialScope{}}
}

type LatestScope struct {
	id front.ScopeAutoId
}

func (b *LatestScope) Scope(inst instance.Core) *ScopeSet {
	return front.LatestScope(b.id.Id(inst))
}


func ScopeLatest() []Scope {
	return []Scope{&LatestScope{}}
}

type DebounceScope struct {
	id front.ScopeAutoId
}

func (d *DebounceScope) Scope(duration time.Duration, limit time.Duration) Scope {
	return scopeFunc(func(inst instance.Core) *ScopeSet {
		return front.DebounceScope(d.id.Id(inst), duration, limit)
	})
}


func ScopeDebounce(duration time.Duration, limit time.Duration) []Scope {
	return []Scope{(&DebounceScope{}).Scope(duration, limit)}
}

type FrameScope struct {
	id front.ScopeAutoId
}

func (d *FrameScope) Scope(frame bool) Scope {
	return scopeFunc(func(inst instance.Core) *ScopeSet {
		return front.FrameScope(d.id.Id(inst), frame)
	})
}
