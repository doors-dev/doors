// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package path

import (
	"errors"
	"reflect"
	"strings"
)

func newBranch(index int, path string, fields map[string]field) (branch, error) {
	parts := strings.Split(path, "/")
	segments := make([]segment, 0, len(parts))
	for i, part := range parts {
		last := i == len(parts)-1
		name, ok := strings.CutPrefix(part, ":")
		if !ok {
			segments = append(segments, newLiteralSegment(part))
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
				return branch{}, errors.New("path parameter cannot combine '*' with '?' or '+'")
			}
			optional = both
			multiple = both
		}
		if optional && !last {
			return branch{}, errors.New("optional path parameter must be the last segment")
		}
		if multiple && !last {
			return branch{}, errors.New("multi-segment path parameter must be the last segment")
		}
		field, ok := fields[name]
		if !ok {
			return branch{}, errors.New("path parameter field " + name + " was not found among exported compatible fields")
		}
		if multiple {
			multiField, ok := field.multi()
			if !ok {
				return branch{}, errors.New("multi-segment path parameter field " + name + " must have type []string")
			}
			segments = append(segments, newMultiSegment(multiField, optional))
		} else {
			singleField, ok := field.single()
			if !ok {
				return branch{}, errors.New("single-segment path parameter field " + name + " cannot be a slice")
			}
			if optional {
				if !singleField.kind.isPtr() {
					return branch{}, errors.New("optional single-segment path parameter field " + name + " must be a pointer")
				}
			} else {
				if singleField.kind.isPtr() {
					return branch{}, errors.New("required single-segment path parameter field " + name + " must not be a pointer")
				}
			}
			segments = append(segments, newSingleSegment(singleField, optional))
		}
	}
	return branch{
		segments:    segments,
		markerIndex: index,
	}, nil
}

type branch struct {
	markerIndex int
	segments    []segment
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
					return nil, errors.New("no value provided for a required field")
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
					return nil, errors.New("no value provided for a required field")
				}
			}
			parts = append(parts, v...)
			continue
		}
		panic("unknown segment type")
	}
	return parts, nil
}

func (b branch) decode(m reflect.Value, parts []string) bool {
	if len(b.segments) == 0 {
		if len(parts) == 0 {
			return true
		}
		return false
	}
	if len(parts) < len(b.segments)-1 {
		return false
	}
	if len(parts) > len(b.segments) {
		if _, ok := b.segments[len(b.segments)-1].multi(); !ok {
			return false
		}
	}
	for i, segment := range b.segments {
		last := i == len(b.segments)-1
		if s, ok := segment.literal(); ok {
			if len(parts) <= i {
				return false
			}
			if parts[i] != s {
				return false
			}
			continue
		}
		if s, ok := segment.single(); ok {
			if !last && s.optional {
				panic("optional capture can only be the last")
			}
			if len(parts) <= i {
				if !s.optional {
					return false
				}
			} else if !s.set(m, parts[i]) {
				return false
			}
			continue
		}
		if s, ok := segment.multi(); ok {
			if !last {
				panic("multi capture can only be the last")
			}
			if len(parts) <= i {
				if !s.optional {
					return false
				}
			} else {
				s.set(m, parts[i:])
			}
			continue
		}
		panic("unknown segment type")
	}
	return true
}

func (b branch) getMarker(m reflect.Value) bool {
	return m.Field(b.markerIndex).Bool()
}

func (b branch) setMarker(m reflect.Value) {
	m.Field(b.markerIndex).SetBool(true)
}
