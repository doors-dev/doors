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

package doors

// JoinScopes returns one [Scopes] value that applies every value in order.
// Without arguments it returns nil.
func JoinScopes(values ...Scopes) Scopes {
	return Join(values...)
}

// JoinActions returns one [Actions] value that runs every value in order.
// Without arguments it returns nil.
func JoinActions(values ...Actions) Actions {
	return Join(values...)
}

// JoinIndicators returns one [Indicators] value that applies every value in
// order. Without arguments it returns nil.
func JoinIndicators(values ...Indicators) Indicators {
	return Join(values...)
}

// Join combines values that support [Joiner.And] into one value, in argument
// order.
//
// It returns the zero value when called without arguments.
func Join[T Joiner[T]](values ...T) T {
	if len(values) == 0 {
		var zero T
		return zero
	}
	result := values[0]
	for _, v := range values[1:] {
		result = result.And(v)
	}
	return result
}

// Joiner is implemented by values that combine with another value of the same
// type, applying the receiver first and then the argument.
type Joiner[T any] interface {
	And(T) T
}
