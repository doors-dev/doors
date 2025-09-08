// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package doors

import "github.com/doors-dev/doors/internal/front"

// Indicator represents a temporary modification to a DOM element.
// Indicators can change content, attributes, or CSS classes and are automatically
// cleaned up when no longer needed.
type Indicator = front.Indicator

type indicate = front.Indicate

// Selector defines how to target DOM elements for indication.
// Elements can be selected by targeting the event source, CSS queries, or parent traversal.
type Selector interface {
	selector() *front.Selector
}

type selector front.Selector

func (s *selector) selector() *front.Selector {
	return (*front.Selector)(s)
}

// SelectorTarget creates a selector that targets the element that triggered the event.
// This is the most common selector type for interactive elements like buttons.
func SelectorTarget() Selector {
	return (*selector)(front.SelectTarget())
}

// SelectorQuery creates a selector that targets the first element matching the CSS selector.
// The query parameter should be a valid CSS selector string (e.g., "#myId", ".myClass", "div.container").
func SelectorQuery(query string) Selector {
	return (*selector)(front.SelectQuery(query))
}

// SelectorParentQuery creates a selector that finds the closest ancestor element matching the CSS sele 2 ctor.
// Starting from the target element's parent, it traverses up the DOM tree to find the first matching ancestor.
// The query parameter should be a valid CSS selector string.
func SelectorParentQuery(query string) Selector {
	return (*selector)(front.SelectParentQuery(query))
}

// IndicatorContent temporarily replaces the innerHTML of the selected element.
// The original content is automatically restored when the indicator is removed.
type IndicatorContent struct {
	Selector Selector // Element selector
	Content  string   // Text content to display
}

func (c IndicatorContent) Indicate() *indicate {
	return front.IndicateContent(c.Selector.selector(), c.Content)
}

// IndicatorAttr temporarily sets an attribute on the selected element.
// The original attribute value (if any) is automatically restored when the indicator is removed.
type IndicatorAttr struct {
	Selector Selector // Element selector
	Name     string   // Attribute name
	Value    string   // Attribute value
}

func (c IndicatorAttr) Indicate() *indicate {
	return front.IndicateAttr(c.Selector.selector(), c.Name, c.Value)
}

// IndicatorClass temporarily adds CSS classes to the selected element.
// Multiple classes can be specified separated by spaces.
// The classes are automatically removed when the indicator is removed.
type IndicatorClass struct {
	Selector Selector // Element selector
	Class    string   // CSS classes to add (space-separated)
}

func (c IndicatorClass) Indicate() *indicate {
	return front.IndicateClass(c.Selector.selector(), c.Class)
}

// IndicatorClassRemove temporarily removes CSS classes from the selected element.
// Multiple classes can be specified separated by spaces.
// The classes are automatically restored when the indicator is removed.
type IndicatorClassRemove struct {
	Selector Selector // Element selector
	Class    string   // CSS classes to remove (space-separated)
}

func (c IndicatorClassRemove) Indicate() *indicate {
	return front.IndicateClassRemove(c.Selector.selector(), c.Class)
}

// IndicatorOnlyContent creates an indicator that changes the content of the target element.
// This is a convenience function equivalent to ContentIndicator{SelectorTarget(), content}.
func IndicatorOnlyContent(content string) []Indicator {
	return []Indicator{IndicatorContent{
		Selector: SelectorTarget(),
		Content:  content,
	}}
}

// IndicatorOnlyAttr creates an indicator that sets an attribute on the target element.
// This is a convenience function equivalent to AttrIndicator{SelectorTarget(), attr, value}.
func IndicatorOnlyAttr(attr string, value string) []Indicator {
	return []Indicator{IndicatorAttr{
		Selector: SelectorTarget(),
		Name:     attr,
		Value:    value,
	}}
}

// IndicatorOnlyClassRemove creates an indicator that removes CSS classes from the target element.
// Multiple classes can be specified separated by spaces.
// This is a convenience function equivalent to ClassRemoveIndicator{SelectorTarget(), class}.
func IndicatorOnlyClassRemove(class string) []Indicator {
	return []Indicator{IndicatorClassRemove{
		Selector: SelectorTarget(),
		Class:    class,
	}}
}

// IndicatorOnlyClass creates an indicator that adds CSS classes to the target element.
// Multiple classes can be specified separated by spaces.
// This is a convenience function equivalent to ClassIndicator{SelectorTarget(), class}.
func IndicatorOnlyClass(class string) []Indicator {
	return []Indicator{IndicatorClass{
		Selector: SelectorTarget(),
		Class:    class,
	}}
}

// IndicatorOnlyContentQuery creates an indicator that changes the content of an element selected by CSS query.
// The query parameter should be a valid CSS selector string.
func IndicatorOnlyContentQuery(query string, content string) []Indicator {
	return []Indicator{IndicatorContent{
		Selector: SelectorQuery(query),
		Content:  content,
	}}
}

// IndicatorOnlyAttrQuery creates an indicator that sets an attribute on an element selected by CSS query.
// The query parameter should be a valid CSS selector string.
func IndicatorOnlyAttrQuery(query string, attr string, value string) []Indicator {
	return []Indicator{IndicatorAttr{
		Selector: SelectorQuery(query),
		Name:     attr,
		Value:    value,
	}}
}

// IndicatorOnlyClassQuery creates an indicator that adds CSS classes to an element selected by CSS query.
// The query parameter should be a valid CSS selector string.
// Multiple classes can be specified separated by spaces.
func IndicatorOnlyClassQuery(query string, class string) []Indicator {
	return []Indicator{IndicatorClass{
		Selector: SelectorQuery(query),
		Class:    class,
	}}
}

// IndicatorOnlyClassRemoveQuery creates an indicator that removes CSS classes from an element selected by CSS query.
// The query parameter should be a valid CSS selector string.
// Multiple classes can be specified separated by spaces.
func IndicatorOnlyClassRemoveQuery(query string, class string) []Indicator {
	return []Indicator{IndicatorClassRemove{
		Selector: SelectorQuery(query),
		Class:    class,
	}}
}

// IndicatorOnlyContentQueryParent creates an indicator that changes the content of a parent element.
// Starting from the target element's parent, it finds the closest ancestor matching the CSS query.
func IndicatorOnlyContentQueryParent(query string, content string) []Indicator {
	return []Indicator{IndicatorContent{
		Selector: SelectorParentQuery(query),
		Content:  content,
	}}
}

// IndicatorOnlyAttrQueryParent creates an indicator that sets an attribute on a parent element.
// Starting from the target element's parent, it finds the closest ancestor matching the CSS query.
func IndicatorOnlyAttrQueryParent(query string, attr string, value string) []Indicator {
	return []Indicator{IndicatorAttr{
		Selector: SelectorParentQuery(query),
		Name:     attr,
		Value:    value,
	}}
}

// IndicatorOnlyClassQueryParent creates an indicator that adds CSS classes to a parent element.
// Starting from the target element's parent, it finds the closest ancestor matching the CSS query.
// Multiple classes can be specified separated by spaces.
func IndicatorOnlyClassQueryParent(query string, class string) []Indicator {
	return []Indicator{IndicatorClass{
		Selector: SelectorParentQuery(query),
		Class:    class,
	}}
}

// IndicatorOnlyClassRemoveQueryParent creates an indicator that removes CSS classes from a parent element.
// Starting from the target element's parent, it finds the closest ancestor matching the CSS query.
// Multiple classes can be specified separated by spaces.
func IndicatorOnlyClassRemoveQueryParent(query string, class string) []Indicator {
	return []Indicator{IndicatorClassRemove{
		Selector: SelectorParentQuery(query),
		Class:    class,
	}}
}
