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
	"errors"
	"net/url"
	"reflect"
	"strings"
)

type pathVariant struct {
	index   int
	pattern string
}

type adapterBuilder[M any] struct {
	value    M
	path     []pathVariant
	fields   map[string]field
	queryField int
}

func (a adapterBuilder[M]) build() (Adapter[M], error) {
	var zero M
	if _, ok := any(zero).(Location); ok {
		return any(locationAdapter{}).(Adapter[M]), nil
	}
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
	return adapter[M]{
		branches: branches,
		queryField: a.queryField,
	}, nil
}

func (a *adapterBuilder[M]) scanFields() error {
	t := reflect.TypeOf(a.value)
	if t.Kind() != reflect.Struct {
		return errors.New("path model must be a struct")
	}
	hasQuery := false
	for i := range t.NumField() {
		f := t.Field(i)
		path, ok := f.Tag.Lookup("path")
		if ok {
			if !f.IsExported() {
				return errors.New("path field " + f.Name + " must be exported")
			}
			if f.Type.Kind() != reflect.Bool {
				return errors.New("path field " + f.Name + " must have type bool")
			}
			a.addPath(f, path)
			continue
		}
		if f.Type == reflect.TypeFor[url.Values]() {
			if !f.IsExported() {
				return errors.New("path field " + f.Name + " must be exported")
			}
			a.queryField = i
			continue
		}
		_, ok = f.Tag.Lookup("query")
		if ok {
			hasQuery = true
			if !f.IsExported() {
				return errors.New("query field " + f.Name + " must be exported")
			}
		}
		a.addField(f, i)
	}
	if hasQuery && a.queryField != -1 {
		return errors.New("path struct contains both url.Values field and `query` tagged field, you can't have both")
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
