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
	"encoding/json"
	"maps"
	"net/url"
	"slices"
	"strings"
)

// Location is a URL path plus query string, held in decoded form.
type Location struct {
	// Segments are the path segments, decoded and without the separating
	// slashes. Empty means the root path.
	Segments []string `json:"segments"`
	// Query holds the query parameters in decoded form. Empty means no query
	// string.
	Query url.Values `json:"query"`
}

// MarshalJSON encodes nil Segments and Query as empty containers instead of
// null.
func (l Location) MarshalJSON() ([]byte, error) {
	type location Location
	if l.Segments == nil {
		l.Segments = []string{}
	}
	if l.Query == nil {
		l.Query = url.Values{}
	}
	return json.Marshal(location(l))
}

// Clone returns a deep copy of l with independent Segments and Query.
func (l Location) Clone() Location {
	return Location{
		Segments: slices.Clone(l.Segments),
		Query:    cloneQuery(l.Query),
	}
}

func cloneQuery(query url.Values) url.Values {
	clone := make(url.Values, len(query))
	for key, value := range query {
		clone[key] = slices.Clone(value)
	}
	return clone
}

// Path returns the escaped path with a leading slash and without the query
// string.
func (l Location) Path() string {
	b := strings.Builder{}
	b.WriteByte('/')
	for i, part := range l.Segments {
		if i != 0 {
			b.WriteByte('/')
		}
		b.WriteString(url.PathEscape(part))
	}
	return b.String()
}

// String returns the escaped path with a leading slash, followed by the encoded
// query string when Query is not empty.
func (l Location) String() string {
	b := strings.Builder{}
	b.WriteByte('/')
	for i, part := range l.Segments {
		if i != 0 {
			b.WriteByte('/')
		}
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

// NewLocationFromEscapedURI parses s into a [Location]. If s is not a valid
// URL reference, it returns the zero Location.
func NewLocationFromEscapedURI(s string) Location {
	u, err := url.Parse(s)
	if err != nil {
		return Location{}
	}
	return NewLocationFromURL(u)
}

// NewLocationFromURL decodes u into a [Location].
func NewLocationFromURL(u *url.URL) Location {
	parts := make([]string, 0)
	trimmed := strings.Trim(u.EscapedPath(), "/")
	if trimmed == "" {
		return Location{
			Segments: parts,
			Query:    u.Query(),
		}
	}
	for part := range strings.SplitSeq(trimmed, "/") {
		decoded, err := url.PathUnescape(part)
		if err != nil {
			parts = append(parts, part)
			continue
		}
		parts = append(parts, decoded)
	}
	return Location{
		Segments: parts,
		Query:    u.Query(),
	}
}
