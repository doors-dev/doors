// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package path

import (
	"maps"
	"net/url"
	"slices"
	"strings"
)

type Location struct {
	Query    url.Values
	Segments []string
}

func (l Location) Path() string {
	b := strings.Builder{}
	for _, part := range l.Segments {
		b.WriteByte('/')
		b.WriteString(url.PathEscape(part))
	}
	return b.String()
}

func (l Location) String() string {
	b := strings.Builder{}
	for _, part := range l.Segments {
		b.WriteByte('/')
		b.WriteString(url.PathEscape(part))
	}
	if len(l.Query) != 0 {
		b.WriteByte('?')
		b.WriteString(l.Query.Encode())
	}
	return b.String()
}

func EqualLocation(a, b Location) bool {
	return slices.Equal(a.Segments, b.Segments) && maps.EqualFunc(a.Query, b.Query, slices.Equal)
}

func NewLocationFromEscapedURI(s string) (Location, error) {
	u, err := url.Parse(s)
	if err != nil {
		return Location{}, err
	}
	return NewLocationFromURL(u)
}

func NewLocationFromURL(u *url.URL) (Location, error) {
	parts := make([]string, 0)
	trimmed := strings.Trim(u.EscapedPath(), "/")
	for part := range strings.SplitSeq(trimmed, "/") {
		decoded, err := url.PathUnescape(part)
		if err != nil {
			return Location{}, err
		}
		parts = append(parts, decoded)
	}
	return Location{
		Segments: parts,
		Query:    u.Query(),
	}, nil
}

func NewLocationAdapter() Adapter[Location] {
	return locationAdapter{}
}

type locationAdapter struct{}

func (l locationAdapter) Assert(a any) (*Location, bool) {
	switch v := a.(type) {
	case Location:
		return &v, true
	case *Location:
		return v, true
	default:
		return nil, false
	}
}

func (l locationAdapter) Decode(a any) (*Location, bool) {
	return l.Assert(a)
}

func (l locationAdapter) Encode(model *Location) (Location, error) {
	return *model, nil
}

func (l locationAdapter) EncodeAny(a any) (Location, error, bool) {
	loc, ok := l.Assert(a)
	if !ok {
		return Location{}, nil, false
	}
	return *loc, nil, true
}

var _ Adapter[Location] = locationAdapter{}
