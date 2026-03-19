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

type pathVariant struct {
	index   int
	pattern string
}

type adapterBuilder[M any] struct {
	value  M
	path   []pathVariant
	fields map[string]field
}

func (a adapterBuilder[M]) build() (adapter[M], error) {
	if err := a.scanFields(); err != nil {
		return nil, err
	}
	if len(a.path) == 0 {
		return nil, errors.New("no path patterns provided in the path model struct")
	}
	branches := make([]branch, 0, len(a.path))
	for _, path := range a.path {
		branch, err := newBranch(path.index, path.pattern, a.fields)
		if err != nil {
			return nil, err
		}
		branches = append(branches, branch)
	}
	return adapter[M](branches), nil
}

func (a *adapterBuilder[M]) scanFields() error {
	t := reflect.TypeOf(a.value)
	if t.Kind() != reflect.Struct {
		return errors.New("Model must be struct")
	}
	for i := range t.NumField() {
		f := t.Field(i)
		path, ok := f.Tag.Lookup("path")
		if ok {
			if !f.IsExported() {
				return errors.New("Path field " + f.Name + " is not exported")
			}
			if f.Type.Kind() != reflect.Bool {
				return errors.New("Path field must be of bool type")
			}
			a.addPath(f, path)
			continue
		}
		_, ok = f.Tag.Lookup("query")
		if ok {
			if !f.IsExported() {
				return errors.New("Query field " + f.Name + " is not exported")
			}
		}
		a.addField(f, i)
	}
	return nil
}

func (a *adapterBuilder[M]) addField(f reflect.StructField, index int) {
	if !f.IsExported() {
		return
	}
	var kind fieldKind
	switch f.Type.Kind() {
	case reflect.Slice:
		if f.Type.Elem().Kind() != reflect.String {
			return
		}
		a.fields[f.Name] = newMultiField(index)
		return
	case reflect.Ptr:
		switch f.Type.Elem().Kind() {
		case reflect.String:
			kind = kindStringPtr
		case reflect.Int, reflect.Int64:
			kind = kindIntPtr
		case reflect.Float64:
			kind = kindFloatPtr
		case reflect.Uint, reflect.Uint64:
			kind = kindUintPtr
		default:
			return
		}
	case reflect.String:
		kind = kindString
	case reflect.Int, reflect.Int64:
		kind = kindInt
	case reflect.Float64:
		kind = kindFloat
	case reflect.Uint, reflect.Uint64:
		kind = kindUint
	default:
		return
	}
	a.fields[f.Name] = newSingleField(index, kind)
}

func (a *adapterBuilder[M]) addPath(f reflect.StructField, path string) {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")
	a.path = append(a.path, pathVariant{
		index:   f.Index[0],
		pattern: path,
	})
}
