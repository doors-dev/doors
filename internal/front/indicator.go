// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package front

import "encoding/json"

type SelectorMode string

const (
	SelectModeTarget      SelectorMode = "target"
	SelectModeQuery       SelectorMode = "query"
	SelectModeParentQuery SelectorMode = "parent_query"
)

type Selector struct {
	mode  SelectorMode
	query string
}

func IntoIndicate(indicator []Indicator) []*Indicate {
	a := make([]*Indicate, len(indicator))
	for i, s := range indicator {
		a[i] = s.Indicate()
	}
	return a
}

func SelectTarget() *Selector {
	return &Selector{
		mode: SelectModeTarget,
	}
}
func SelectQuery(query string) *Selector {
	return &Selector{
		mode:  SelectModeQuery,
		query: query,
	}
}
func SelectParentQuery(query string) *Selector {
	return &Selector{
		mode:  SelectModeParentQuery,
		query: query,
	}
}

func (s *Selector) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{string(s.mode), s.query})
}

type Indicator interface {
	Indicate() *Indicate
}

type Indicate struct {
	selector *Selector
	kind     string
	param1   string
	param2   string
}

func IndicateAttr(s *Selector, name string, value string) *Indicate {
	return &Indicate{
		selector: s,
		kind:     "attr",
		param1:   name,
		param2:   value,
	}
}

func IndicateClass(s *Selector, class string) *Indicate {
	return &Indicate{
		selector: s,
		kind:     "class",
		param1:   class,
	}
}

func IndicateClassRemove(s *Selector, class string) *Indicate {
	return &Indicate{
		selector: s,
		kind:     "remove_class",
		param1:   class,
	}
}

func IndicateContent(s *Selector, content string) *Indicate {
	return &Indicate{
		selector: s,
		kind:     "content",
		param1:   content,
	}
}

func (i *Indicate) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{i.selector, i.kind, i.param1, i.param2})
}
