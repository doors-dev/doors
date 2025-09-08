// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package common

import (
	"github.com/a-h/templ"
	"log/slog"
)


func NewAttrs() *Attrs {
	return &Attrs{
		regular: make(templ.Attributes),
		objects: make(map[string]any),
		arrays:  make(map[string][]any),
	}
}

type Attrs struct {
	regular templ.Attributes
	objects map[string]any
	arrays  map[string][]any
}

func (a *Attrs) Items() []templ.KeyValue[string, any] {
	return a.a().Items()
}

func (a *Attrs) Join(attrs *Attrs) {
	for name := range attrs.regular {
		a.Set(name, attrs.regular[name])
	}
	for name := range attrs.objects {
		a.SetObject(name, attrs.objects[name])
	}
	for name := range attrs.arrays {
		for _, v := range attrs.arrays[name] {
			a.AppendArray(name, v)
		}
	}
}

func (a *Attrs) marshal(value any) (string, error) {
	b, err := MarshalJSON(value)
	if err != nil {
		return "", err
	}
	return AsString(&b), nil
}

func (a *Attrs) a() templ.Attributes {
	output := make(templ.Attributes)
	for name := range a.regular {
		output[name] = a.regular[name]
	}
	for name := range a.objects {
		s, err := a.marshal(a.objects[name])
		if err != nil {
			slog.Error("object attribute marshaling err", slog.String("json_error", err.Error()), slog.String("attr_name", name))
			continue
		}
		output[name] = s
	}
	for name := range a.arrays {
		s, err := a.marshal(a.arrays[name])
		if err != nil {
			slog.Error("array attribute marshaling err", slog.String("json_error", err.Error()), slog.String("attr_name", name))
			continue
		}
		output[name] = s
	}
	return output
}

func (a *Attrs) SetRaw(attrs templ.Attributes) {
	for name := range attrs {
		a.regular[name] = attrs[name]
	}
}

func (a *Attrs) Set(name string, value any) {
	a.regular[name] = value
}

func (a *Attrs) SetObject(name string, value any) {
	a.objects[name] = value
}

func (a *Attrs) AppendArray(name string, value any) {
	arr, ok := a.arrays[name]
	if !ok {
		arr = []any{value}
	} else {
		arr = append(arr, value)
	}
	a.arrays[name] = arr
}
