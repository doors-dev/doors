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

package path

import "reflect"

func newLiteralSegment(value string) segment {
	return segment{
		entity: literalSegment(value),
	}
}

func newSingleSegment(field singleField, optional bool) segment {
	return segment{
		entity: singleSegment{
			field:    field,
			optional: optional,
		},
	}
}

func newMultiSegment(field multiField, optional bool) segment {
	return segment{
		entity: multiSegment{
			field:    field,
			optional: optional,
		},
	}
}

type segment struct {
	entity any
}

func (a segment) literal() (string, bool) {
	c, ok := a.entity.(literalSegment)
	return string(c), ok
}

func (a segment) single() (singleSegment, bool) {
	c, ok := a.entity.(singleSegment)
	return c, ok
}

func (a segment) multi() (multiSegment, bool) {
	c, ok := a.entity.(multiSegment)
	return c, ok
}

type literalSegment string

type singleSegment struct {
	field    singleField
	optional bool
}

func (c singleSegment) get(m reflect.Value) (string, bool) {
	return c.field.get(m)
}

func (c singleSegment) set(m reflect.Value, v string) bool {
	return c.field.set(m, v)
}

type multiSegment struct {
	field    multiField
	optional bool
}

func (c multiSegment) get(m reflect.Value) []string {
	return c.field.get(m)
}

func (c multiSegment) set(m reflect.Value, v []string) {
	if len(v) == 0 {
		if !c.optional {
			panic("non-option field can't receive empty value")
		}
	}
	c.field.set(m, v)
}
