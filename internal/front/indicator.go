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

package front

import "encoding/json"

type SelectorMode string

const (
	SelectModeTarget      SelectorMode = "target"
	SelectModeQuery       SelectorMode = "query"
	SelectModeQueryAll    SelectorMode = "query_all"
	SelectModeParentQuery SelectorMode = "parent_query"
)

type Selector struct {
	mode  SelectorMode
	query string
}

func IntoIndicate(indicator []Indicator) []Indicate {
	a := make([]Indicate, len(indicator))
	for i, s := range indicator {
		a[i] = s.Indicate()
	}
	return a
}

func SelectTarget() Selector {
	return Selector{
		mode: SelectModeTarget,
	}
}
func SelectQuery(query string) Selector {
	return Selector{
		mode:  SelectModeQuery,
		query: query,
	}
}
func SelectQueryAll(query string) Selector {
	return Selector{
		mode:  SelectModeQueryAll,
		query: query,
	}
}
func SelectQueryParent(query string) Selector {
	return Selector{
		mode:  SelectModeParentQuery,
		query: query,
	}
}

func (s Selector) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{string(s.mode), s.query})
}

type Indicator interface {
	Indicate() Indicate
}

type Indicate struct {
	selector Selector
	kind     string
	param1   string
	param2   string
}

func IndicateAttr(s Selector, name string, value string) Indicate {
	return Indicate{
		selector: s,
		kind:     "attr",
		param1:   name,
		param2:   value,
	}
}

func IndicateClass(s Selector, class string) Indicate {
	return Indicate{
		selector: s,
		kind:     "class",
		param1:   class,
	}
}

func IndicateClassRemove(s Selector, class string) Indicate {
	return Indicate{
		selector: s,
		kind:     "remove_class",
		param1:   class,
	}
}

func IndicateContent(s Selector, content string) Indicate {
	return Indicate{
		selector: s,
		kind:     "content",
		param1:   content,
	}
}

func (i Indicate) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{i.selector, i.kind, i.param1, i.param2})
}
