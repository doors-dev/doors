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
	return Location{}, errors.New("Adepter not found for the provided model")
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
		fields: make(map[string]field),
	}.build()
}

type adapter[M any] []branch

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
	for _, branch := range a {
		var model M
		v := reflect.ValueOf(&model).Elem()
		if branch.decode(v, l.Segments) {
			branch.setMarker(v)
			if err := queryDecoder.Decode(&model, l.Query); err != nil {
				return nil, false
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
	for _, b := range a {
		if len(a) != 1 && !b.getMarker(v) {
			continue
		}
		segments, err := b.encode(v)
		if err != nil {
			return Location{}, err
		}
		query, err := queryEncoder.Encode(model)
		if err != nil {
			return Location{}, err
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
