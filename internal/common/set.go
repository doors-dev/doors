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

import (
	"cmp"
	"iter"
	"slices"
)

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

type orderedMapEntry[K cmp.Ordered, V any] struct {
	key   K
	value V
}

type OrderedMap[K cmp.Ordered, V any] []orderedMapEntry[K, V]

func (s *OrderedMap[K, V]) Get(k K) (V, bool) {
	index, ok := slices.BinarySearchFunc(*s, k, func(e orderedMapEntry[K, V], k K) int {
		return cmp.Compare(e.key, k)
	})
	if !ok {
		return *new(V), false
	}
	return (*s)[index].value, true
}

func (s *OrderedMap[K, V]) Set(k K, v V) {
	index, ok := slices.BinarySearchFunc(*s, k, func(e orderedMapEntry[K, V], k K) int {
		return cmp.Compare(e.key, k)
	})
	if ok {
		(*s)[index] = orderedMapEntry[K, V]{k, v}
	} else {
		*s = slices.Insert(*s, index, orderedMapEntry[K, V]{k, v})
	}
}

func (s OrderedMap[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, e := range s {
			if !yield(e.key, e.value) {
				return
			}
		}
	}
}

func (s *OrderedMap[K, V]) Delete(k K) bool {
	index, ok := slices.BinarySearchFunc(*s, k, func(e orderedMapEntry[K, V], k K) int {
		return cmp.Compare(e.key, k)
	})
	if !ok {
		return false
	}
	*s = slices.Delete(*s, index, index+1)
	return true
}

func (s OrderedMap[K, V]) Len() int {
	return len(s)
}
