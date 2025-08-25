// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package path

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type field struct {
	i int
	f reflect.StructField
} 

type adapterBuilder[M any] struct {
	g       *group
	fields  map[string]field
	name    string
}

func newAdapterBuilder[M any]() *adapterBuilder[M] {
	return &adapterBuilder[M]{
		g:       newGroup(),
		fields:  make(map[string]field),
		name:    "",
	}
}

func (a *adapterBuilder[M]) processPath(f field, p string) error {
    if f.f.Type.Kind() != reflect.Bool {
        return errors.New("Path field must be of boolean type")
    }
	if !f.f.IsExported() {
		return errors.New("Path field must be exported")
	}
	re := regexp.MustCompile(`\s+`)
	p = re.ReplaceAllString(strings.Trim(p, "/"), "")
    b, err := newBranch(p, f)
	if err != nil {
		return err
	}
    a.g.append(b)
	return nil
}

func (a *adapterBuilder[M]) processQuery(f reflect.StructField) error {
	if !f.IsExported() {
		return errors.New("Query field must be exported")
	}
	return nil
}

func (a *adapterBuilder[M]) readStruct() error {
	var m M
	t := reflect.TypeOf(m)
	if t.Kind() != reflect.Struct {
		return errors.New("Model must be struct")
	}
	pathFound := false
	for i := range t.NumField() {
		f := t.Field(i)
		t := f.Tag
        val, ok := t.Lookup("p")
        if !ok {
            val, ok = t.Lookup("path")
        }
		if ok {
			if !ok {
				val = string(t)
			}
			err := a.processPath(field{f: f, i: i}, val)
			if err != nil {
				return err
			}
			pathFound = true
		}
		q := t.Get("query")
		if q != "" {
			err := a.processQuery(f)
			if err != nil {
				return err
			}
			continue
		}
		if f.IsExported() {
			a.fields[f.Name] = field{
				i: i,
				f: f,
			}
		}
	}
	if !pathFound {
		return errors.New("Path fields not found")
	}
	a.name = GetAdapterName(&m)
	return nil
}


func (a *adapterBuilder[M]) processParams() error {
	params := make(map[string][]*atom)
	a.g.collectParams(params)
	for param := range params {
		field, ok := a.fields[param]
		if !ok {
			return errors.New(fmt.Sprint("Param field not found or not exported", param))
		}
		for _, a := range params[param] {
			capture, err := newCapture(field, a)
			if err != nil {
				return err
			}
			a.setCapture(capture)
		}
	}
	return nil
}

func (a *adapterBuilder[M]) build() (*Adapter[M], error) {
	err := a.readStruct()
	if err != nil {
		return nil, err
	}
	err = a.processParams()
	if err != nil {
		return nil, err
	}
	return &Adapter[M]{
		g:       a.g,
		name:    a.name,
	}, nil
}
