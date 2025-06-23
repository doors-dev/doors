package common


func NewSet[T comparable]() Set[T] {
	return Set[T]{
		m: make(map[T]struct{}),
	}
}

type Set[T comparable] struct {
	m map[T]struct{}
}

func (s *Set[T]) Slice() []T {
	slice := make([]T, len(s.m))
	i := 0
	for v := range s.m {
		slice[i] = v
		i += 1
	}
	return slice
}

func (s *Set[T]) Iter() map[T]struct{} {
	return s.m
}

func (s *Set[T]) Len() int {
	return len(s.m)
}

func (s *Set[T]) IsEmpty() bool {
	return s.Len() == 0
}

func (s *Set[T]) Has(v T) bool {
	_, has := s.m[v]
	return has
}

func (s *Set[T]) Add(v T) bool {
	if s.Has(v) {
		return false
	}
	s.m[v] = struct{}{}
	return true
}

func (s *Set[T]) Remove(v T) bool {
	if !s.Has(v) {
		return false
	}
	delete(s.m, v)
	return true
}
