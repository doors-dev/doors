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

package doors

import (
	"github.com/doors-dev/doors/internal/front"
)

type Indicator = front.Indicator

// Indicators is a temporary DOM change applied on the client.
//
// Indicators are commonly used for pending UI such as "Loading..." text,
// temporary classes, or transient attributes while a request is in flight.
type Indicators interface {
	Indicators() []Indicator
	Joiner[Indicators]
}

func indicatorsOrNil(indicators Indicators) []Indicator {
	if indicators == nil {
		return nil
	}
	return indicators.Indicators()
}

type joinedIndicators []Indicators

func (is joinedIndicators) And(i Indicators) Indicators {
	c := make(joinedIndicators, len(is), len(is)+1)
	copy(c, is)
	c = append(c, i)
	return c
}

func (is joinedIndicators) Indicators() []Indicator {
	output := make([]Indicator, 0)
	for _, ind := range is {
		if ind == nil {
			continue
		}
		output = append(output, ind.Indicators()...)
	}
	return output
}

var _ Indicators = joinedIndicators(nil)

type IndicatorContent struct {
	Selector Selector // Target element
	Content  string   // Replacement content. WARNING: escaping is not applied.
}

// IndicateContent builds an [IndicatorContent] that targets the event element.
// WARNING: escaping is not applied.
func IndicateContent(content string) IndicatorContent {
	return IndicatorContent{Selector: SelectorTarget(), Content: content}
}

// IndicateContentQuery builds an [IndicatorContent] that targets the first
// element matching query.
func IndicateContentQuery(query, content string) IndicatorContent {
	return IndicatorContent{Selector: SelectorQuery(query), Content: content}
}

// IndicateContentQueryAll builds an [IndicatorContent] that targets every
// element matching query.
func IndicateContentQueryAll(query, content string) IndicatorContent {
	return IndicatorContent{Selector: SelectorQueryAll(query), Content: content}
}

// IndicateContentQueryParent builds an [IndicatorContent] that targets the
// closest ancestor matching query.
func IndicateContentQueryParent(query, content string) IndicatorContent {
	return IndicatorContent{Selector: SelectorQueryParent(query), Content: content}
}

func (ic IndicatorContent) And(i Indicators) Indicators {
	return joinedIndicators([]Indicators{ic, i})
}

func (ic IndicatorContent) Indicators() []Indicator {
	return []Indicator{front.IndicatorContent(ic.Selector, ic.Content)}
}

var _ Indicators = IndicatorContent{}

// IndicatorAttr temporarily sets an attribute on the selected element.
type IndicatorAttr struct {
	Selector Selector // Target element
	Name     string   // Attribute name
	Value    string   // Attribute value
}

// IndicateAttr builds an [IndicatorAttr] that targets the event element.
func IndicateAttr(name, value string) IndicatorAttr {
	return IndicatorAttr{Selector: SelectorTarget(), Name: name, Value: value}
}

// IndicateAttrQuery builds an [IndicatorAttr] that targets the first
// element matching query.
func IndicateAttrQuery(query, name, value string) IndicatorAttr {
	return IndicatorAttr{Selector: SelectorQuery(query), Name: name, Value: value}
}

// IndicateAttrQueryAll builds an [IndicatorAttr] that targets every
// element matching query.
func IndicateAttrQueryAll(query, name, value string) IndicatorAttr {
	return IndicatorAttr{Selector: SelectorQueryAll(query), Name: name, Value: value}
}

// IndicateAttrQueryParent builds an [IndicatorAttr] that targets the
// closest ancestor matching query.
func IndicateAttrQueryParent(query, name, value string) IndicatorAttr {
	return IndicatorAttr{Selector: SelectorQueryParent(query), Name: name, Value: value}
}

func (ia IndicatorAttr) And(i Indicators) Indicators {
	return joinedIndicators([]Indicators{ia, i})
}

func (ia IndicatorAttr) Indicators() []Indicator {
	return []Indicator{front.IndicatorAttr(ia.Selector, ia.Name, ia.Value)}
}

var _ Indicators = IndicatorAttr{}

// IndicatorClass temporarily adds CSS classes to the selected element.
type IndicatorClass struct {
	Selector Selector // Target element
	Class    string   // Space-separated classes
}

// IndicateClass builds an [IndicatorClass] that targets the event element.
func IndicateClass(class string) IndicatorClass {
	return IndicatorClass{Selector: SelectorTarget(), Class: class}
}

// IndicateClassQuery builds an [IndicatorClass] that targets the first
// element matching query.
func IndicateClassQuery(query, class string) IndicatorClass {
	return IndicatorClass{Selector: SelectorQuery(query), Class: class}
}

// IndicateClassQueryAll builds an [IndicatorClass] that targets every
// element matching query.
func IndicateClassQueryAll(query, class string) IndicatorClass {
	return IndicatorClass{Selector: SelectorQueryAll(query), Class: class}
}

// IndicateClassQueryParent builds an [IndicatorClass] that targets the
// closest ancestor matching query.
func IndicateClassQueryParent(query, class string) IndicatorClass {
	return IndicatorClass{Selector: SelectorQueryParent(query), Class: class}
}

func (ic IndicatorClass) And(i Indicators) Indicators {
	return joinedIndicators([]Indicators{ic, i})
}

func (ic IndicatorClass) Indicators() []Indicator {
	return []Indicator{front.IndicatorClass(ic.Selector, ic.Class)}
}

var _ Indicators = IndicatorClass{}

// IndicatorClassRemove temporarily removes CSS classes from the selected element.
type IndicatorClassRemove struct {
	Selector Selector // Target element
	Class    string   // Space-separated classes
}

// IndicateClassRemove builds an [IndicatorClassRemove] that targets the event element.
func IndicateClassRemove(class string) IndicatorClassRemove {
	return IndicatorClassRemove{Selector: SelectorTarget(), Class: class}
}

// IndicateClassRemoveQuery builds an [IndicatorClassRemove] that targets the
// first element matching query.
func IndicateClassRemoveQuery(query, class string) IndicatorClassRemove {
	return IndicatorClassRemove{Selector: SelectorQuery(query), Class: class}
}

// IndicateClassRemoveQueryAll builds an [IndicatorClassRemove] that targets
// every element matching query.
func IndicateClassRemoveQueryAll(query, class string) IndicatorClassRemove {
	return IndicatorClassRemove{Selector: SelectorQueryAll(query), Class: class}
}

// IndicateClassRemoveQueryParent builds an [IndicatorClassRemove] that
// targets the closest ancestor matching query.
func IndicateClassRemoveQueryParent(query, class string) IndicatorClassRemove {
	return IndicatorClassRemove{Selector: SelectorQueryParent(query), Class: class}
}

func (irc IndicatorClassRemove) And(i Indicators) Indicators {
	return joinedIndicators([]Indicators{irc, i})
}

func (irc IndicatorClassRemove) Indicators() []Indicator {
	return []Indicator{front.IndicatorClassRemove(irc.Selector, irc.Class)}
}

var _ Indicators = IndicatorClassRemove{}
