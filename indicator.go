// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package doors

import "github.com/doors-dev/doors/internal/front"

// Indicator is a temporary DOM modification.
// It can change content, attributes, or classes, and is cleaned up automatically.
type Indicator = front.Indicator

type indicate = front.Indicate

// Selector targets DOM elements for indicators.
// Select by event source, CSS query, or closest matching parent.
type Selector interface {
	selector() front.Selector
}

type selector front.Selector

func (s selector) selector() front.Selector {
	return (front.Selector)(s)
}

// SelectorTarget selects the element that triggered the event.
func SelectorTarget() Selector {
	return (selector)(front.SelectTarget())
}

// SelectorQuery selects the first element matching a CSS query (e.g. "#id", ".class").
func SelectorQuery(query string) Selector {
	return (selector)(front.SelectQuery(query))
}

// SelectorQuery selects all elements matching a CSS query (e.g. "#id", ".class").
func SelectorQueryAll(query string) Selector {
	return (selector)(front.SelectQueryAll(query))
}

// SelectorParentQuery selects the closest ancestor matching a CSS query.
func SelectorParentQuery(query string) Selector {
	return (selector)(front.SelectParentQuery(query))
}

// IndicatorContent temporarily replaces innerHTML on the selected element.
type IndicatorContent struct {
	Selector Selector // Target element
	Content  string   // Replacement content
}

func (c IndicatorContent) Indicate() indicate {
	return front.IndicateContent(c.Selector.selector(), c.Content)
}

// IndicatorAttr temporarily sets an attribute on the selected element.
type IndicatorAttr struct {
	Selector Selector // Target element
	Name     string   // Attribute name
	Value    string   // Attribute value
}

func (c IndicatorAttr) Indicate() indicate {
	return front.IndicateAttr(c.Selector.selector(), c.Name, c.Value)
}

// IndicatorClass temporarily adds CSS classes to the selected element.
type IndicatorClass struct {
	Selector Selector // Target element
	Class    string   // Space-separated classes
}

func (c IndicatorClass) Indicate() indicate {
	return front.IndicateClass(c.Selector.selector(), c.Class)
}

// IndicatorClassRemove temporarily removes CSS classes from the selected element.
type IndicatorClassRemove struct {
	Selector Selector // Target element
	Class    string   // Space-separated classes
}

func (c IndicatorClassRemove) Indicate() indicate {
	return front.IndicateClassRemove(c.Selector.selector(), c.Class)
}

// *Only* helpers are convenience functions for simple indication.
// They target the event source element, a CSS query, or a closest matching parent.

// IndicatorOnlyContent sets content on the event target element.
func IndicatorOnlyContent(content string) []Indicator {
	return []Indicator{IndicatorContent{
		Selector: SelectorTarget(),
		Content:  content,
	}}
}

// IndicatorOnlyAttr sets an attribute on the event target element.
func IndicatorOnlyAttr(attr string, value string) []Indicator {
	return []Indicator{IndicatorAttr{
		Selector: SelectorTarget(),
		Name:     attr,
		Value:    value,
	}}
}

// IndicatorOnlyClassRemove removes classes from the event target element.
func IndicatorOnlyClassRemove(class string) []Indicator {
	return []Indicator{IndicatorClassRemove{
		Selector: SelectorTarget(),
		Class:    class,
	}}
}

// IndicatorOnlyClass adds classes to the event target element.
func IndicatorOnlyClass(class string) []Indicator {
	return []Indicator{IndicatorClass{
		Selector: SelectorTarget(),
		Class:    class,
	}}
}

// IndicatorOnlyContentQuery sets content on the first element matching a CSS query.
func IndicatorOnlyContentQuery(query string, content string) []Indicator {
	return []Indicator{IndicatorContent{
		Selector: SelectorQuery(query),
		Content:  content,
	}}
}

// IndicatorOnlyAttrQuery sets an attribute on the first element matching a CSS query.
func IndicatorOnlyAttrQuery(query string, attr string, value string) []Indicator {
	return []Indicator{IndicatorAttr{
		Selector: SelectorQuery(query),
		Name:     attr,
		Value:    value,
	}}
}

// IndicatorOnlyClassQuery adds classes to the first element matching a CSS query.
func IndicatorOnlyClassQuery(query string, class string) []Indicator {
	return []Indicator{IndicatorClass{
		Selector: SelectorQuery(query),
		Class:    class,
	}}
}

// IndicatorOnlyClassRemoveQuery removes classes from the first element matching a CSS query.
func IndicatorOnlyClassRemoveQuery(query string, class string) []Indicator {
	return []Indicator{IndicatorClassRemove{
		Selector: SelectorQuery(query),
		Class:    class,
	}}
}

// IndicatorOnlyContentQueryAll sets content on all elements matching a CSS query.
func IndicatorOnlyContentQueryAll(query string, content string) []Indicator {
	return []Indicator{IndicatorContent{
		Selector: SelectorQueryAll(query),
		Content:  content,
	}}
}

// IndicatorOnlyAttrQueryAll sets an attribute on all elements matching a CSS query.
func IndicatorOnlyAttrQueryAll(query string, attr string, value string) []Indicator {
	return []Indicator{IndicatorAttr{
		Selector: SelectorQueryAll(query),
		Name:     attr,
		Value:    value,
	}}
}

// IndicatorOnlyClassQueryAll adds classes to all elements matching a CSS query.
func IndicatorOnlyClassQueryAll(query string, class string) []Indicator {
	return []Indicator{IndicatorClass{
		Selector: SelectorQueryAll(query),
		Class:    class,
	}}
}

// IndicatorOnlyClassRemoveQueryAll removes classes from all elements matching a CSS query.
func IndicatorOnlyClassRemoveQueryAll(query string, class string) []Indicator {
	return []Indicator{IndicatorClassRemove{
		Selector: SelectorQueryAll(query),
		Class:    class,
	}}
}

// IndicatorOnlyContentQueryParent sets content on the closest ancestor matching a CSS query.
func IndicatorOnlyContentQueryParent(query string, content string) []Indicator {
	return []Indicator{IndicatorContent{
		Selector: SelectorParentQuery(query),
		Content:  content,
	}}
}

// IndicatorOnlyAttrQueryParent sets an attribute on the closest ancestor matching a CSS query.
func IndicatorOnlyAttrQueryParent(query string, attr string, value string) []Indicator {
	return []Indicator{IndicatorAttr{
		Selector: SelectorParentQuery(query),
		Name:     attr,
		Value:    value,
	}}
}

// IndicatorOnlyClassQueryParent adds classes to the closest ancestor matching a CSS query.
func IndicatorOnlyClassQueryParent(query string, class string) []Indicator {
	return []Indicator{IndicatorClass{
		Selector: SelectorParentQuery(query),
		Class:    class,
	}}
}

// IndicatorOnlyClassRemoveQueryParent removes classes from the closest ancestor matching a CSS query.
func IndicatorOnlyClassRemoveQueryParent(query string, class string) []Indicator {
	return []Indicator{IndicatorClassRemove{
		Selector: SelectorParentQuery(query),
		Class:    class,
	}}
}
