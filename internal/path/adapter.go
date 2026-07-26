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
	"log/slog"
	"net/url"
	"reflect"
	"sync"

	"github.com/go-playground/form/v4"
)

type modelAdapters sync.Map

var adapters modelAdapters

func (a *modelAdapters) get(t reflect.Type) (adapter, bool) {
	entry, ok := (*sync.Map)(a).Load(t)
	if !ok {
		return adapter{}, false
	}
	return entry.(adapter), true
}

func (a *modelAdapters) set(t reflect.Type, ad adapter) {
	(*sync.Map)(a).Store(t, ad)
}

// Encode returns the [Location] for m.
//
// m may be a Location, a value implementing [Encoder], a path model struct, or
// a pointer to one. It returns an error if m is none of these, if its path tags
// are invalid, or if it does not encode to a complete path.
func Encode(m any) (Location, error) {
	switch m := (any)(m).(type) {
	case Location:
		return m, nil
	case Encoder:
		return m.Encode()
	}
	adapter, err := get(m)
	if err != nil {
		return Location{}, err
	}
	return adapter.encode(m)
}

// GetModelAdapter returns the [ModelAdapter] for path model M. It returns an
// error if M is not a struct or its path tags are missing or invalid.
func GetModelAdapter[M any]() (ModelAdapter[M], error) {
	adapter, err := get(new(M))
	if err != nil {
		return ModelAdapter[M]{}, err
	}
	return (ModelAdapter[M])(adapter), nil
}

func get(sample any) (adapter, error) {
	t := reflect.TypeOf(sample)
	if t == nil {
		return adapter{}, errors.New("model must be struct")
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return adapter{}, errors.New("model must be struct")
	}
	adapter, ok := adapters.get(t)
	if ok {
		return adapter, nil
	}
	adapter, err := adapterBuilder{
		sample:     sample,
		fields:     make(map[string]field),
		queryField: -1,
	}.build()
	if err != nil {
		return adapter, err
	}
	adapters.set(t, adapter)
	return adapter, nil
}

// Encoder is implemented by models that build their own [Location] instead of
// describing their path with tags.
type Encoder interface {
	// Encode returns the location the model points to.
	Encode() (Location, error)
}

type AnyModelAdapter interface {
	EncodeAny(any) (Location, error, bool)
}

// ModelAdapter decodes locations into path model M and encodes M back.
type ModelAdapter[M any] adapter

// Decode returns the model decoded from l, and false if l matches no path
// variant of M.
func (ma ModelAdapter[M]) Decode(l Location) (*M, bool) {
	m := new(M)
	if (adapter)(ma).decode(l, m) {
		return m, true
	}
	return nil, false
}

// Encode returns the location for m, or an error if m selects no path variant
// or leaves a path capture empty.
func (ma ModelAdapter[M]) Encode(m *M) (Location, error) {
	return (adapter)(ma).encode(m)
}

type adapter struct {
	prefix     *branch
	branches   []fieldBranch
	queryField int
}

func (a adapter) decode(l Location, ref any) bool {
	v := reflect.ValueOf(ref).Elem()
	segments := l.Segments
	var prefixMutations []func()
	if a.prefix != nil {
		var ok bool
		if prefixMutations, ok = a.prefix.decode(v, segments); !ok {
			return false
		}
		segments = segments[a.prefix.len():]
	}
	for _, branch := range a.branches {
		mutations, ok := branch.decode(v, segments)
		if !ok {
			continue
		}
		mutations = append(mutations, branch.setMarker(v))
		if a.queryField == -1 {
			mutations = append(mutations, func() {
				if err := queryDecoder.Decode(ref, l.Query); err != nil {
					slog.Error("Query decoding error", "error", err)
				}
			})
		} else {
			query := cloneQuery(l.Query)
			mutations = append(mutations, func() {
				v.Field(a.queryField).Set(reflect.ValueOf(query))
			})
		}
		for _, mutation := range prefixMutations {
			mutation()
		}
		for _, mutation := range mutations {
			mutation()
		}
		return true
	}
	return false
}

func (a adapter) encode(m any) (Location, error) {
	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	var prefixSegments []string
	if a.prefix != nil {
		var err error
		if prefixSegments, err = a.prefix.encode(v); err != nil {
			return Location{}, err
		}
	}
	for _, b := range a.branches {
		if len(a.branches) != 1 && !b.getMarker(v) {
			continue
		}
		segments, err := b.encode(v)
		if err != nil {
			return Location{}, err
		}
		if len(prefixSegments) != 0 {
			segments = append(prefixSegments, segments...)
		}
		var query url.Values
		if a.queryField == -1 {
			query, err = queryEncoder.Encode(m)
			if err != nil {
				return Location{}, err
			}
		} else {
			query = v.Field(a.queryField).Interface().(url.Values)
		}
		return Location{
			Segments: segments,
			Query:    query,
		}, nil
	}
	return Location{}, errors.New("no path variant selected")
}

var queryDecoder *form.Decoder
var queryEncoder *form.Encoder

func init() {
	queryDecoder = form.NewDecoder()
	queryDecoder.SetMode(form.ModeExplicit)
	queryDecoder.SetTagName("query")
	queryEncoder = form.NewEncoder()
	queryEncoder.SetMode(form.ModeExplicit)
	queryEncoder.SetTagName("query")
}
