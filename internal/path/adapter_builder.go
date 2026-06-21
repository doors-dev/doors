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

	"github.com/doors-dev/doors/internal/common"
)

type pathVariant struct {
	fieldIndex   int
	patternIndex int
	pattern      string
}

type adapterBuilder struct {
	sample       any
	prefix       string
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
	branches := make([]fieldBranch, 0, len(a.path))
	for _, path := range a.path {
		branch, err := newFieldBranch(path.fieldIndex, path.patternIndex, path.pattern, a.fields)
		if err != nil {
			return adapter{}, err
		}
		branches = append(branches, branch)
	}
	var prefix *branch
	if a.prefix != "" {
		branch, err := newPrefixBranch(a.prefix, a.fields)
		if err != nil {
			return adapter{}, err
		}
		prefix = &branch
	}
	return adapter{
		prefix:     prefix,
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
		pathTag := false
		queryTag := false
		for name, value := range common.IterTags(f.Tag) {
			if name == "query" {
				queryTag = true
				continue
			}
			if name == "path" {
				name = "/"
			}
			if !strings.Contains(name, "/") {
				continue
			}
			if a.multiPattern {
				return errors.New("only single multipattern is allowed")
			}
			if !f.IsExported() {
				return errors.New("path field " + f.Name + " must be exported")
			}
			if f.Type.Kind() == reflect.Bool {
				a.addBoolPath(f, name, value)
				pathTag = true
				continue
			}
			if f.Type.Kind() == reflect.Int {
				if len(a.path) != 0 {
					return errors.New("only single multipattern is allowed")
				}
				a.multiPattern = true
				a.addIntPath(f, name, value)
				pathTag = true
				continue
			}
			return errors.New("path field " + f.Name + " must have type bool or int")
		}
		if pathTag {
			continue
		}
		if queryTag {
			hasQuery = true
			if !f.IsExported() {
				return errors.New("query field " + f.Name + " must be exported")
			}
		}
		if f.Type == reflect.TypeFor[url.Values]() {
			if !f.IsExported() {
				return errors.New("path field " + f.Name + " must be exported")
			}
			a.queryField = i
			continue
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
		if f.Type != reflect.TypeFor[[]string]() {
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

func (a *adapterBuilder) addBoolPath(f reflect.StructField, prefix string, path string) {
	a.path = append(a.path, pathVariant{
		fieldIndex:   f.Index[0],
		patternIndex: -1,
		pattern:      join(prefix, path),
	})
}

func (a *adapterBuilder) addIntPath(f reflect.StructField, prefix string, path string) {
	a.prefix = trim(prefix)
	index := 0
	for variant := range strings.SplitSeq(path, "|") {
		a.path = append(a.path, pathVariant{
			fieldIndex:   f.Index[0],
			patternIndex: index,
			pattern:      trim(variant),
		})
		index += 1
	}
}

func join(prefix string, path string) string {
	return trim(trim(prefix) + "/" + trim(path))
}

func trim(path string) string {
	path = strings.TrimSpace(path)
	return strings.Trim(path, "/")
}
