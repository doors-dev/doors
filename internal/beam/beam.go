package beam

import (
	"context"
	"sync"

	"github.com/doors-dev/doors/internal/node"
	"github.com/doors-dev/doors/internal/shredder"
)

type Cancel = func()

type isDone = bool

type Beam[T any] interface {
	// Sub subscribes to the value stream. The onValue callback is called immediately
	// with the current value (in the same goroutine), and again on every update.
	//
	// The subscription continues as long as:
	//   - The context is not canceled
	//   - The onValue function returns false
	//
	// Returns true if the context is still valid at the time of subscription;
	// false means the subscription did not start and no value will be emitted.
	Sub(ctx context.Context, onValue func(context.Context, T) bool) bool

	// SubExt is an extended version of Sub. It behaves the same, but also:
	//   - Accepts an onCancel callback, which is invoked when the subscription is canceled
	//     due to context cancellation.
	//   - Returns a Cancel function, which can be used to programmatically stop the subscription.
	//
	// Returns the Cancel function and a boolean indicating whether the subscription was started.
	SubExt(ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (Cancel, bool)

	// ReadAndSub first returns the current value, and then subscribes to updates.
	// The onValue function is invoked on every subsequent update, like Sub.
	//
	// Returns the initial value and a boolean:
	//   - If true, the value is valid and subscription was established.
	//   - If false, the value is undefined and subscription did not start (context was canceled).
	ReadAndSub(ctx context.Context, onValue func(context.Context, T) bool) (T, bool)

	// ReadAndSubExt behaves like ReadAndSub, but also:
	//   - Accepts an onCancel callback triggered when the subscription ends due to context cancellation.
	//   - Returns a Cancel function to allow manual termination.
	//
	// Returns the initial value, a Cancel function, and a boolean indicating success.
	// If the boolean is false, the value is undefined and no subscription was established.
	ReadAndSubExt(ctx context.Context, onValue func(context.Context, T) bool, onCancel func()) (T, Cancel, bool)

	// Read returns the current value of the Beam.
	//
	// If the returned boolean is true, the value is valid.
	// If false, the context was already canceled and the value is undefined.
	Read(ctx context.Context) (T, bool)

	// AddWatcher attaches a Watcher with full lifecycle control.
	// Allows handling custom logic on initialization, updates, and cancellation.
	AddWatcher(ctx context.Context, w Watcher[T]) (Cancel, bool)
	addWatcher(ctx context.Context, w node.Watcher) bool
	sync(uint, *shredder.Collector[func()]) (*T, bool)
}

func NewBeamExt[T any, T2 any](source Beam[T], cast func(T) T2, updateIf func(new T2, old T2) bool) Beam[T2] {
	return &beam[T, T2]{
		source: source,
		values: make(map[uint]*entry[T2]),
		mu:     sync.Mutex{},
		cast: func(v *T) *T2 {
			v2 := cast(*v)
			return &v2
		},
		upd: func(new *T2, old *T2) bool {
			if updateIf == nil {
				return true
			}
			return updateIf(*new, *old)
		},
	}
}

func NewBeam[T any, T2 comparable](source Beam[T], cast func(T) T2, distinct bool) Beam[T2] {
	var upd func(*T2, *T2) bool = nil
	if distinct {
		upd = func(new *T2, old *T2) bool {
			return *new != *old
		}
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
	upd    func(new *T2, old *T2) bool
	null   T2
}

func (b *beam[T, T2]) addWatcher(ctx context.Context, w node.Watcher) bool {
	return b.source.addWatcher(ctx, w)
}

func (b *beam[T, T2]) syncEntry(seq uint, c *shredder.Collector[func()]) *entry[T2] {
	e, has := b.values[seq]
	if has {
		return e
	}
	if c != nil {
		c.Put(func() {
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
	if !b.upd(newVal, prevVal) {
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

func (b *beam[T, T2]) sync(seq uint, c *shredder.Collector[func()]) (*T2, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	entry := b.syncEntry(seq, c)
	if entry == nil {
		return nil, false
	}
	b.values[seq] = entry

	return entry.val, entry.updated
}
