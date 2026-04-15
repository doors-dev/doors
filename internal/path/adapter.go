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

	"github.com/go-playground/form/v4"
)

type Adapters []AnyAdapter

func (as *Adapters) Add(a AnyAdapter) {
	*as = append(*as, a)
}

func (as Adapters) Encode(v any) (Location, error) {
	for _, a := range as {
		l, err, match := a.EncodeAny(v)
		if !match {
			continue
		}
		if err != nil {
			return Location{}, err
		}
		return l, nil
	}
	return Location{}, errors.New("adapter not found for the provided model")
}

type AnyAdapter interface {
	EncodeAny(any) (Location, error, bool)
}

type Adapter[M any] interface {
	AnyAdapter
	Decode(any) (*M, bool)
	Encode(model *M) (Location, error)
	Assert(any) (*M, bool)
}

func NewAdapter[M any]() (Adapter[M], error) {
	return adapterBuilder[M]{
		fields:   make(map[string]field),
		queryField: -1,
	}.build()
}

type adapter[M any] struct {
	branches []branch
	queryField int
}

func (a adapter[M]) DecodeAny(v any) (any, bool) {
	return a.Decode(v)
}

func (a adapter[M]) Assert(v any) (*M, bool) {
	switch v := v.(type) {
	case M:
		return &v, true
	case *M:
		return v, true
	default:
		return nil, false
	}
}

func (a adapter[M]) Decode(v any) (*M, bool) {
	m, ok := a.Assert(v)
	if ok {
		return m, true
	}
	if loc, ok := v.(Location); ok {
		return a.DecodeLocation(loc)
	}
	return nil, false
}

func (a adapter[M]) DecodeLocation(l Location) (*M, bool) {
	for _, branch := range a.branches {
		var model M
		v := reflect.ValueOf(&model).Elem()
		if branch.decode(v, l.Segments) {
			branch.setMarker(v)
			if a.queryField == -1 {
				if err := queryDecoder.Decode(&model, l.Query); err != nil {
					return nil, false
				}
			} else {
				v.Field(a.queryField).Set(reflect.ValueOf(l.Query))
			}
			return &model, true
		}
	}
	return nil, false
}

func (a adapter[M]) EncodeAny(v any) (Location, error, bool) {
	m, ok := a.Assert(v)
	if !ok {
		return Location{}, nil, false
	}
	l, err := a.Encode(m)
	return l, err, true
}

func (a adapter[M]) Encode(model *M) (Location, error) {
	v := reflect.ValueOf(model).Elem()
	for _, b := range a.branches {
		if len(a.branches) != 1 && !b.getMarker(v) {
			continue
		}
		segments, err := b.encode(v)
		if err != nil {
			return Location{}, err
		}
		var query url.Values
		if a.queryField == -1 {
			query, err = queryEncoder.Encode(model)
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
