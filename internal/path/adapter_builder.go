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
	fieldIndex   int
	patternIndex int
	pattern      string
}

type adapterBuilder struct {
	sample       any
	path         []pathVariant
	fields       map[string]field
	multiPattern bool
	queryField   int
}

func (a adapterBuilder) build() (adapter, error) {
	if err := a.scanFields(); err != nil {
		return adapter{}, err
	}
	if len(a.path) == 0 {
		return adapter{}, errors.New("no path patterns provided in the path model struct")
	}
	branches := make([]branch, 0, len(a.path))
	for _, path := range a.path {
		branch, err := newBranch(path.fieldIndex, path.patternIndex, path.pattern, a.fields)
		if err != nil {
			return adapter{}, err
		}
		branches = append(branches, branch)
	}
	return adapter{
		branches:   branches,
		queryField: a.queryField,
	}, nil
}

func (a *adapterBuilder) scanFields() error {
	t := reflect.TypeOf(a.sample)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return errors.New("path model must be a struct")
	}
	hasQuery := false
	for i := range t.NumField() {
		f := t.Field(i)
		path, ok := f.Tag.Lookup("path")
		if ok {
			if a.multiPattern {
				return errors.New("only single multipattern is allowed")
			}
			if !f.IsExported() {
				return errors.New("path field " + f.Name + " must be exported")
			}
			if f.Type.Kind() == reflect.Bool {
				a.addPath(f, path, false)
				continue
			}
			if f.Type.Kind() == reflect.Int {
				if len(a.path) != 0 {
					return errors.New("only single multipattern is allowed")
				}
				a.multiPattern = true
				a.addPath(f, path, true)
				continue
			}
			return errors.New("path field " + f.Name + " must have type bool or int")
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

func (a *adapterBuilder) addField(f reflect.StructField, index int) {
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
	case reflect.Pointer:
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

func (a *adapterBuilder) addPath(f reflect.StructField, path string, multi bool) {
	if !multi {
		path = strings.TrimSpace(path)
		path = strings.Trim(path, "/")
		a.path = append(a.path, pathVariant{
			fieldIndex:   f.Index[0],
			patternIndex: -1,
			pattern:      path,
		})
		return
	}
	index := 0
	for variant := range strings.SplitSeq(path, "|") {
		variant = strings.TrimSpace(variant)
		variant = strings.Trim(variant, "/")
		a.path = append(a.path, pathVariant{
			fieldIndex:   f.Index[0],
			patternIndex: index,
			pattern:      variant,
		})
		index += 1
	}
}
