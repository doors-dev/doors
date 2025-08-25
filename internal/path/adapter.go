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
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/go-playground/form/v4"
)

type Location struct {
	Path  string
	Query url.Values
}


func NewRequestLocation(r *http.Request) *Location {
	return &Location{
		Path:  r.URL.Path,
		Query: r.URL.Query(),
	}
}

func (l Location) String() string {
	if len(l.Query) != 0 {
		return fmt.Sprintf("%s?%s", l.Path, l.Query.Encode())
	}
	return l.Path
}

func GetAdapterName(m any) string {
	t := reflect.TypeOf(m)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() == "" {
		log.Fatalf("Can't use anonimus struct")
	}
	return t.PkgPath() + "." + t.Name()
}

type AnyAdapter interface {
	DecodeAny(*Location) (any, bool)
	EncodeAny(any) (*Location, error)
	Belongs(any) bool
	GetName() string
}

func NewAdapter[M any]() (*Adapter[M], error) {
	reader := newAdapterBuilder[M]()
	return reader.build()
}

type Adapter[M any] struct {
	g    *group
	name string
}

func (a *Adapter[M]) Belongs(am any) bool {
	_, match := am.(*M)
	if match {
		return true
	}
	_, match = am.(M)
	return match
}

func (a *Adapter[M]) DecodeAny(l *Location) (any, bool) {
	return a.Decode(l)
}

func (a *Adapter[M]) GetRef(ma any) (*M, bool) {
	m, ok := ma.(*M)
	if !ok {
		direct, ok := ma.(M)
		if !ok {
			return nil, false
		}
		m = &direct
	}
	return m, true
}

func (a *Adapter[M]) EncodeAny(am any) (*Location, error) {
	m, ok := a.GetRef(am)
	if !ok {
		return nil, errors.New("Model missmatch")
	}
	return a.Encode(m)
}

func (a *Adapter[M]) Decode(l *Location) (*M, bool) {
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
		return nil, false
	}
	for _, mut := range mutations {
		err := mut(&m)
		if err != nil {
			//log
			return nil, false
		}
	}
	err := queryDecoder.Decode(&m, l.Query)
	if err != nil {
		//log
		return nil, false
	}
	return &m, true
}

func (a *Adapter[M]) Encode(m *M) (*Location, error) {
	parts, err := a.g.encode(m)
	if err != nil {
		return nil, err
	}
	query, err := queryEncoder.Encode(m)
	for key := range query {
		if len(query[key]) == 0 || (len(query[key]) == 1 && query[key][0] == "") {
			delete(query, key)
		}
	}
	if err != nil {
		return nil, err
	}
	return &Location{
		Path:  strings.Join(append([]string{"/"}, parts...), ""),
		Query: query,
	}, nil
}

func (a *Adapter[M]) GetName() string {
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
