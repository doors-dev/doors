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

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/doors-dev/doors/internal/common"
)

func newFieldBranch(index int, patternIndex int, path string, fields map[string]field) (fieldBranch, error) {
	b, err := newBranch(path, fields, true)
	if err != nil {
		return fieldBranch{}, err
	}
	return fieldBranch{
		fieldIndex:   index,
		patternIndex: patternIndex,
		branch:       b,
	}, nil
}

func newPrefixBranch(path string, fields map[string]field) (branch, error) {
	return newBranch(path, fields, false)
}

func newBranch(path string, fields map[string]field, tail bool) (branch, error) {
	b := branch{
		tail: tail,
	}
	if path == "" {
		return b, nil
	}
	parts := strings.Split(path, "/")
	b.segments = make([]segment, 0, len(parts))
	for i, part := range parts {
		last := tail && i == len(parts)-1
		name, ok := strings.CutPrefix(part, ":")
		if !ok {
			b.segments = append(b.segments, newLiteralSegment(part))
			continue
		}
		var optional bool
		var multiple bool
		name, optional = strings.CutSuffix(name, "?")
		name, multiple = strings.CutSuffix(name, "+")
		if multiple && !optional {
			name, optional = strings.CutSuffix(name, "?")
		}
		var both bool
		name, both = strings.CutSuffix(name, "*")
		if both {
			if optional || multiple {
				return branch{}, fmt.Errorf("%w: path parameter cannot combine '*' with '?' or '+'", common.ErrPathModel)
			}
			optional = both
			multiple = both
		}
		if optional && !last {
			return branch{}, fmt.Errorf("%w: optional path parameter must be the last segment", common.ErrPathModel)
		}
		if multiple && !last {
			return branch{}, fmt.Errorf("%w: multi-segment path parameter must be the last segment", common.ErrPathModel)
		}
		field, ok := fields[name]
		if !ok {
			return branch{}, fmt.Errorf("%w: path parameter %q has no matching exported field; path marker fields cannot be used as captures", common.ErrPathModel, name)
		}
		if multiple {
			multiField, ok := field.multi()
			if !ok {
				return branch{}, fmt.Errorf("%w: multi-segment path parameter field %q must have type []string", common.ErrPathModel, name)
			}
			b.segments = append(b.segments, newMultiSegment(name, multiField, optional))
		} else {
			singleField, ok := field.single()
			if !ok {
				return branch{}, fmt.Errorf("%w: single-segment path parameter field %q cannot be a slice", common.ErrPathModel, name)
			}
			if optional {
				if !singleField.kind.isPtr() {
					return branch{}, fmt.Errorf("%w: optional single-segment path parameter field %q must be a pointer", common.ErrPathModel, name)
				}
			} else {
				if singleField.kind.isPtr() {
					return branch{}, fmt.Errorf("%w: required single-segment path parameter field %q must not be a pointer", common.ErrPathModel, name)
				}
			}
			b.segments = append(b.segments, newSingleSegment(name, singleField, optional))
		}
	}
	return b, nil
}

type fieldBranch struct {
	fieldIndex   int
	patternIndex int
	branch
}

func (b fieldBranch) getMarker(m reflect.Value) bool {
	if b.patternIndex == -1 {
		return m.Field(b.fieldIndex).Bool()
	}
	return m.Field(b.fieldIndex).Int() == int64(b.patternIndex)
}

func (b fieldBranch) setMarker(m reflect.Value) func() {
	if b.patternIndex == -1 {
		return func() {
			m.Field(b.fieldIndex).SetBool(true)
		}
	}
	return func() {
		m.Field(b.fieldIndex).SetInt(int64(b.patternIndex))
	}
}

type branch struct {
	segments []segment
	tail     bool
}

func (b branch) encode(m reflect.Value) ([]string, error) {
	parts := make([]string, 0, len(b.segments))
	for i, segment := range b.segments {
		last := i == len(b.segments)-1
		if s, ok := segment.literal(); ok {
			parts = append(parts, s)
			continue
		}
		if s, ok := segment.single(); ok {
			if !last && s.optional {
				panic("optional capture can only be the last")
			}
			v, ok := s.get(m)
			if !ok {
				if !s.optional {
					return nil, fmt.Errorf("%w: no value provided for required field %q", common.ErrPathEncode, s.name)
				}
				continue
			}
			parts = append(parts, v)
			continue
		}
		if s, ok := segment.multi(); ok {
			if !last {
				panic("multi capture can only be the last")
			}
			v := s.get(m)
			if len(v) == 0 {
				if !s.optional {
					return nil, fmt.Errorf("%w: no value provided for required field %q", common.ErrPathEncode, s.name)
				}
			}
			if slices.Contains(v, "") {
				return nil, fmt.Errorf("%w: empty value in multi-segment field %q", common.ErrPathEncode, s.name)
			}
			parts = append(parts, v...)
			continue
		}
		panic("unknown segment type")
	}
	return parts, nil
}

func (b branch) len() int {
	return len(b.segments)
}

func (b branch) decode(m reflect.Value, parts []string) ([]func(), bool) {
	if len(b.segments) == 0 {
		if len(parts) == 0 {
			return nil, true
		}
		return nil, false
	}
	if len(parts) < len(b.segments)-1 {
		return nil, false
	}
	if b.tail && len(parts) > len(b.segments) {
		if _, ok := b.segments[len(b.segments)-1].multi(); !ok {
			return nil, false
		}
	}
	mutations := make([]func(), 0, len(b.segments))
	for i, segment := range b.segments {
		last := b.tail && i == len(b.segments)-1
		if s, ok := segment.literal(); ok {
			if len(parts) <= i {
				return nil, false
			}
			if parts[i] != s {
				return nil, false
			}
			continue
		}
		if s, ok := segment.single(); ok {
			if !last && s.optional {
				panic("optional capture can only be the last")
			}
			if len(parts) <= i {
				if !s.optional {
					return nil, false
				}
			} else {
				mutation, ok := s.set(m, parts[i])
				if !ok {
					return nil, false
				}
				mutations = append(mutations, mutation)
			}
			continue
		}
		if s, ok := segment.multi(); ok {
			if !last {
				panic("multi capture can only be the last")
			}
			if len(parts) <= i {
				if !s.optional {
					return nil, false
				}
			} else {
				mutations = append(mutations, s.set(m, slices.Clone(parts[i:])))
			}
			continue
		}
		panic("unknown segment type")
	}
	return mutations, true
}
