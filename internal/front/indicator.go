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

type indicatorKind string

const (
	indicatorAttr        indicatorKind = "attr"
	indicatorClass       indicatorKind = "class"
	indicatorRemoveClass indicatorKind = "remove_class"
	indicatorContent     indicatorKind = "content"
)

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

type Indicator struct {
	selector Selector
	kind     indicatorKind
	param1   string
	param2   string
}

func IndicatorAttr(s Selector, name string, value string) Indicator {
	return Indicator{
		selector: s,
		kind:     indicatorAttr,
		param1:   name,
		param2:   value,
	}
}

func IndicatorClass(s Selector, class string) Indicator {
	return Indicator{
		selector: s,
		kind:     indicatorClass,
		param1:   class,
	}
}

func IndicatorClassRemove(s Selector, class string) Indicator {
	return Indicator{
		selector: s,
		kind:     indicatorRemoveClass,
		param1:   class,
	}
}

func IndicatorContent(s Selector, content string) Indicator {
	return Indicator{
		selector: s,
		kind:     indicatorContent,
		param1:   content,
	}
}

func (i Indicator) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{i.selector, i.kind, i.param1, i.param2})
}
