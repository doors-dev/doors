package doors

import "github.com/doors-dev/doors/internal/front"

type Indicator = front.Indicator

type indicate = front.Indicate

type Selector interface {
	selector() *front.Selector
}

type selector front.Selector

func (s *selector) selector() *front.Selector {
	return (*front.Selector)(s)
}

func SelectorTarget() Selector {
	return (*selector)(front.SelectTarget())
}

func SelectorQuery(query string) Selector {
	return (*selector)(front.SelectQuery(query))
}
func SelectorParentQuery(query string) Selector {
	return (*selector)(front.SelectParentQuery(query))
}

type ContentIndicator struct {
	Selector Selector
	Content  string
}

func (c ContentIndicator) Indicate() *indicate {
	return front.IndicateContent(c.Selector.selector(), c.Content)
}

type AttrIndicator struct {
	Selector Selector
	Name     string
	Value    string
}

func (c AttrIndicator) Indicate() *indicate {
	return front.IndicateAttr(c.Selector.selector(), c.Name, c.Value)
}

type ClassIndicator struct {
	Selector Selector
	Class    string
}

func (c ClassIndicator) Indicate() *indicate {
	return front.IndicateClass(c.Selector.selector(), c.Class)
}

type ClassRemoveIndicator struct {
	Selector Selector
	Class    string
}

func (c ClassRemoveIndicator) Indicate() *indicate {
	return front.IndicateClassRemove(c.Selector.selector(), c.Class)
}

func IndicatorContent(content string) []Indicator {
	return []Indicator{ContentIndicator{
		Selector: SelectorTarget(),
		Content:  content,
	}}
}

func IndicatorAttr(attr string, value string) []Indicator {
	return []Indicator{AttrIndicator{
		Selector: SelectorTarget(),
		Name:     attr,
		Value:    value,
	}}
}

func IndicatorClassRemove(class string) []Indicator {
	return []Indicator{ClassRemoveIndicator{
		Selector: SelectorTarget(),
		Class:    class,
	}}
}

func IndicatorClass(class string) []Indicator {
	return []Indicator{ClassIndicator{
		Selector: SelectorTarget(),
		Class:    class,
	}}
}

func IndicatorContentQuery(query string, content string) []Indicator {
	return []Indicator{ContentIndicator{
		Selector: SelectorQuery(query),
		Content:  content,
	}}
}

func IndicatorAttrQuery(query string, attr string, value string) []Indicator {
	return []Indicator{AttrIndicator{
		Selector: SelectorQuery(query),
		Name:     attr,
		Value:    value,
	}}
}

func IndicatorClassQuery(query string, class string) []Indicator {
	return []Indicator{ClassIndicator{
		Selector: SelectorQuery(query),
		Class:    class,
	}}
}

func IndicatorClassRemoveQuery(query string, class string) []Indicator {
	return []Indicator{ClassRemoveIndicator{
		Selector: SelectorQuery(query),
		Class:    class,
	}}
}

func IndicatorContentQueryParent(query string, content string) []Indicator {
	return []Indicator{ContentIndicator{
		Selector: SelectorParentQuery(query),
		Content:  content,
	}}
}

func IndicatorAttrQueryParent(query string, attr string, value string) []Indicator {
	return []Indicator{AttrIndicator{
		Selector: SelectorParentQuery(query),
		Name:     attr,
		Value:    value,
	}}
}

func IndicatorClassQueryParent(query string, class string) []Indicator {
	return []Indicator{ClassIndicator{
		Selector: SelectorParentQuery(query),
		Class:    class,
	}}
}

func IndicatorClassRemoveQueryParent(query string, class string) []Indicator {
	return []Indicator{ClassRemoveIndicator{
		Selector: SelectorParentQuery(query),
		Class:    class,
	}}
}
