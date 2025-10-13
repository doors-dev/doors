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
	"log/slog"
	"sort"

	"github.com/a-h/templ"
)

func NewAttrs() *Attrs {
	return &Attrs{}
}

type Attrs struct {
	regular templ.Attributes
	objects map[string]any
	arrays  map[string][]any
}

func (a *Attrs) Items() []templ.KeyValue[string, any] {
	totalLen := len(a.regular) + len(a.objects) + len(a.arrays)
	items := make([]templ.KeyValue[string, any], totalLen)
	i := 0
	for key, value := range a.regular {
		if len(a.objects) != 0 {
			if _, has := a.objects[key]; has {
				totalLen -= 1
				continue
			}
		}
		if len(a.arrays) != 0 {
			if _, has := a.arrays[key]; has {
				totalLen -= 1
				continue
			}
		}
		items[i] = templ.KeyValue[string, any]{Key: key, Value: value}
		i++
	}
	for key, obj := range a.objects {
		value, err := a.marshal(obj)
		if err != nil {
			slog.Error("object attribute marshaling err", slog.String("json_error", err.Error()), slog.String("attr_name", key))
			totalLen -= 1
			continue
		}
		items[i] = templ.KeyValue[string, any]{Key: key, Value: value}
		i++
	}
	for key, arr := range a.arrays {
		value, err := a.marshal(arr)
		if err != nil {
			slog.Error("array attribute marshaling err", slog.String("json_error", err.Error()), slog.String("attr_name", key))
			totalLen -= 1
			continue
		}
		items[i] = templ.KeyValue[string, any]{Key: key, Value: value}
		i++
	}
	items = items[:totalLen]
	sort.Slice(items, func(i, j int) bool {
		return items[i].Key < items[j].Key
	})
	return items
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

func (a *Attrs) SetRaw(attrs templ.Attributes) {
	for name, value := range attrs {
		a.Set(name, value)
	}
}

func (a *Attrs) joinClass(name string, value any) bool {
	if name != "class" {
		return false
	}
	str, ok := value.(string)
	if !ok {
		return false
	}
	existing, ok := a.regular[name]
	if !ok {
		return false
	}
	existingStr, ok := existing.(string)
	if !ok {
		return false
	}
	a.regular[name] = existingStr + " " + str
	return true
}

func (a *Attrs) Set(name string, value any) {
	if a.regular == nil {
		a.regular = make(templ.Attributes)
	}
	if a.joinClass(name, value) {
		return
	}
	a.regular[name] = value
}

func (a *Attrs) SetObject(name string, value any) {
	if a.objects == nil {
		a.objects = make(map[string]any)
	}
	a.objects[name] = value
}

func (a *Attrs) AppendArray(name string, value any) {
	if a.arrays == nil {
		a.arrays = make(map[string][]any)
	}
	arr, ok := a.arrays[name]
	if !ok {
		arr = []any{value}
	} else {
		arr = append(arr, value)
	}
	a.arrays[name] = arr
}
