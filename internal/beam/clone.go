package beam

type Cloner[T any] interface {
	Clone() T
}

type dereferencer[T any] = func(*T) T

func newDereferencer[T any]() dereferencer[T] {
	var t T
	if _, ok := (any)(t).(Cloner[T]); ok {
		return func(t *T) T {
			return (any)(*t).(Cloner[T]).Clone()
		}
	}
	if _, ok := (any)(&t).(Cloner[T]); ok {
		return func(t *T) T {
			return (any)(t).(Cloner[T]).Clone()
		}
	}
	return func(t *T) T {
		return *t
	}
}
