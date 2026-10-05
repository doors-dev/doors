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

package beam

import (
	"context"
)

type Lenser[T any] interface {
	Beamer[T]
	Update(context.Context, T) <-chan error
	Mutate(context.Context, func(T) T) <-chan error
}

type Lens[T1, T2 any] = *lens[T1, T2]

func NewLens[T1 any, T2 any](source Lenser[T1], get func(T1) T2, set func(T1, T2) T1, equal func(new T2, old T2) bool) Lens[T1, T2] {
	if equal == nil {
		equal = NeverEqual
	}
	return &lens[T1, T2]{
		Beam: &beam[T1, T2]{
			beam:   source,
			values: make(map[uint]entry[T2]),
			get:    get,
			equal:  equal,
		},
		source: source,
		set:    set,
	}
}

type lens[T1 any, T2 any] struct {
	Beam[T1, T2]
	source Lenser[T1]
	set    func(T1, T2) T1
}

func (l *lens[T1, T2]) Mutate(ctx context.Context, m func(T2) T2) <-chan error {
	return l.source.Mutate(ctx, func(sourceV T1) T1 {
		v := l.get(sourceV)
		v = m(v)
		return l.set(sourceV, v)
	})
}

func (l *lens[T1, T2]) Update(ctx context.Context, v T2) <-chan error {
	return l.source.Mutate(ctx, func(sourceV T1) T1 {
		return l.set(sourceV, v)
	})
}

var _ Lenser[any] = (*lens[any, any])(nil)
