package doors

import "github.com/doors-dev/doors/internal/front"

// Selector picks the elements an [Indicator] applies to.
//
// Query selectors search the whole document. Target and parent selectors need
// an event element and select nothing without one.
type Selector = front.Selector

// SelectorTarget selects the event element.
func SelectorTarget() Selector {
	return front.SelectTarget()
}

// SelectorQuery selects the first element in the document matching query.
func SelectorQuery(query string) Selector {
	return front.SelectQuery(query)
}

// SelectorQueryAll selects every element in the document matching query.
func SelectorQueryAll(query string) Selector {
	return front.SelectQueryAll(query)
}

// SelectorQueryParent selects the closest ancestor of the event element
// matching query.
func SelectorQueryParent(query string) Selector {
	return front.SelectQueryParent(query)
}
