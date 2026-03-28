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

// Location is a parsed or generated URL path plus query string.
type Location struct {
	// Query holds the decoded query parameters.
	Query    url.Values
	// Segments holds the decoded path segments without leading or trailing
	// slashes.
	Segments []string
}

// Path returns the escaped path portion of l without the query string.
func (l Location) Path() string {
	b := strings.Builder{}
	for _, part := range l.Segments {
		b.WriteByte('/')
		b.WriteString(url.PathEscape(part))
	}
	return b.String()
}

// String returns l encoded as `/<segments>?<query>`.
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

// EqualLocation reports whether a and b encode the same path and query.
func EqualLocation(a, b Location) bool {
	return slices.Equal(a.Segments, b.Segments) && maps.EqualFunc(a.Query, b.Query, slices.Equal)
}

// NewLocationFromEscapedURI parses s into a [Location].
func NewLocationFromEscapedURI(s string) (Location, error) {
	u, err := url.Parse(s)
	if err != nil {
		return Location{}, err
	}
	return NewLocationFromURL(u)
}

// NewLocationFromURL decodes u into a [Location].
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

// NewLocationAdapter returns an adapter that treats [Location] as a path
// model.
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
