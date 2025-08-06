package beam

import (
	"context"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/node"
)

type Cancel = func()

type isDone = bool

type Beam[T any] interface {
	// Sub subscribes to the value stream. The onValue callback is called immediately
	// with the current value (in the same goroutine), and again on every update.
	//
	// The subscription continues until:
	//   - The context is canceled
	//   - The onValue function returns true (indicating done)
	//
	// Returns true if the subscription was successfully established;
	// false means the context was already canceled.
	Sub(ctx context.Context, onValue func(context.Context, T) bool) bool

	// SubExt is an extended version of Sub that provides additional control.
	// It behaves the same as Sub, but also:
	//   - Accepts an onCancel callback, invoked when the subscription ends due to context cancellation
	//   - Returns a Cancel function for manual subscription termination
	//
	// Returns the Cancel function and a boolean indicating whether the subscription was established.
	SubExt(ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (Cancel, bool)

	// ReadAndSub returns the current value and then subscribes to future updates.
	// The onValue function is invoked on every subsequent update.
	//
	// Returns the initial value and a boolean:
	//   - If true, the value is valid and subscription was established
	//   - If false, the context was canceled and the returned value is undefined
	ReadAndSub(ctx context.Context, onValue func(context.Context, T) bool) (T, bool)

	// ReadAndSubExt behaves like ReadAndSub with extended control options.
	// It provides the same functionality as ReadAndSub, but also:
	//   - Accepts an onCancel callback for handling cancellation events
	//   - Returns a Cancel function for manual termination
	//
	// Returns the initial value, Cancel function, and success boolean.
	// If the boolean is false, the value is undefined and no subscription was established.
	ReadAndSubExt(ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (T, Cancel, bool)

	// Read returns the current value of the Beam without establishing a subscription.
	//
	// Returns the current value and a boolean:
	//   - If true, the value is valid
	//   - If false, the context was canceled and the value is undefined
	Read(ctx context.Context) (T, bool)

	// AddWatcher attaches a Watcher for full lifecycle control over subscription events.
	// Watchers receive separate callbacks for initialization, updates, and cancellation,
	// allowing for more sophisticated subscription management.
	//
	// Returns a Cancel function and a boolean indicating whether the watcher was added.
	AddWatcher(ctx context.Context, w Watcher[T]) (Cancel, bool)

	addWatcher(ctx context.Context, w node.Watcher) bool
	sync(uint, *common.FuncCollector) (*T, bool)
}

func NewBeamExt[T any, T2 any](source Beam[T], cast func(T) T2, distinct func(new T2, old T2) bool) Beam[T2] {
	return &beam[T, T2]{
		source: source,
		values: make(map[uint]*entry[T2]),
		mu:     sync.Mutex{},
		cast: func(v *T) *T2 {
			v2 := cast(*v)
			return &v2
		},
		upd: distinct,
	}
}

func NewBeam[T any, T2 comparable](source Beam[T], cast func(T) T2) Beam[T2] {
	upd := func(new T2, old T2) bool {
		return new != old
	}
	return &beam[T, T2]{
		source: source,
		values: make(map[uint]*entry[T2]),
		mu:     sync.Mutex{},
		cast: func(v *T) *T2 {
			v2 := cast(*v)
			return &v2
		},
		upd: upd,
	}
}

type entry[T any] struct {
	val     *T
	updated bool
}

type beam[T any, T2 any] struct {
	source Beam[T]
	values map[uint]*entry[T2]
	mu     sync.Mutex
	cast   func(*T) *T2
	upd    func(new T2, old T2) bool
	null   T2
}

func (b *beam[T, T2]) addWatcher(ctx context.Context, w node.Watcher) bool {
	return b.source.addWatcher(ctx, w)
}

func (b *beam[T, T2]) syncEntry(seq uint, c *common.FuncCollector) *entry[T2] {
	e, has := b.values[seq]
	if has {
		return e
	}
	if c != nil {
		c.Add(func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			for s := range b.values {
				if s < seq {
					delete(b.values, s)
				}
			}
		})
	}
	sourceVal, updated := b.source.sync(seq, c)
	if sourceVal == nil {
		return nil
	}
	if !updated {
		prevEntry, has := b.values[seq-1]
		if has {
			return &entry[T2]{
				val:     prevEntry.val,
				updated: false,
			}
		}
		return &entry[T2]{
			val:     b.cast(sourceVal),
			updated: false,
		}
	}
	newVal := b.cast(sourceVal)
	if b.upd == nil {
		return &entry[T2]{
			val:     newVal,
			updated: true,
		}
	}
	var prevVal *T2 = nil
	prevEntry, has := b.values[seq-1]
	if has {
		prevVal = prevEntry.val
	} else {
		sourcePrevVal, _ := b.source.sync(seq-1, nil)
		if sourcePrevVal != nil {
			prevVal = b.cast(sourcePrevVal)
		}
	}
	if prevVal == nil {
		return &entry[T2]{
			val:     newVal,
			updated: true,
		}
	}
	if !b.upd(*newVal, *prevVal) {
		return &entry[T2]{
			val:     prevVal,
			updated: false,
		}
	}
	return &entry[T2]{
		val:     newVal,
		updated: true,
	}

}

func (b *beam[T, T2]) sync(seq uint, c *common.FuncCollector) (*T2, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	entry := b.syncEntry(seq, c)
	if entry == nil {
		return nil, false
	}
	b.values[seq] = entry

	return entry.val, entry.updated
}
