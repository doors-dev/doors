// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package path

import (
	"reflect"
	"strconv"
)

func newSingleField(index int, kind fieldKind) field {
	return field{
		entity: singleField{
			index: index,
			kind:  kind,
		},
	}
}

func newMultiField(index int) field {
	return field{
		entity: multiField(index),
	}
}

type field struct {
	entity any
}

func (f field) multi() (multiField, bool) {
	s, ok := f.entity.(multiField)
	return s, ok
}

func (f field) single() (singleField, bool) {
	s, ok := f.entity.(singleField)
	return s, ok
}

type multiField int

func (f multiField) get(m reflect.Value) []string {
	return m.Field(int(f)).Interface().([]string)
}

func (f multiField) set(m reflect.Value, v []string) {
	m.Field(int(f)).Set(reflect.ValueOf(v))
}

type fieldKind int

const (
	kindString fieldKind = iota
	kindUint
	kindInt
	kindFloat
	kindStringPtr
	kindUintPtr
	kindIntPtr
	kindFloatPtr
)

func (f fieldKind) isPtr() bool {
	return f >= kindStringPtr
}

type singleField struct {
	kind  fieldKind
	index int
}

func (f singleField) get(m reflect.Value) (string, bool) {
	field := m.Field(f.index)
	switch f.kind {
	case kindString:
		return field.String(), true
	case kindStringPtr:
		if field.IsNil() {
			return "", false
		}
		return field.Elem().String(), true
	case kindInt:
		return strconv.FormatInt(field.Int(), 10), true
	case kindIntPtr:
		if field.IsNil() {
			return "", false
		}
		return strconv.FormatInt(field.Elem().Int(), 10), true
	case kindUint:
		return strconv.FormatUint(field.Uint(), 10), true

	case kindUintPtr:
		if field.IsNil() {
			return "", false
		}
		return strconv.FormatUint(field.Elem().Uint(), 10), true
	case kindFloat:
		return strconv.FormatFloat(field.Float(), 'g', -1, 64), true
	case kindFloatPtr:
		if field.IsNil() {
			return "", false
		}
		return strconv.FormatFloat(field.Elem().Float(), 'g', -1, 64), true
	default:
		panic("unknown field type")
	}
}

func (f singleField) set(m reflect.Value, v string) bool {
	field := m.Field(f.index)
	switch f.kind {
	case kindString:
		field.SetString(v)
		return true
	case kindStringPtr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field.Elem().SetString(v)
		return true
	case kindInt:
		num, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return false
		}
		field.SetInt(num)
		return true
	case kindIntPtr:
		num, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return false
		}
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field.Elem().SetInt(num)
		return true
	case kindUint:
		num, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return false
		}
		field.SetUint(num)
		return true
	case kindUintPtr:
		num, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return false
		}
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field.Elem().SetUint(num)
		return true
	case kindFloat:
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		field.SetFloat(num)
		return true
	case kindFloatPtr:
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field.Elem().SetFloat(num)
		return true
	default:
		panic("unknown field type")
	}
}
