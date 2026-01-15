// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package common

func NewSet[T comparable]() Set[T] {
	return Set[T](make(map[T]struct{}))
}

type Set[T comparable] map[T]struct{}

func (s Set[T]) Slice() []T {
	slice := make([]T, len(s))
	i := 0
	for v := range s {
		slice[i] = v
		i += 1
	}
	return slice
}

func (s Set[T]) Iter() map[T]struct{} {
	return s
}

func (s Set[T]) Len() int {
	return len(s)
}

func (s Set[T]) IsEmpty() bool {
	return s.Len() == 0
}

func (s Set[T]) Has(v T) bool {
	_, has := s[v]
	return has
}

func (s Set[T]) Add(v T) bool {
	if s.Has(v) {
		return false
	}
	s[v] = struct{}{}
	return true
}

func (s Set[T]) Remove(v T) bool {
	if !s.Has(v) {
		return false
	}
	delete(s, v)
	return true
}

func (s Set[T]) Clear() {
	clear(s)
}
