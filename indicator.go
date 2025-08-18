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

// ContentIndicator temporarily replaces the innerHTML of the selected element.
// The original content is automatically restored when the indicator is removed.
type ContentIndicator struct {
	Selector Selector // Element selector
	Content  string   // Text content to display
}

func (c ContentIndicator) Indicate() *indicate {
	return front.IndicateContent(c.Selector.selector(), c.Content)
}

// AttrIndicator temporarily sets an attribute on the selected element.
// The original attribute value (if any) is automatically restored when the indicator is removed.
type AttrIndicator struct {
	Selector Selector // Element selector
	Name     string   // Attribute name
	Value    string   // Attribute value
}

func (c AttrIndicator) Indicate() *indicate {
	return front.IndicateAttr(c.Selector.selector(), c.Name, c.Value)
}

// ClassIndicator temporarily adds CSS classes to the selected element.
// Multiple classes can be specified separated by spaces.
// The classes are automatically removed when the indicator is removed.
type ClassIndicator struct {
	Selector Selector // Element selector
	Class    string   // CSS classes to add (space-separated)
}

func (c ClassIndicator) Indicate() *indicate {
	return front.IndicateClass(c.Selector.selector(), c.Class)
}

// ClassRemoveIndicator temporarily removes CSS classes from the selected element.
// Multiple classes can be specified separated by spaces.
// The classes are automatically restored when the indicator is removed.
type ClassRemoveIndicator struct {
	Selector Selector // Element selector
	Class    string   // CSS classes to remove (space-separated)
}

func (c ClassRemoveIndicator) Indicate() *indicate {
	return front.IndicateClassRemove(c.Selector.selector(), c.Class)
}

// IndicatorContent creates an indicator that changes the content of the target element.
// This is a convenience function equivalent to ContentIndicator{SelectorTarget(), content}.
func IndicatorContent(content string) []Indicator {
	return []Indicator{ContentIndicator{
		Selector: SelectorTarget(),
		Content:  content,
	}}
}

// IndicatorAttr creates an indicator that sets an attribute on the target element.
// This is a convenience function equivalent to AttrIndicator{SelectorTarget(), attr, value}.
func IndicatorAttr(attr string, value string) []Indicator {
	return []Indicator{AttrIndicator{
		Selector: SelectorTarget(),
		Name:     attr,
		Value:    value,
	}}
}

// IndicatorClassRemove creates an indicator that removes CSS classes from the target element.
// Multiple classes can be specified separated by spaces.
// This is a convenience function equivalent to ClassRemoveIndicator{SelectorTarget(), class}.
func IndicatorClassRemove(class string) []Indicator {
	return []Indicator{ClassRemoveIndicator{
		Selector: SelectorTarget(),
		Class:    class,
	}}
}

// IndicatorClass creates an indicator that adds CSS classes to the target element.
// Multiple classes can be specified separated by spaces.
// This is a convenience function equivalent to ClassIndicator{SelectorTarget(), class}.
func IndicatorClass(class string) []Indicator {
	return []Indicator{ClassIndicator{
		Selector: SelectorTarget(),
		Class:    class,
	}}
}

// IndicatorContentQuery creates an indicator that changes the content of an element selected by CSS query.
// The query parameter should be a valid CSS selector string.
func IndicatorContentQuery(query string, content string) []Indicator {
	return []Indicator{ContentIndicator{
		Selector: SelectorQuery(query),
		Content:  content,
	}}
}

// IndicatorAttrQuery creates an indicator that sets an attribute on an element selected by CSS query.
// The query parameter should be a valid CSS selector string.
func IndicatorAttrQuery(query string, attr string, value string) []Indicator {
	return []Indicator{AttrIndicator{
		Selector: SelectorQuery(query),
		Name:     attr,
		Value:    value,
	}}
}

// IndicatorClassQuery creates an indicator that adds CSS classes to an element selected by CSS query.
// The query parameter should be a valid CSS selector string.
// Multiple classes can be specified separated by spaces.
func IndicatorClassQuery(query string, class string) []Indicator {
	return []Indicator{ClassIndicator{
		Selector: SelectorQuery(query),
		Class:    class,
	}}
}

// IndicatorClassRemoveQuery creates an indicator that removes CSS classes from an element selected by CSS query.
// The query parameter should be a valid CSS selector string.
// Multiple classes can be specified separated by spaces.
func IndicatorClassRemoveQuery(query string, class string) []Indicator {
	return []Indicator{ClassRemoveIndicator{
		Selector: SelectorQuery(query),
		Class:    class,
	}}
}

// IndicatorContentQueryParent creates an indicator that changes the content of a parent element.
// Starting from the target element's parent, it finds the closest ancestor matching the CSS query.
func IndicatorContentQueryParent(query string, content string) []Indicator {
	return []Indicator{ContentIndicator{
		Selector: SelectorParentQuery(query),
		Content:  content,
	}}
}

// IndicatorAttrQueryParent creates an indicator that sets an attribute on a parent element.
// Starting from the target element's parent, it finds the closest ancestor matching the CSS query.
func IndicatorAttrQueryParent(query string, attr string, value string) []Indicator {
	return []Indicator{AttrIndicator{
		Selector: SelectorParentQuery(query),
		Name:     attr,
		Value:    value,
	}}
}

// IndicatorClassQueryParent creates an indicator that adds CSS classes to a parent element.
// Starting from the target element's parent, it finds the closest ancestor matching the CSS query.
// Multiple classes can be specified separated by spaces.
func IndicatorClassQueryParent(query string, class string) []Indicator {
	return []Indicator{ClassIndicator{
		Selector: SelectorParentQuery(query),
		Class:    class,
	}}
}

// IndicatorClassRemoveQueryParent creates an indicator that removes CSS classes from a parent element.
// Starting from the target element's parent, it finds the closest ancestor matching the CSS query.
// Multiple classes can be specified separated by spaces.
func IndicatorClassRemoveQueryParent(query string, class string) []Indicator {
	return []Indicator{ClassRemoveIndicator{
		Selector: SelectorParentQuery(query),
		Class:    class,
	}}
}
