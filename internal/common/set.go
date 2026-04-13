// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
