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
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/go-playground/form/v4"
)

type Location struct {
	Path     string
	Query    url.Values
	Fragment string
}

func NewRequestLocation(r *http.Request) Location {
	return Location{
		Path:  r.URL.Path,
		Query: r.URL.Query(),
	}
}

func (l Location) String() string {
	fragment := ""
	if l.Fragment != "" {
		fragment = "#" + l.Fragment
	}
	if len(l.Query) != 0 {
		return fmt.Sprintf("%s?%s%s", l.Path, l.Query.Encode(), fragment)
	}
	return fmt.Sprintf("%s%s", l.Path, fragment)
}

func NewLocationAdapter() Adapter[Location] {
	return locationAdapter{}
}

type locationAdapter struct{}

func (la locationAdapter) Belongs(am any) bool {
	_, match := am.(Location)
	if !match {
		_, match = am.(*Location)
	}
	return match
}

func (la locationAdapter) Decode(l Location) (Location, bool) {
	return l, true
}

// DecodeAny implements [Adapter].
func (la locationAdapter) DecodeAny(l Location) (any, bool) {
	return l, true
}

func (la locationAdapter) Encode(l Location) (Location, error) {
	return l, nil
}

func (la locationAdapter) EncodeAny(am any) (Location, error) {
	m, ok := am.(Location)
	if !ok {
		ref, ok := am.(*Location)
		if !ok {
			return Location{}, errors.New("Model missmatch")
		}
		m = *ref
	}
	return m, nil
}

// Name implements [Adapter].
func (la locationAdapter) Name() string {
	panic("unimplemented")
}

var _ Adapter[Location] = locationAdapter{}

func GetAdapterName(m any) string {
	t := reflect.TypeOf(m)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() == "" {
		panic("Path must be named struct")
	}
	return t.PkgPath() + "." + t.Name()
}

type AnyAdapter interface {
	DecodeAny(Location) (any, bool)
	EncodeAny(any) (Location, error)
	Belongs(any) bool
	Name() string
}

type Adapter[M any] interface {
	AnyAdapter
	Decode(Location) (M, bool)
	Encode(M) (Location, error)
}

func NewAdapter[M any]() (Adapter[M], error) {
	var m M
	if _, ok := any(m).(Location); ok {
		return any(locationAdapter{}).(Adapter[M]), nil
	}
	reader := newAdapterBuilder[M]()
	return reader.build()
}

type adapter[M any] struct {
	g    *group
	name string
}

func (a *adapter[M]) Belongs(am any) bool {
	_, match := am.(M)
	if !match {
		_, match = am.(*M)
	}
	return match
}

func (a *adapter[M]) DecodeAny(l Location) (any, bool) {
	return a.Decode(l)
}

func (a *adapter[M]) EncodeAny(am any) (Location, error) {
	m, ok := am.(M)
	if !ok {
		ref, ok := am.(*M)
		if !ok {
			return Location{}, errors.New("Model missmatch")
		}
		m = *ref
	}
	return a.Encode(m)
}

func (a *adapter[M]) Decode(l Location) (M, bool) {
	p := []rune(l.Path)
	if len(p) != 0 && p[0] == '/' {
		p = p[1:]
	}
	if len(p) != 0 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	var m M
	mutations, ok := a.g.decode(p)
	if !ok {
		return m, false
	}
	for _, mut := range mutations {
		err := mut(&m)
		if err != nil {
			//log
			return m, false
		}
	}
	err := queryDecoder.Decode(&m, l.Query)
	if err != nil {
		//log
		return m, false
	}
	return m, true
}

func (a *adapter[M]) Encode(m M) (Location, error) {
	parts, err := a.g.encode(&m)
	if err != nil {
		return Location{}, err
	}
	query, err := queryEncoder.Encode(&m)
	if err != nil {
		return Location{}, err
	}
	builder := strings.Builder{}
	builder.WriteByte('/')
	for _, part := range parts {
		builder.WriteString(part)
	}
	return Location{
		Path:  builder.String(),
		Query: query,
	}, nil
}

func (a *adapter[M]) Name() string {
	return a.name
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
